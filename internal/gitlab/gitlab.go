package gitlab

import (
	"fmt"
	"sync"
	"tools/internal/color"
	"tools/internal/gitremote"
	. "tools/internal/log"
)

type GitLabConfig struct {
	EnvTokenVariableName string                             `yaml:"tokenEnvVar"`    // The environment variable name for the GitLab token
	HostName             string                             `yaml:"hostName"`       // Gitlab host name
	CloneDirectory       string                             `yaml:"cloneDirectory"` // Where to clone projects in local directory structure
	Groups               []GitLabGroupConfig                `yaml:"groups"`
	Projects             []gitremote.GitRemoteProjectConfig `yaml:"projects"`
	RateLimitPerSecond   int                                `yaml:"rateLimitPerSecond"` // 0 is interpreted as no limit
}

type GitLabGroupConfig struct {
	Name          string `yaml:"name"`
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
	cloneArchived := p.GroupConfig.CloneArchived
	Log.Tracef("Clone archived %s=%t", p.GroupConfig.Name, cloneArchived)
	return cloneArchived
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

func (gitlab GitLabConfig) fetchProjectsForGroup(token string, group GitlabApiGroup, rootGroupConfig GitLabGroupConfig, projectChannel chan ProjectMetadata) {

	var allProjects []ProjectMetadata

	projects, err := gitlab.fetchProjects(token, group)
	if err != nil {
		Log.Printf("Failed to fetch projects for group %s: %v", group.Name, err)
	}
	allProjects = append(allProjects, projects...)
	for _, project := range projects {
		project.Group = group
		project.GitLabConfig = gitlab
		project.GroupConfig = rootGroupConfig
		projectChannel <- project
	}
}

func channelSubgroups(gitlab GitLabConfig, token string, groupId string, gwg *sync.WaitGroup, groupChannel chan GitlabApiGroup) {
	subgroups, err := gitlab.fetchSubgroups(token, groupId)
	if err != nil {
		Log.Errorf(fmt.Sprintf("failed to fetch subgroups for group %s: %w", groupId, err))
	}
	for _, subgroup := range subgroups {
		gwg.Add(1)
		go func() {
			groupChannel <- subgroup
		}()
	}
	// Matching add is where group is sent to channel
	gwg.Done()
}

func (gitlab GitLabConfig) ChannelGroups(token string, rootGroupConfig GitLabGroupConfig, subGroupsChannel chan<- GitlabApiGroup) {

	gwg := sync.WaitGroup{}
	groupChannel := make(chan GitlabApiGroup, 20)

	rootGroup, err := gitlab.fetchGroupInfo(token, rootGroupConfig.Name)
	if err != nil {
		Log.Errorf("failed to fetch rootGroupConfig info for rootGroupConfig %s: %w", rootGroupConfig.Name, err)
	}

	gwg.Add(1)
	go func() {
		// Start by adding root group to the work list
		groupId := rootGroup.ID
		channelSubgroups(gitlab, token, fmt.Sprintf("%d", groupId), &gwg, groupChannel)
	}()

	go func() {
		for {
			receivedGroup, ok := <-groupChannel
			if !ok {
				break
			}
			subGroupsChannel <- receivedGroup
			groupId := receivedGroup.ID
			channelSubgroups(gitlab, token, fmt.Sprintf("%d", groupId), &gwg, groupChannel)
		}
		close(subGroupsChannel)
	}()
	gwg.Wait()
	close(groupChannel)
}

func (gitlab GitLabConfig) ChannelProjects(token string, rootGroupConfig GitLabGroupConfig, gitlabProjectChannel chan ProjectMetadata) {
	pwg := sync.WaitGroup{}
	groupChannel := make(chan GitlabApiGroup, 10)
	pwg.Add(1)
	go func() {
		defer pwg.Done()
		gitlab.ChannelGroups(token, rootGroupConfig, groupChannel)
	}()

	go func() {
		for {
			receivedGroup, ok := <-groupChannel
			if !ok {
				break
			}

			pwg.Add(1)
			go func() {
				defer pwg.Done()
				gitlab.fetchProjectsForGroup(token, receivedGroup, rootGroupConfig, gitlabProjectChannel)
			}()
		}
	}()
	pwg.Wait()
	close(gitlabProjectChannel)

	Log.Debugf("All projects fetched for group ... %s", color.FgGreen(rootGroupConfig.Name))
}

func (gitlab GitLabConfig) fetchProjects(token string, group GitlabApiGroup) ([]ProjectMetadata, error) {
	return gitlabGet[[]ProjectMetadata](token, fmt.Sprintf("%s/groups/%d/projects", gitlab.getBaseUrl(), group.ID))
}

func (gitlab GitLabConfig) fetchSubgroups(token, groupID string) ([]GitlabApiGroup, error) {
	return gitlabGet[[]GitlabApiGroup](token, fmt.Sprintf("%s/groups/%s/subgroups", gitlab.getBaseUrl(), groupID))
}

func (gitlab GitLabConfig) fetchGroupInfo(token string, groupID string) (*GitlabApiGroup, error) {
	return gitlabGet[*GitlabApiGroup](token, fmt.Sprintf("%s/groups/%s", gitlab.getBaseUrl(), groupID))
}
