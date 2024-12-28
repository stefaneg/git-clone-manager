package cloneCommand

import (
	"context"
	"gcm/internal/appConfig"
	"gcm/internal/channel"
	"gcm/internal/cloneCommand/terminalView"
	"gcm/internal/gitlab"
	"gcm/internal/gitrepo"
	"gcm/internal/log"
	"gcm/internal/view"
	"github.com/samber/lo"
	"golang.org/x/term"
	"os"
	"path/filepath"
	"time"
)

func ExecuteCloneCommand(config *appConfig.AppConfig) {

	startTime := time.Now()
	timeElapsedView := view.NewTimeElapsedView(startTime, os.Stdout, time.Since)

	// Check if output is a TTY
	isTTY := term.IsTerminal(int(os.Stdout.Fd()))

	var cloneViewModels []view.View

	var cloneChannelsRateLimited []<-chan *gitrepo.Repository
	for _, gitLabConfig := range config.GitLab {
		absPath, _ := filepath.Abs(gitLabConfig.CloneDirectory)
		cloneViewModel := terminalView.NewCloneViewModel(gitLabConfig.HostName, absPath)

		token := gitLabConfig.RetrieveTokenFromEnv()
		if token == "" {
			logger.Log.Printf("Gitlab token env variable %s not set for %s; skipping", gitLabConfig.EnvTokenVariableName, gitLabConfig.HostName)
			continue
		}

		// logger.Log.Infof("Cloning %s groups & %s projects from %s into %s", color.FgMagenta(fmt.Sprintf("%d", len(gitLabConfig.Groups))), color.FgMagenta(fmt.Sprintf("%d", len(gitLabConfig.Projects))), color.FgCyan(gitLabConfig.HostName), color.FgCyan(gitLabConfig.CloneDirectory))

		err := os.MkdirAll(gitLabConfig.CloneDirectory, os.ModePerm)
		if err != nil {
			logger.Log.Fatalf("Failed to create clone root directory: %v", err)
		}

		labApi := gitlab.NewAPIClient(token, gitLabConfig.HostName)
		channeledApi := gitlab.NewChanneledApi(labApi, &gitLabConfig, cloneViewModel.GroupProjectCount, cloneViewModel.GroupCount)
		remoteRepoChannel := channeledApi.ScheduleDirectProjects(cloneViewModel.DirectProjectCount)

		gitlabGroupProjectsChannel := channeledApi.ScheduleGitlabGroupProjectsFetch(gitLabConfig.Groups)
		reposChannel := gitlab.ConvertProjectsToRepos(gitlabGroupProjectsChannel)

		var potentialClonesChannel []<-chan *gitrepo.Repository
		potentialClonesChannel = append(potentialClonesChannel, reposChannel, remoteRepoChannel)
		in := lo.FanIn(appConfig.DefaultChannelBufferLength, potentialClonesChannel...)
		var cloneChannelRateLimited = channel.RateLimit[*gitrepo.Repository](gitrepo.FilterCloneNeeded(in, cloneViewModel.ArchivedCloneCounter, cloneViewModel.CloneCount), gitLabConfig.GetConfiguredCloneRate(), 10)

		cloneChannelsRateLimited = append(cloneChannelsRateLimited, cloneChannelRateLimited)
		cloneView := terminalView.NewCloneView(cloneViewModel, isTTY, os.Stdout)
		cloneViewModels = append(cloneViewModels, cloneView)

	}
	logFilePath, _ := filepath.Abs("gcm.log")
	errvm := terminalView.NewErrorViewModel(logFilePath)
	errview := terminalView.NewErrorView(errvm, os.Stdout)
	cnvm := terminalView.NewClonedNowViewModel()
	cnv := terminalView.NewClonedNowView(cnvm, os.Stdout)
	cloneViewModels = append(cloneViewModels, cnv, timeElapsedView, errview)
	//... plug in error view, direct errors through channel to it

	compositeView := view.NewCompositeView(cloneViewModels)
	ctx, stopRenderLoop := context.WithCancel(context.Background())
	if isTTY {
		go view.StartTTYRenderLoop(compositeView, os.Stdout, ctx)
	}
	gitrepo.CloneRepositories(lo.FanIn(appConfig.DefaultChannelBufferLength, cloneChannelsRateLimited...), cnvm.ClonedNowCount, errvm.ErrorChannel)

	stopRenderLoop()

	if !isTTY {
		compositeView.Render(0)
	}
}
