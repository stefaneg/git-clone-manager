package gitlab

import (
	"encoding/json"
	"fmt"
	"gcm/internal/log"
	"io"
	"net/http"
)

type Group struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Project struct {
	Name              string `json:"name"`
	SSHURLToRepo      string `json:"ssh_url_to_repo"`
	PathWithNamespace string `json:"path_with_namespace"`
	Archived          bool   `json:"archived"`
	Group             *Group
	GroupConfig       *GroupConfig
	GitLabConfig      *GitLabConfig
}

func (p Project) CloneArchived() bool {
	cloneArchived := p.GroupConfig.CloneArchived
	return cloneArchived
}

func (p Project) CloneRootDirectory() string {
	return p.GitLabConfig.CloneDirectory
}

/* Repository API manages access to the Gitlab API.
It adheres to the Repository pattern as well - it is at the boundary to external data (Gitlab API).
All methods should be synchronous - channels and pipes handled in other classes/methods.
*/

type APIClient struct {
	hostName string
	token    string
}

func NewAPIClient(token, hostName string) *APIClient {
	return &APIClient{
		hostName: hostName,
		token:    token,
	}
}

func (apiClient APIClient) url() string {
	return fmt.Sprintf("https://%s/api/v4", apiClient.hostName)
}

func (apiClient APIClient) fetchProjects(group *Group) ([]Project, error) {
	return gitlabGet[[]Project](apiClient.token, fmt.Sprintf("%s/groups/%d/projects", apiClient.url(), group.ID))
}

func (apiClient APIClient) fetchSubgroups(groupID string) ([]Group, error) {
	return gitlabGet[[]Group](apiClient.token, fmt.Sprintf("%s/groups/%s/subgroups", apiClient.url(), groupID))
}

func (apiClient APIClient) fetchGroupInfo(groupID string) (*Group, error) {
	return gitlabGet[*Group](apiClient.token, fmt.Sprintf("%s/groups/%s", apiClient.url(), groupID))
}

func gitlabGet[T any](token string, url string) (T, error) {
	var emptyResult T
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return emptyResult, err
	}
	req.Header.Set("PRIVATE-TOKEN", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return emptyResult, err
	}
	defer func(body io.ReadCloser) {
		err := body.Close()
		if err != nil {
			logger.Log.Errorf("Failed to close response body: %v", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return emptyResult, fmt.Errorf("GitLab API request on %s failed with status: %s", url, resp.Status)
	}

	var decodedResult T
	if err := json.NewDecoder(resp.Body).Decode(&decodedResult); err != nil {
		return emptyResult, err
	}

	return decodedResult, nil
}
