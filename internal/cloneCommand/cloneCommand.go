package cloneCommand

import (
	"fmt"
	"github.com/samber/lo"
	"os"
	"time"
	"tools/internal/appConfig"
	"tools/internal/channel"
	"tools/internal/color"
	"tools/internal/gitlab"
	"tools/internal/gitrepo"
	"tools/internal/log"
)

func ExecuteCloneCommand(config *appConfig.AppConfig) {
	startTime := time.Now()
	var cloneChannelsRateLimited []<-chan *gitrepo.Repository
	for _, gitLabConfig := range config.GitLab {

		token := gitLabConfig.RetrieveTokenFromEnv()
		if token == "" {
			logger.Log.Printf("Gitlab token env variable %s not set for %s; skipping", color.FgRed(gitLabConfig.EnvTokenVariableName), color.FgCyan(gitLabConfig.HostName))
			continue
		}

		logger.Log.Infof("Cloning %s groups & %s projects from %s into %s", color.FgMagenta(fmt.Sprintf("%d", len(gitLabConfig.Groups))), color.FgMagenta(fmt.Sprintf("%d", len(gitLabConfig.Projects))), color.FgCyan(gitLabConfig.HostName), color.FgCyan(gitLabConfig.CloneDirectory))

		err := os.MkdirAll(gitLabConfig.CloneDirectory, os.ModePerm)
		if err != nil {
			logger.Log.Fatalf("Failed to create clone directory: %v", err)
		}

		labApi := gitlab.NewGitlabAPI(token, gitLabConfig.HostName)
		channeledApi := gitlab.NewChanneledApi(labApi, &gitLabConfig)

		gitlabGroupProjectsChannel := channeledApi.ScheduleGitlabGroupProjectsFetch(gitLabConfig.Groups)
		reposChannel := gitlab.ConvertProjectsToRepos(gitlabGroupProjectsChannel)

		remoteRepoChannel := gitlab.ScheduleRemoteProjects(gitLabConfig)

		var potentialClonesChannel []<-chan *gitrepo.Repository
		potentialClonesChannel = append(potentialClonesChannel, reposChannel, remoteRepoChannel)
		in := lo.FanIn(appConfig.DefaultChannelBufferLength, potentialClonesChannel...)
		var cloneChannelRateLimited = channel.RateLimit[*gitrepo.Repository](gitrepo.FilterCloneNeeded(in), gitLabConfig.GetConfiguredCloneRate(), 10)

		cloneChannelsRateLimited = append(cloneChannelsRateLimited, cloneChannelRateLimited)
	}
	cloneCount := gitrepo.CloneRepositories(lo.FanIn(appConfig.DefaultChannelBufferLength, cloneChannelsRateLimited...))

	logger.Log.Infof(color.FgGreen("%d repos, %d cloned. %.2f seconds"), 999, cloneCount, time.Since(startTime).Seconds())
}
