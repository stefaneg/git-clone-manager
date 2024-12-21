package cloneCommand

import (
	"fmt"
	"github.com/samber/lo"
	"os"
	"time"
	"tools/internal/appConfig"
	"tools/internal/channel"
	"tools/internal/color"
	"tools/internal/counter"
	"tools/internal/gitlab"
	"tools/internal/gitrepo"
	"tools/internal/log"
)

func ExecuteCloneCommand(config *appConfig.AppConfig) {
	startTime := time.Now()
	clonedNowCounter := counter.NewCounter()
	cloneCounter := counter.NewCounter()
	archivedClonesCounter := counter.NewCounter()
	projectCounter := counter.NewCounter()
	groupCounter := counter.NewCounter()

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
			logger.Log.Fatalf("Failed to create clone root directory: %v", err)
		}

		labApi := gitlab.NewGitlabAPI(token, gitLabConfig.HostName)
		channeledApi := gitlab.NewChanneledApi(labApi, &gitLabConfig, projectCounter, groupCounter)
		remoteRepoChannel := channeledApi.ScheduleRemoteProjects()

		gitlabGroupProjectsChannel := channeledApi.ScheduleGitlabGroupProjectsFetch(gitLabConfig.Groups)
		reposChannel := gitlab.ConvertProjectsToRepos(gitlabGroupProjectsChannel)

		var potentialClonesChannel []<-chan *gitrepo.Repository
		potentialClonesChannel = append(potentialClonesChannel, reposChannel, remoteRepoChannel)
		in := lo.FanIn(appConfig.DefaultChannelBufferLength, potentialClonesChannel...)
		var cloneChannelRateLimited = channel.RateLimit[*gitrepo.Repository](gitrepo.FilterCloneNeeded(in, archivedClonesCounter, cloneCounter), gitLabConfig.GetConfiguredCloneRate(), 10)

		cloneChannelsRateLimited = append(cloneChannelsRateLimited, cloneChannelRateLimited)
	}
	gitrepo.CloneRepositories(lo.FanIn(appConfig.DefaultChannelBufferLength, cloneChannelsRateLimited...), clonedNowCounter)

	logger.Log.Infof("%s projects in %s groups\n%s git clones (%s archived) \n%s cloned now\n%s seconds",
		color.FgMagenta(fmt.Sprintf("%d", projectCounter.Count())),
		color.FgMagenta(fmt.Sprintf("%d", groupCounter.Count())),
		color.FgMagenta(fmt.Sprintf("%d", cloneCounter.Count())),
		color.FgMagenta(fmt.Sprintf("%d", archivedClonesCounter.Count())),
		color.FgMagenta(fmt.Sprintf("%d", clonedNowCounter.Count())),
		color.FgGreen(fmt.Sprintf("%.2f", time.Since(startTime).Seconds())))
}
