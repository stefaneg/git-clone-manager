package gitlab

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"tools/internal/color"
	"tools/internal/gitrepo"
	logger "tools/internal/log"
)

type GitLabConfig struct {
	EnvTokenVariableName string                `yaml:"tokenEnvVar"`    // The environment variable name for the GitLab token
	HostName             string                `yaml:"hostName"`       // Gitlab host name
	CloneDirectory       string                `yaml:"cloneDirectory"` // Where to clone projects in local directory structure
	Groups               []GitLabConfigGroup   `yaml:"groups"`
	Projects             []GitLabConfigProject `yaml:"projects"`
}

type GitLabConfigGroup struct {
	ID            string `yaml:"id"`
	CloneArchived bool   `yaml:"cloneArchived"`
}

type GitLabConfigProject struct {
	Name     string `yaml:"name"`
	FullPath string `yaml:"fullPath"`
}

type GitlabApiGroup struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (gitlab GitLabConfig) GetCloneDirectory() string {
	return gitlab.CloneDirectory
}

func (gitlab GitLabConfig) getBaseUrl() string {
	return fmt.Sprintf("https://%s/api/v4", gitlab.HostName)
}

func (gitlab GitLabConfig) FetchAllGroupsRecursively(token string, group *GitlabApiGroup) ([]GitlabApiGroup, error) {
	var allGroups []GitlabApiGroup

	// Add the current group to the list
	allGroups = append(allGroups, *group)

	// Fetch subgroups
	subgroups, err := gitlab.fetchSubgroups(token, fmt.Sprintf("%d", group.ID))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch subgroups for group %s: %w", group.ID, err)
	}

	// Recursively fetch subgroups for each of the subgroups
	for _, subgroup := range subgroups {
		subGroups, err := gitlab.FetchAllGroupsRecursively(token, &subgroup)
		if err != nil {
			logger.Log.Printf("Failed to fetch subgroups for group %s: %v", subgroup.Name, err)
		} else {
			allGroups = append(allGroups, subGroups...)
		}
	}

	return allGroups, nil
}

func (gitlab GitLabConfig) fetchProjectsForGroups(token string, groups []GitlabApiGroup) ([]gitrepo.GitRepoSpec, error) {
	var allProjects []gitrepo.GitRepoSpec

	for _, group := range groups {
		projects, err := gitlab.fetchProjects(token, fmt.Sprintf("%d", group.ID))
		if err != nil {
			return nil, fmt.Errorf("failed to fetch projects for group %d: %w", group.ID, err)
		}
		allProjects = append(allProjects, projects...)
	}

	return allProjects, nil
}

func (gitlab GitLabConfig) GetGroupProjects(token string, group GitLabConfigGroup) ([]gitrepo.GitRepoSpec, error) {

	rootGroup, err := gitlab.fetchGroupInfo(token, group.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch group info for group %s: %w", group.ID, err)
	}
	allGroups, err := gitlab.FetchAllGroupsRecursively(token, rootGroup)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch all groups recursively %s: %w", group.ID, err)
	}

	allProjects, err := gitlab.fetchProjectsForGroups(token, allGroups)
	if err != nil {
		return nil, err
	}

	return allProjects, nil
}

func (gitlab GitLabConfig) fetchProjects(token, groupID string) ([]gitrepo.GitRepoSpec, error) {
	logger.Log.Debugf("Fetching projects in group %s\n", color.FgCyan(groupID))
	url := fmt.Sprintf("%s/groups/%s/projects", gitlab.getBaseUrl(), groupID)

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
			logger.Log.Errorf("Failed to close response body: %v", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitLab API request failed with status: %s", resp.Status)
	}

	var projects []gitrepo.GitRepoSpec
	if err := json.NewDecoder(resp.Body).Decode(&projects); err != nil {
		return nil, err
	}

	return projects, nil
}

func (gitlab GitLabConfig) fetchSubgroups(token, groupID string) ([]GitlabApiGroup, error) {
	logger.Log.Debugf("Fetching subgroups in group %s", color.FgCyan(groupID))
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
			logger.Log.Errorf("Failed to close response body: %v", err)
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

func (project GitLabConfigProject) AsGitRepoSpec(gitlab GitLabConfig) *gitrepo.GitRepoSpec {
	var gitlabProjectSpec = gitrepo.GitRepoSpec{
		Name:              project.Name,
		PathWithNamespace: project.FullPath,
		SSHURLToRepo:      fmt.Sprintf("git@%s:%s", gitlab.HostName, project.FullPath),
	}

	return &gitlabProjectSpec
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
