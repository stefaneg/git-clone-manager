package main

import (
	"flag"
	"fmt"
	"github.com/samber/lo"
	"os"
	"path/filepath"
	"sync"
	"time"
	"tools/internal/color"
	"tools/internal/counter"
	"tools/internal/ext"
	"tools/internal/gitlab"
	"tools/internal/gitrepo"
	. "tools/internal/log"
	"tools/internal/pipe"
	typex "tools/type"

	"gopkg.in/yaml.v2"
)

type Config struct {
	GitLab []gitlab.GitLabConfig `yaml:"gitlab"`
}

const DefaultGitlabRateLimit = 9

func main() {

	//f, _ := os.Create("trace.out")
	//defer f.Close()
	//trace.Start(f)
	//defer trace.Stop()

	// Process parameters
	var verbose = typex.NullableBool{}
	startTime := time.Now()

	flag.Var(&verbose, "verbose", "Print verbose output")
	flag.Parse()
	InitLogger(verbose.Val(false))

	config, err := loadConfig("workingCopies.yaml")

	if err != nil {
		Log.Fatalf("Failed to load configuration: %v", err)
		os.Exit(1)
	}

	// Set up channel/pipes
	// gitlab projects -> convert to repos -> check existence/archival state -> clone

	var cloneChannelsRateLimited []<-chan gitrepo.Repository
	for _, gitLabConfig := range config.GitLab {

		token := gitLabConfig.RetrieveTokenFromEnv()
		if token == "" {
			Log.Printf("Gitlab token env variable %s not set for %s; skipping", color.FgRed(gitLabConfig.EnvTokenVariableName), color.FgCyan(gitLabConfig.HostName))
			continue
		}
		labApi := gitlab.NewGitlabAPI(token, gitLabConfig.HostName)

		Log.Infof("Cloning %s groups & %s projects from %s into %s", color.FgMagenta(fmt.Sprintf("%d", len(gitLabConfig.Groups))), color.FgMagenta(fmt.Sprintf("%d", len(gitLabConfig.Projects))), color.FgCyan(gitLabConfig.HostName), color.FgCyan(gitLabConfig.CloneDirectory))

		// Working copy manager responsibility
		err = os.MkdirAll(gitLabConfig.CloneDirectory, os.ModePerm)
		if err != nil {
			Log.Fatalf("Failed to create clone directory: %v", err)
		}

		// Clone/pull rate is a general git concern - not gitlab specific
		var cloneRatePerSecond = ext.DefaultValue(gitLabConfig.RateLimitPerSecond, DefaultGitlabRateLimit)

		gitlabProjectChannel := make(chan gitlab.ProjectMetadata, 20)
		checkCloneChannel := make(chan gitrepo.Repository, 20)

		// Start piping projects to the checkCloneChannel - which will again channel to clone.
		go convertProjectsAndChannelToRepos(gitlabProjectChannel, checkCloneChannel)()

		channeledApi := gitlab.NewChanneledApi(labApi, &gitLabConfig)
		// Iterate through configured groups and pipe fetched projects to the gitlabProjectChannel
		var projectChannels []<-chan gitlab.ProjectMetadata

		for _, group := range gitLabConfig.Groups {
			projectChannels = append(projectChannels, channeledApi.FetchAndChannelGroupProjects(&group))

		}
		forwardChannels(projectChannels, gitlabProjectChannel, 10)

		_, gitCloneChannel := filterCloneNeeded(checkCloneChannel)
		var cloneChannelRateLimited = pipe.RateLimit[gitrepo.Repository](gitCloneChannel, cloneRatePerSecond, 10)
		cloneChannelsRateLimited = append(cloneChannelsRateLimited, cloneChannelRateLimited)

		// Iterate through directly referred projects
		for _, prj := range gitLabConfig.Projects {
			repo := gitrepo.CreateFromGitRemoteConfig(prj, gitLabConfig.HostName, gitLabConfig.CloneDirectory)
			checkCloneChannel <- *repo
		}
	}

	cloneCount := cloneGitRepos(lo.FanIn(10, cloneChannelsRateLimited...))

	Log.Infof(color.FgGreen("%d repos, %d cloned. %.2f seconds"), 999, cloneCount, time.Since(startTime).Seconds())
}

func forwardChannels[T any](inboundChannels []<-chan T, outputChannel chan T, channelBufferCap int) {
	combinedChannel := lo.FanIn(channelBufferCap, inboundChannels...)
	go func() {
		for {
			project, ok := <-combinedChannel
			if !ok {
				break
			}
			outputChannel <- project
		}
		close(outputChannel)
	}()
}

func filterCloneNeeded(checkCloneChannel chan gitrepo.Repository) (*counter.Counter, chan gitrepo.Repository) {
	gitCloneChannel := make(chan gitrepo.Repository, 20)
	projectCounter := counter.NewCounter()
	checkWaitGroup := sync.WaitGroup{}
	go func() {
		for {
			receivedRepo, ok := <-checkCloneChannel
			if !ok {
				Log.Tracef("%s \n", "Clone channel close, wait for last clone to finish, then breaking")
				break
			}
			projectCounter.Add(1)
			needsCloning, _ := receivedRepo.CheckNeedsCloning()
			if needsCloning {
				checkWaitGroup.Add(1)
				go func() {
					defer checkWaitGroup.Done()
					Log.Debugf("Adding %s to clone queue ", receivedRepo.Name)
					gitCloneChannel <- receivedRepo
				}()
			} else {
				Log.Tracef("%s \n", "Clone not needed")
			}
		}
		checkWaitGroup.Wait()
		close(gitCloneChannel)
	}()
	return projectCounter, gitCloneChannel
}

func cloneGitRepos(rateLimitedClone <-chan gitrepo.Repository) int {
	cloneWaitGroup := sync.WaitGroup{}
	cloneCount := 0
	for {
		receivedRepo, ok := <-rateLimitedClone
		if !ok {
			break
		}
		cloneCount++
		cloneWaitGroup.Add(1)
		go func() {
			defer cloneWaitGroup.Done()
			err := receivedRepo.CloneProject()
			if err != nil {
				Log.Errorf("Failed to clone project %s: %v", color.FgRed(receivedRepo.Name), err)
			}
		}()
	}
	cloneWaitGroup.Wait()
	return cloneCount
}

func convertProjectsAndChannelToRepos(gitlabProjectChannel chan gitlab.ProjectMetadata, gitRepoChannel chan gitrepo.Repository) func() {
	return func() {
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
			gitRepoChannel <- gitRepo
		}
		close(gitRepoChannel)
	}
}

func loadConfig(configFileName string) (*Config, error) {
	configFilePath := filepath.Join("./", configFileName)

	if _, err := os.Stat(configFileName); os.IsNotExist(err) {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("could not determine home directory: %v", err)
		}
		configFilePath = filepath.Join(homeDir, configFileName)
		if _, err := os.Stat(configFileName); os.IsNotExist(err) {
			return nil, fmt.Errorf("config file not found in current directory or home directory")
		}
	}

	data, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("could not read config file: %v", err)
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal config file: %v", err)
	}

	return &config, nil
}
