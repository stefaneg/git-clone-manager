package gitlab

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
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
	Groups               []GitLabGroupConfig                `yaml:"groups"`
	Projects             []gitremote.GitRemoteProjectConfig `yaml:"projects"`
}

type GitLabGroupConfig struct {
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
	Group             GitlabApiGroup
	GroupConfig       GitLabGroupConfig
	GitLabConfig      GitLabConfig
}

func (p ProjectMetadata) CloneArchived() bool {
	return p.GroupConfig.CloneArchived
}

func (p ProjectMetadata) CloneRootDirectory() string {
	return p.GitLabConfig.CloneDirectory
}

func (gitlab GitLabConfig) GetCloneDirectory() string {
	return gitlab.CloneDirectory
}

func (gitlab GitLabConfig) getBaseUrl() string {
	return fmt.Sprintf("https://%s/api/v4", gitlab.HostName)
}

func (gitlab GitLabConfig) fetchProjectsForGroup(token string, group GitlabApiGroup, projectChannel chan ProjectMetadata) {

	var allProjects []ProjectMetadata

	projects, err := gitlab.fetchProjects(token, group)
	if err != nil {
		l.Log.Printf("Failed to fetch projects for group %s: %v", group.Name, err)
	}
	allProjects = append(allProjects, projects...)
	for _, project := range projects {
		project.Group = group
		project.GitLabConfig = gitlab
		projectChannel <- project
	}
}

func (gitlab GitLabConfig) OpenGroupsChannel(token string, rootGroupConfig GitLabGroupConfig, subGroupsChannel chan<- GitlabApiGroup) {

	gwg := sync.WaitGroup{}
	groupChannel := make(chan GitlabApiGroup, 20)

	rootGroup, err := gitlab.fetchGroupInfo(token, rootGroupConfig.ID)
	if err != nil {
		l.Log.Errorf("failed to fetch rootGroupConfig info for rootGroupConfig %s: %w", rootGroupConfig.ID, err)
	}

	//l.Log.Debugf(color.FgGreen("ADD"))
	gwg.Add(1)
	go func() {
		// Start by adding root group to the work list
		subgroups, err := gitlab.fetchSubgroups(token, fmt.Sprintf("%d", rootGroup.ID))
		//l.Log.Debugf("Received %n subgroups ", len(subgroups))
		if err != nil {
			l.Log.Errorf(fmt.Sprintf("failed to fetch subgroups for group %s: %w", rootGroup.ID, err))
		}
		for _, subgroup := range subgroups {
			//l.Log.Debugf("Added %n to groupChannel from root %s", subgroup.ID, rootGroupConfig.ID)
			//l.Log.Debugf(color.FgGreen("ADD"))
			gwg.Add(1)
			groupChannel <- subgroup
		}
		//l.Log.Debugf("Finished fetching subgroups for ROOT GROUP %s", rootGroup.ID)
		//l.Log.Debugf(color.FgRed("DONE"))
		gwg.Done()
	}()

	go func() {
		for {
			receivedGroup, ok := <-groupChannel
			if !ok {
				//l.Log.Debugf("%s %s \n", color.FgRed("Group channel close FOR GROUP FETCH, breaking"), rootGroupConfig.ID)
				break
			}
			////l.Log.Debugf("%s %s \n", color.FgGreen("Processing............. "), receivedGroup.Name)
			subGroupsChannel <- receivedGroup

			subgroups, err := gitlab.fetchSubgroups(token, fmt.Sprintf("%d", receivedGroup.ID))
			if err != nil {
				l.Log.Errorf(fmt.Sprintf("failed to fetch subgroups for group %s: %w", receivedGroup.ID, err))
			}
			//l.Log.Debugf("Queueing subgroups ", len(subgroups))
			for _, subgroup := range subgroups {
				//l.Log.Debugf(color.FgGreen("ADD %n"), subgroup.ID)
				gwg.Add(1)
				go func() {
					groupChannel <- subgroup
				}()
			}
			//l.Log.Debugf(color.FgRed("DONE %n"), receivedGroup.ID)
			gwg.Done()
		}
		close(subGroupsChannel)
		//l.Log.Debugf("CLOSED subgroupsChannel ")
	}()

	gwg.Wait()
	//l.Log.Debugf("Closing group channel in GROUPS fetch %s", rootGroupConfig.ID)
	close(groupChannel)
	//l.Log.Debugf("All groups fetched...")
}

func (gitlab GitLabConfig) OpenGroupProjectChannel(token string, rootGroupConfig GitLabGroupConfig, gitlabProjectChannel chan ProjectMetadata) {

	pwg := sync.WaitGroup{}

	////l.Log.Debugf("Opening channel for receiving groups %s", color.FgMagenta(rootGroupConfig.ID))
	groupChannel := make(chan GitlabApiGroup, 10)

	////l.Log.Debugf(color.FgMagenta("ADD opening groups channel"))
	pwg.Add(1)
	go func() {
		gitlab.OpenGroupsChannel(token, rootGroupConfig, groupChannel)
		////l.Log.Debugf(color.FgCyan("DONE groups channel should be closed"))
		pwg.Done()
	}()
	// Process received groups
	go func() {
		for {
			receivedGroup, ok := <-groupChannel
			if !ok {
				//l.Log.Debugf("%s %s \n", color.FgRed("Group channel close for PROJECTS, breaking"), rootGroupConfig.ID)
				break
			}
			////l.Log.Debugf(color.FgCyan("RECEIVED group to fetch projects... %s", receivedGroup.Name))
			////l.Log.Debugf(color.FgMagenta("ADD %s"), receivedGroup.Name)
			pwg.Add(1)
			go func() {
				gitlab.fetchProjectsForGroup(token, receivedGroup, gitlabProjectChannel)
				////l.Log.Debugf(color.FgCyan("DONE %s"), receivedGroup.Name)
				pwg.Done()
			}()
		}
	}()

	////l.Log.Debugf("Waiting for all GROUP fetches to complete %s", color.FgGreen(rootGroupConfig.ID))
	pwg.Wait()
	////l.Log.Debugf("Closing channels for ROOT GROUP when fetching Group Projects %s", color.FgRed(rootGroupConfig.ID))

	close(gitlabProjectChannel)

	logrus.Debugf("All projects fetched for group ... %s", color.FgGreen(rootGroupConfig.ID))
}

func (gitlab GitLabConfig) fetchProjects(token string, group GitlabApiGroup) ([]ProjectMetadata, error) {
	//l.Log.Debugf("Fetching projects in group %s\n", color.FgCyan(group))
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
	//l.Log.Debugf("Fetching subgroups in group %s", color.FgCyan(groupID))
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
	//l.Log.Debugf("RETURNING Fetched subgroups in group %s", color.FgCyan(groupID))
	return subgroups, nil
}

func (gitlab GitLabConfig) fetchGroupInfo(token string, groupID string) (*GitlabApiGroup, error) {
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
	return &group, nil
}
