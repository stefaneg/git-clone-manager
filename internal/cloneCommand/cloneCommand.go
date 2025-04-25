package cloneCommand

import (
	"fmt"
	"gcm/internal/appConfig"
	"gcm/internal/channel"
	"gcm/internal/cloneCommand/terminalView"
	"gcm/internal/fs"
	"gcm/internal/gitlab"
	"gcm/internal/gitrepo"
	logger "gcm/internal/log"
	"gcm/internal/sh"
	"github.com/samber/lo"
	"path/filepath"
)

// GetEnvFunc is the type of os.Getenv
type GetEnvFunc func(string) string
type CloneCommand struct {
	vm            *terminalView.CloneCommandViewModel
	filesystem    fs.FileSystem
	apiFactory    gitlab.APIFactory
	getEnv        GetEnvFunc
	commandRunner sh.CommandRunner
}

func NewCloneCommand(
	vm *terminalView.CloneCommandViewModel,
	filesystem fs.FileSystem,
	apiFactory gitlab.APIFactory,
	getEnv GetEnvFunc,
	commandRunner sh.CommandRunner,
) *CloneCommand {
	return &CloneCommand{
		vm:            vm,
		filesystem:    filesystem,
		apiFactory:    apiFactory,
		getEnv:        getEnv,
		commandRunner: commandRunner,
	}
}

func (command *CloneCommand) Execute(config *appConfig.AppConfig) {
	errorChannel := command.vm.ErrorViewModel.ErrorChannel
	filesystem := command.filesystem
	var cloneChannelsRateLimited []<-chan gitrepo.GitRepo
	for _, gitLabConfig := range config.GitLab {
		absPath, _ := filepath.Abs(gitLabConfig.CloneDirectory)
		cloneViewModel := command.vm.AddGitLabCloneVM(gitLabConfig.HostName, absPath)
		token := command.getEnv(gitLabConfig.EnvTokenVariableName)
		if token == "" {
			errorChannel <- fmt.Errorf(
				"Gitlab token env variable %s not set for %s; skipping",
				gitLabConfig.EnvTokenVariableName,
				gitLabConfig.HostName,
			)
			continue
		}
		err := filesystem.MkDir(fs.DirectoryPath(gitLabConfig.CloneDirectory))
		if err != nil {
			logger.Log.Fatalf("Failed to create clone root directory: %v", err)
		}

		labApi := command.apiFactory(token, gitLabConfig.HostName)
		channeledApi := gitlab.NewChanneledApi(
			labApi,
			filesystem,
			&gitLabConfig,
			cloneViewModel.GroupProjectCount,
			cloneViewModel.GroupCount,
			errorChannel,
		)
		remoteRepoChannel := channeledApi.ScheduleDirectProjects(cloneViewModel.DirectProjectCount)

		gitlabGroupProjectsChannel := channeledApi.ScheduleFetchGitlabGroupProjects(gitLabConfig.Groups)
		reposChannel := gitlab.ConvertProjectsToRepos(gitlabGroupProjectsChannel, filesystem)

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

	fanInRepos := lo.FanIn(appConfig.DefaultChannelBufferLength, cloneChannelsRateLimited...)

	gitrepo.CloneRepositories(
		fanInRepos,
		command.vm.ClonedNowViewModel.ClonedNowCount,
		errorChannel,
		command.commandRunner,
	)
}
