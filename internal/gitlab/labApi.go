package gitlab

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"tools/internal/log"
)

type GitlabApiGroup struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ProjectMetadata struct {
	Name              string `json:"name"`
	SSHURLToRepo      string `json:"ssh_url_to_repo"`
	PathWithNamespace string `json:"path_with_namespace"`
	Archived          bool   `json:"archived"`
	Group             *GitlabApiGroup
	GroupConfig       *GitLabGroupConfig
	GitLabConfig      *GitLabConfig
}

func (p ProjectMetadata) CloneArchived() bool {
	cloneArchived := p.GroupConfig.CloneArchived
	return cloneArchived
}

func (p ProjectMetadata) CloneRootDirectory() string {
	return p.GitLabConfig.CloneDirectory
}

/* Repository API manages access to the Gitlab API.
It adheres to the Repository pattern as well - it is at the boundary to external data (Gitlab API).
All methods should be synchronous - channels and pipes handled in other classes/methods.
*/

type RepositoryAPI struct {
	hostName string
	token    string
}

func NewGitlabAPI(token, hostName string) *RepositoryAPI {
	return &RepositoryAPI{
		hostName: hostName,
		token:    token,
	}
}

func (labApi RepositoryAPI) url() string {
	return fmt.Sprintf("https://%s/api/v4", labApi.hostName)
}

func (labApi RepositoryAPI) fetchProjects(group *GitlabApiGroup) ([]ProjectMetadata, error) {
	return gitlabGet[[]ProjectMetadata](labApi.token, fmt.Sprintf("%s/groups/%d/projects", labApi.url(), group.ID))
}

func (labApi RepositoryAPI) fetchSubgroups(groupID string) ([]GitlabApiGroup, error) {
	return gitlabGet[[]GitlabApiGroup](labApi.token, fmt.Sprintf("%s/groups/%s/subgroups", labApi.url(), groupID))
}

func (labApi RepositoryAPI) fetchGroupInfo(groupID string) (*GitlabApiGroup, error) {
	return gitlabGet[*GitlabApiGroup](labApi.token, fmt.Sprintf("%s/groups/%s", labApi.url(), groupID))
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
