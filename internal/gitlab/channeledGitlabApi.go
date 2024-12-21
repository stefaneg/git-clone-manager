package gitlab

import (
	"fmt"
	"github.com/samber/lo"
	"sync"
	"tools/internal/color"
	"tools/internal/gitrepo"
	. "tools/internal/log"
)

const GroupChannelBufferSize = 20
const ProjectChannelBufferSize = 20

type ChanneledApi struct {
	api    *RepositoryAPI
	config *GitLabConfig
}

// NEXT: ADD Reporting counters and error channel handler...

func NewChanneledApi(repo *RepositoryAPI, config *GitLabConfig) *ChanneledApi {
	return &ChanneledApi{api: repo, config: config}
}

func (channeledApi *ChanneledApi) fetchProjectsForGroup(group *GitlabApiGroup, rootGroupConfig *GitLabGroupConfig, projectChannel chan ProjectMetadata) {
	projects, err := channeledApi.api.fetchProjects(group)
	if err != nil {
		Log.Printf("Failed to fetch projects for group %s: %v", group.Name, err)
	}
	for _, project := range projects {
		project.Group = group
		project.GitLabConfig = channeledApi.config
		project.GroupConfig = rootGroupConfig
		projectChannel <- project
	}
}

func (channeledApi *ChanneledApi) channelSubgroups(groupId string, gwg *sync.WaitGroup, groupChannel chan *GitlabApiGroup) {
	subgroups, err := channeledApi.api.fetchSubgroups(groupId)
	if err != nil {
		Log.Errorf("failed to fetch subgroups for group %s: %v", groupId, err)
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

func (channeledApi *ChanneledApi) channelGroups(rootGroupConfig *GitLabGroupConfig, subGroupsChannel chan<- *GitlabApiGroup) {

	gwg := sync.WaitGroup{}
	groupWorkList := make(chan *GitlabApiGroup, GroupChannelBufferSize)

	rootGroup, err := channeledApi.api.fetchGroupInfo(rootGroupConfig.Name)
	if err != nil {
		Log.Errorf("failed to fetch rootGroupConfig info for rootGroupConfig %s: %v", rootGroupConfig.Name, err)
	}

	// Matching Done is where subgroups have been fetched and all sent to fetch channel
	gwg.Add(1)

	go func() {
		// Start by adding root group to the work list
		groupId := rootGroup.ID
		channeledApi.channelSubgroups(fmt.Sprintf("%d", groupId), &gwg, groupWorkList)
	}()

	go func() {
		for {
			receivedGroup, ok := <-groupWorkList
			if !ok {
				break
			}
			subGroupsChannel <- receivedGroup
			groupId := receivedGroup.ID
			channeledApi.channelSubgroups(fmt.Sprintf("%d", groupId), &gwg, groupWorkList)
		}
		close(subGroupsChannel)
	}()
	gwg.Wait()
	close(groupWorkList)
}

func (channeledApi *ChanneledApi) FetchAndChannelGroupProjects(rootGroupConfig *GitLabGroupConfig) chan ProjectMetadata {
	pwg := sync.WaitGroup{}
	groupChannel := make(chan *GitlabApiGroup, GroupChannelBufferSize)
	gitlabProjectChannel := make(chan ProjectMetadata, ProjectChannelBufferSize)
	pwg.Add(1)
	go func() {
		defer pwg.Done()
		channeledApi.channelGroups(rootGroupConfig, groupChannel)
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
				channeledApi.fetchProjectsForGroup(receivedGroup, rootGroupConfig, gitlabProjectChannel)
			}()
		}
		pwg.Wait()
		close(gitlabProjectChannel)
	}()

	Log.Debugf("All projects fetched for group ... %s", color.FgGreen(rootGroupConfig.Name))
	return gitlabProjectChannel
}

func (channeledApi *ChanneledApi) ScheduleGitlabGroupProjectsFetch(groups []GitLabGroupConfig) <-chan ProjectMetadata {
	var projectChannels []<-chan ProjectMetadata
	for _, group := range groups {
		projectChannels = append(projectChannels, channeledApi.FetchAndChannelGroupProjects(&group))
	}
	return lo.FanIn(ProjectChannelBufferSize, projectChannels...)
}

func ConvertProjectsToRepos(gitlabProjectChannel <-chan ProjectMetadata) chan *gitrepo.Repository {
	gitRepoChannel := make(chan *gitrepo.Repository, 10)

	go func() {
		for {
			receivedProject, ok := <-gitlabProjectChannel
			if !ok {
				break
			}
			gitRepo := gitrepo.Repository{
				Name:              receivedProject.Name,
				SSHURLToRepo:      receivedProject.SSHURLToRepo,
				PathWithNamespace: receivedProject.PathWithNamespace,
				Archived:          receivedProject.Archived,
				CloneOptions:      receivedProject,
			}
			gitRepoChannel <- &gitRepo
		}
		close(gitRepoChannel)
	}()
	return gitRepoChannel
}

func (channeledApi *ChanneledApi) ScheduleRemoteProjects() chan *gitrepo.Repository {
	repoChannel := make(chan *gitrepo.Repository, GroupChannelBufferSize)
	go func() {
		for _, prj := range channeledApi.config.Projects {
			repo := gitrepo.CreateFromGitRemoteConfig(prj, channeledApi.config.HostName, channeledApi.config.CloneDirectory)
			repoChannel <- repo
		}

		close(repoChannel)
	}()
	return repoChannel
}
