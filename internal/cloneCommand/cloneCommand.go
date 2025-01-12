package cloneCommand

import (
	"fmt"
	"gcm/internal/appConfig"
	"gcm/internal/channel"
	"gcm/internal/cloneCommand/terminalView"
	"gcm/internal/gitlab"
	"gcm/internal/gitrepo"
	logger "gcm/internal/log"
	"github.com/samber/lo"
	"os"
	"path/filepath"
)

type CloneCommandView struct {
}

func ExecuteCloneCommand(
	config *appConfig.AppConfig,
	cloneView *terminalView.GitLabCloneView,
	clonedNowViewModel *terminalView.ClonedNowViewModel,
	errorChannel chan error,
) {

	var cloneChannelsRateLimited []<-chan *gitrepo.Repository
	for _, gitLabConfig := range config.GitLab {
		absPath, _ := filepath.Abs(gitLabConfig.CloneDirectory)
		cloneViewModel := terminalView.NewGitLabCloneViewModel(gitLabConfig.HostName, absPath)
		cloneView.AddViewModel(cloneViewModel)

		token := gitLabConfig.RetrieveTokenFromEnv()
		if token == "" {
			errorChannel <- fmt.Errorf(
				"Gitlab token env variable %s not set for %s; skipping",
				gitLabConfig.EnvTokenVariableName,
				gitLabConfig.HostName,
			)
			continue
		}

		err := os.MkdirAll(gitLabConfig.CloneDirectory, os.ModePerm)
		if err != nil {
			logger.Log.Fatalf("Failed to create clone root directory: %v", err)
		}

		labApi := gitlab.NewAPIClient(token, gitLabConfig.HostName)
		channeledApi := gitlab.NewChanneledApi(
			labApi, &gitLabConfig, cloneViewModel.GroupProjectCount, cloneViewModel.GroupCount,
		)
		remoteRepoChannel := channeledApi.ScheduleDirectProjects(cloneViewModel.DirectProjectCount)

		gitlabGroupProjectsChannel := channeledApi.ScheduleGitlabGroupProjectsFetch(gitLabConfig.Groups)
		reposChannel := gitlab.ConvertProjectsToRepos(gitlabGroupProjectsChannel)

		var potentialClonesChannel []<-chan *gitrepo.Repository
		potentialClonesChannel = append(potentialClonesChannel, reposChannel, remoteRepoChannel)
		in := lo.FanIn(appConfig.DefaultChannelBufferLength, potentialClonesChannel...)
		var cloneChannelRateLimited = channel.RateLimit[*gitrepo.Repository](
			gitrepo.FilterCloneNeeded(
				in, cloneViewModel.ArchivedCloneCounter, cloneViewModel.CloneCount,
			), gitLabConfig.GetConfiguredCloneRate(), appConfig.DefaultChannelBufferLength,
		)

		cloneChannelsRateLimited = append(cloneChannelsRateLimited, cloneChannelRateLimited)
	}

	gitrepo.CloneRepositories(
		lo.FanIn(appConfig.DefaultChannelBufferLength, cloneChannelsRateLimited...), clonedNowViewModel.ClonedNowCount,
		errorChannel,
	)
}
