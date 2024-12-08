package gitlab

import (
	"fmt"
	"os"
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
	return cloneArchived
}

func (p ProjectMetadata) CloneRootDirectory() string {
	return p.GitLabConfig.CloneDirectory
}

func (gitLabConfig GitLabConfig) RetrieveTokenFromEnv() string {
	token := os.Getenv(gitLabConfig.EnvTokenVariableName)
	return token
}

func fetchProjectsForGroup(gitlab *RepositoryAPI, group GitlabApiGroup, rootGroupConfig GitLabGroupConfig, projectChannel chan ProjectMetadata, gitlabConfig GitLabConfig) {

	var allProjects []ProjectMetadata

	projects, err := gitlab.fetchProjects(group)
	if err != nil {
		Log.Printf("Failed to fetch projects for group %s: %v", group.Name, err)
	}
	allProjects = append(allProjects, projects...)
	for _, project := range projects {
		project.Group = group
		project.GitLabConfig = gitlabConfig
		project.GroupConfig = rootGroupConfig
		projectChannel <- project
	}
}

func channelSubgroups(gitlab *RepositoryAPI, groupId string, gwg *sync.WaitGroup, groupChannel chan GitlabApiGroup) {
	subgroups, err := gitlab.fetchSubgroups(groupId)
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

func ChannelGroups(gitlab *RepositoryAPI, rootGroupConfig GitLabGroupConfig, subGroupsChannel chan<- GitlabApiGroup) {

	gwg := sync.WaitGroup{}
	groupChannel := make(chan GitlabApiGroup, 20)

	rootGroup, err := gitlab.fetchGroupInfo(rootGroupConfig.Name)
	if err != nil {
		Log.Errorf("failed to fetch rootGroupConfig info for rootGroupConfig %s: %w", rootGroupConfig.Name, err)
	}

	gwg.Add(1)
	go func() {
		// Start by adding root group to the work list
		groupId := rootGroup.ID
		channelSubgroups(gitlab, fmt.Sprintf("%d", groupId), &gwg, groupChannel)
	}()

	go func() {
		for {
			receivedGroup, ok := <-groupChannel
			if !ok {
				break
			}
			subGroupsChannel <- receivedGroup
			groupId := receivedGroup.ID
			channelSubgroups(gitlab, fmt.Sprintf("%d", groupId), &gwg, groupChannel)
		}
		close(subGroupsChannel)
	}()
	gwg.Wait()
	close(groupChannel)
}

func ChannelProjects(gitlab *RepositoryAPI, rootGroupConfig GitLabGroupConfig, gitlabProjectChannel chan ProjectMetadata, gitlabConfig GitLabConfig) {
	pwg := sync.WaitGroup{}
	groupChannel := make(chan GitlabApiGroup, 10)
	pwg.Add(1)
	go func() {
		defer pwg.Done()
		ChannelGroups(gitlab, rootGroupConfig, groupChannel)
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
				fetchProjectsForGroup(gitlab, receivedGroup, rootGroupConfig, gitlabProjectChannel, gitlabConfig)
			}()
		}
	}()
	pwg.Wait()
	close(gitlabProjectChannel)

	Log.Debugf("All projects fetched for group ... %s", color.FgGreen(rootGroupConfig.Name))
}
