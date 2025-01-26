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
	errorChannel chan error,
	vm *terminalView.CloneCommandViewModel,
) {

	var cloneChannelsRateLimited []<-chan gitrepo.GitRepo
	for _, gitLabConfig := range config.GitLab {
		absPath, _ := filepath.Abs(gitLabConfig.CloneDirectory)
		cloneViewModel := vm.AddGitLabCloneVM(gitLabConfig.HostName, absPath)
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
			labApi, &gitLabConfig, cloneViewModel.GroupProjectCount, cloneViewModel.GroupCount, errorChannel,
		)
		remoteRepoChannel := channeledApi.ScheduleDirectProjects(cloneViewModel.DirectProjectCount)

		gitlabGroupProjectsChannel := channeledApi.ScheduleGitlabGroupProjectsFetch(gitLabConfig.Groups)
		reposChannel := gitlab.ConvertProjectsToRepos(gitlabGroupProjectsChannel)

		var potentialClonesChannel []<-chan gitrepo.GitRepo
		potentialClonesChannel = append(potentialClonesChannel, reposChannel, remoteRepoChannel)
		in := lo.FanIn(appConfig.DefaultChannelBufferLength, potentialClonesChannel...)
		var cloneChannelRateLimited = channel.RateLimit[gitrepo.GitRepo](
			gitrepo.FilterCloneNeeded(
				in, cloneViewModel.ArchivedCloneCounter, cloneViewModel.CloneCount, errorChannel,
			), gitLabConfig.GetConfiguredCloneRate(), appConfig.DefaultChannelBufferLength,
		)

		cloneChannelsRateLimited = append(cloneChannelsRateLimited, cloneChannelRateLimited)
	}

	gitrepo.CloneRepositories(
		lo.FanIn(appConfig.DefaultChannelBufferLength, cloneChannelsRateLimited...),
		vm.ClonedNowViewModel.ClonedNowCount,
		errorChannel,
	)
}
