package gitlab

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"tools/internal/color"
	"tools/internal/gitremote"
	l "tools/internal/log"
)

type GitLabConfig struct {
	EnvTokenVariableName string                             `yaml:"tokenEnvVar"`    // The environment variable name for the GitLab token
	HostName             string                             `yaml:"hostName"`       // Gitlab host name
	CloneDirectory       string                             `yaml:"cloneDirectory"` // Where to clone projects in local directory structure
	Groups               []GitLabConfigGroup                `yaml:"groups"`
	Projects             []gitremote.GitRemoteProjectConfig `yaml:"projects"`
}

type GitLabConfigGroup struct {
	ID            string `yaml:"id"`
	CloneArchived bool   `yaml:"cloneArchived"`
}

type GitlabApiGroup struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ProjectMetadata struct {
	Name              string `json:"name"`
	SSHURLToRepo      string `json:"ssh_url_to_repo"`
	PathWithNamespace string `json:"path_with_namespace"`
	Archived          bool   `json:"archived"`
}

func (gitlab GitLabConfig) GetCloneDirectory() string {
	return gitlab.CloneDirectory
}

func (gitlab GitLabConfig) getBaseUrl() string {
	return fmt.Sprintf("https://%s/api/v4", gitlab.HostName)
}

func (gitlab GitLabConfig) fetchAllGroupsRecursively(token string, group *GitlabApiGroup, groupChannel chan GitlabApiGroup) {
	var allGroups []GitlabApiGroup

	// Add the current group to the list
	allGroups = append(allGroups, *group)

	// Fetch subgroups
	subgroups, err := gitlab.fetchSubgroups(token, fmt.Sprintf("%d", group.ID))
	if err != nil {
		l.Log.Errorf(fmt.Sprintf("failed to fetch subgroups for group %s: %w", group.ID, err))
	}
	// Recursively fetch subgroups for each of the subgroups
	for _, subgroup := range subgroups {
		groupChannel <- subgroup
		go gitlab.fetchAllGroupsRecursively(token, &subgroup, groupChannel)
	}
}

func (gitlab GitLabConfig) fetchProjectsForGroup(token string, group GitlabApiGroup, projectChannel chan ProjectMetadata) {
	var allProjects []ProjectMetadata

	projects, err := gitlab.fetchProjects(token, group)
	if err != nil {
		l.Log.Printf("Failed to fetch projects for group %s: %v", group.Name, err)
	}
	allProjects = append(allProjects, projects...)
	for _, project := range projects {
		projectChannel <- project
	}
}

func (gitlab GitLabConfig) GetGroupProjects(token string, group GitLabConfigGroup, repoChannel chan ProjectMetadata, fetchWaitGroup *sync.WaitGroup) error {

	l.Log.Debugf("Opening channel for %s", color.FgMagenta(group.ID))
	groupChannel := make(chan GitlabApiGroup, 10)

	go func() {
		for {
			receivedGroup, ok := <-groupChannel
			if !ok {
				l.Log.Debugf("%s %s \n", color.FgRed("Channel close, breaking"), group.ID)
				break
			}
			l.Log.Debugf(color.FgGreen("RECEIVIED group ", len(receivedGroup.Name)))
			go func() {
				fetchWaitGroup.Add(1)
				defer fetchWaitGroup.Done()
				gitlab.fetchAllGroupsRecursively(token, &receivedGroup, groupChannel)
			}()
			go func() {
				fetchWaitGroup.Add(1)
				defer fetchWaitGroup.Done()
				gitlab.fetchProjectsForGroup(token, receivedGroup, repoChannel)
			}()
		}
	}()

	rootGroup, err := gitlab.fetchGroupInfo(token, group.ID, groupChannel)
	if err != nil {
		return fmt.Errorf("failed to fetch group info for group %s: %w", group.ID, err)
	}
	go gitlab.fetchAllGroupsRecursively(token, rootGroup, groupChannel)

	return nil
}

func (gitlab GitLabConfig) fetchProjects(token string, group GitlabApiGroup) ([]ProjectMetadata, error) {
	l.Log.Debugf("Fetching projects in group %s\n", color.FgCyan(group))
	url := fmt.Sprintf("%s/groups/%d/projects", gitlab.getBaseUrl(), group.ID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("PRIVATE-TOKEN", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(body io.ReadCloser) {
		err := body.Close()
		if err != nil {
			l.Log.Errorf("Failed to close response body: %v", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitLab API request failed with status: %s", resp.Status)
	}

	var projects []ProjectMetadata
	if err := json.NewDecoder(resp.Body).Decode(&projects); err != nil {
		return nil, err
	}

	return projects, nil
}

func (gitlab GitLabConfig) fetchSubgroups(token, groupID string) ([]GitlabApiGroup, error) {
	l.Log.Debugf("Fetching subgroups in group %s", color.FgCyan(groupID))
	url := fmt.Sprintf("%s/groups/%s/subgroups", gitlab.getBaseUrl(), groupID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("PRIVATE-TOKEN", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			l.Log.Errorf("Failed to close response body: %v", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitLab API request failed with status: %s", resp.Status)
	}

	var subgroups []GitlabApiGroup
	if err := json.NewDecoder(resp.Body).Decode(&subgroups); err != nil {
		return nil, err
	}

	return subgroups, nil
}

func (gitlab GitLabConfig) fetchGroupInfo(token string, groupID string, channel chan GitlabApiGroup) (*GitlabApiGroup, error) {
	url := fmt.Sprintf("%s/groups/%s", gitlab.getBaseUrl(), groupID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("PRIVATE-TOKEN", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch group info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch group info: status code %d", resp.StatusCode)
	}

	var group GitlabApiGroup
	if err := json.NewDecoder(resp.Body).Decode(&group); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	channel <- group
	return &group, nil
}
