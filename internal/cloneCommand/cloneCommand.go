package cloneCommand

import (
	"context"
	"github.com/samber/lo"
	"golang.org/x/term"
	"os"
	"time"
	"tools/internal/appConfig"
	"tools/internal/channel"
	"tools/internal/cloneCommand/terminalView"
	"tools/internal/color"
	"tools/internal/gitlab"
	"tools/internal/gitrepo"
	"tools/internal/log"
	"tools/internal/view"
)

func ExecuteCloneCommand(config *appConfig.AppConfig) {

	startTime := time.Now()

	cloneViewModel := terminalView.NewCloneViewModel()
	// Check if output is a TTY
	isTTY := term.IsTerminal(int(os.Stdout.Fd()))

	timeElapsedView := view.NewTimeElapsedView(startTime, os.Stdout, time.Since)
	r := terminalView.NewCloneView(cloneViewModel, isTTY, os.Stdout, timeElapsedView)
	ctx, stopRenderLoop := context.WithCancel(context.Background())
	if isTTY {
		go r.StartTTYRenderLoop(ctx)
	}

	var cloneChannelsRateLimited []<-chan *gitrepo.Repository
	for _, gitLabConfig := range config.GitLab {

		token := gitLabConfig.RetrieveTokenFromEnv()
		if token == "" {
			logger.Log.Printf("Gitlab token env variable %s not set for %s; skipping", color.FgRed(gitLabConfig.EnvTokenVariableName), color.FgCyan(gitLabConfig.HostName))
			continue
		}

		// logger.Log.Infof("Cloning %s groups & %s projects from %s into %s", color.FgMagenta(fmt.Sprintf("%d", len(gitLabConfig.Groups))), color.FgMagenta(fmt.Sprintf("%d", len(gitLabConfig.Projects))), color.FgCyan(gitLabConfig.HostName), color.FgCyan(gitLabConfig.CloneDirectory))

		err := os.MkdirAll(gitLabConfig.CloneDirectory, os.ModePerm)
		if err != nil {
			logger.Log.Fatalf("Failed to create clone root directory: %v", err)
		}

		labApi := gitlab.NewAPIClient(token, gitLabConfig.HostName)
		channeledApi := gitlab.NewChanneledApi(labApi, &gitLabConfig, cloneViewModel.ProjectCount, cloneViewModel.GroupCount)
		remoteRepoChannel := channeledApi.ScheduleRemoteProjects()

		gitlabGroupProjectsChannel := channeledApi.ScheduleGitlabGroupProjectsFetch(gitLabConfig.Groups)
		reposChannel := gitlab.ConvertProjectsToRepos(gitlabGroupProjectsChannel)

		var potentialClonesChannel []<-chan *gitrepo.Repository
		potentialClonesChannel = append(potentialClonesChannel, reposChannel, remoteRepoChannel)
		in := lo.FanIn(appConfig.DefaultChannelBufferLength, potentialClonesChannel...)
		var cloneChannelRateLimited = channel.RateLimit[*gitrepo.Repository](gitrepo.FilterCloneNeeded(in, cloneViewModel.ArchivedCloneCounter, cloneViewModel.CloneCount), gitLabConfig.GetConfiguredCloneRate(), 10)

		cloneChannelsRateLimited = append(cloneChannelsRateLimited, cloneChannelRateLimited)
	}
	gitrepo.CloneRepositories(lo.FanIn(appConfig.DefaultChannelBufferLength, cloneChannelsRateLimited...), cloneViewModel.ClonedNowCount)
	stopRenderLoop()

	if !isTTY {
		r.RenderNonTTY()
	}
}
