package gitlab

import (
	"fmt"
	"sync"
	"tools/internal/color"
	. "tools/internal/log"
)

type ChanneledApi struct {
	api *RepositoryAPI
}

func NewChanneledApi(repo *RepositoryAPI) *ChanneledApi {
	return &ChanneledApi{api: repo}
}

func (channeledApi *ChanneledApi) fetchProjectsForGroup(group *GitlabApiGroup, rootGroupConfig *GitLabGroupConfig, projectChannel chan ProjectMetadata, gitlabConfig *GitLabConfig) {

	var allProjects []ProjectMetadata

	projects, err := channeledApi.api.fetchProjects(group)
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

func (channeledApi *ChanneledApi) channelSubgroups(groupId string, gwg *sync.WaitGroup, groupChannel chan *GitlabApiGroup) {
	subgroups, err := channeledApi.api.fetchSubgroups(groupId)
	if err != nil {
		Log.Errorf(fmt.Sprintf("failed to fetch subgroups for group %s: %w", groupId, err))
	}
	for _, subgroup := range subgroups {
		gwg.Add(1)
		go func() {
			groupChannel <- &subgroup
		}()
	}
	// Matching add is where group is sent to channel
	gwg.Done()
}

func (channeledApi *ChanneledApi) channelGroups(rootGroupConfig *GitLabGroupConfig, subGroupsChannel chan<- *GitlabApiGroup, api *ChanneledApi) {

	gwg := sync.WaitGroup{}
	groupChannel := make(chan *GitlabApiGroup, 20)

	rootGroup, err := api.api.fetchGroupInfo(rootGroupConfig.Name)
	if err != nil {
		Log.Errorf("failed to fetch rootGroupConfig info for rootGroupConfig %s: %w", rootGroupConfig.Name, err)
	}

	gwg.Add(1)
	go func() {
		// Start by adding root group to the work list
		groupId := rootGroup.ID
		channeledApi.channelSubgroups(fmt.Sprintf("%d", groupId), &gwg, groupChannel)
	}()

	go func() {
		for {
			receivedGroup, ok := <-groupChannel
			if !ok {
				break
			}
			subGroupsChannel <- receivedGroup
			groupId := receivedGroup.ID
			channeledApi.channelSubgroups(fmt.Sprintf("%d", groupId), &gwg, groupChannel)
		}
		close(subGroupsChannel)
	}()
	gwg.Wait()
	close(groupChannel)
}

func (channeledApi *ChanneledApi) FetchAndChannelGroupProjects(rootGroupConfig *GitLabGroupConfig, gitlabProjectChannel chan ProjectMetadata, gitlabConfig *GitLabConfig) {
	pwg := sync.WaitGroup{}
	groupChannel := make(chan *GitlabApiGroup, 10)
	pwg.Add(1)
	go func() {
		defer pwg.Done()
		channeledApi.channelGroups(rootGroupConfig, groupChannel, channeledApi)
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
				channeledApi.fetchProjectsForGroup(receivedGroup, rootGroupConfig, gitlabProjectChannel, gitlabConfig)
			}()
		}
	}()
	pwg.Wait()
	close(gitlabProjectChannel)

	Log.Debugf("All projects fetched for group ... %s", color.FgGreen(rootGroupConfig.Name))
}
