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

	gitlabProjectChannel := make(chan gitlab.ProjectMetadata, 20)
	checkCloneChannel := make(chan gitrepo.Repository, 20)

	go convertProjectsToRepos(gitlabProjectChannel, checkCloneChannel)()

	var projectChannels []<-chan gitlab.ProjectMetadata

	for _, gitLabConfig := range config.GitLab {
		Log.Infof("Cloning groups/projects in %s into %s", color.FgCyan(gitLabConfig.HostName), color.FgCyan(gitLabConfig.CloneDirectory))
		err = os.MkdirAll(gitLabConfig.GetCloneDirectory(), os.ModePerm)
		if err != nil {
			Log.Fatalf("Failed to create clone directory: %v", err)
		}

		token := os.Getenv(gitLabConfig.EnvTokenVariableName)
		if token == "" {
			Log.Printf("Environment variable %s not set; skipping", gitLabConfig.EnvTokenVariableName)
			continue
		}

		// Iterate through configured groups and fetch gitlab groups
		for _, group := range gitLabConfig.Groups {
			gitlabProjectChannel := make(chan gitlab.ProjectMetadata, 100)
			projectChannels = append(projectChannels, gitlabProjectChannel)
			go gitLabConfig.ChannelProjects(token, group, gitlabProjectChannel)
		}

		// Iterate through directly referred projects
		for _, prj := range gitLabConfig.Projects {
			repo := gitrepo.CreateFromGitRemoteConfig(prj, gitLabConfig.HostName, gitLabConfig.CloneDirectory)
			checkCloneChannel <- *repo
		}
	}
	forwardChannels(projectChannels, gitlabProjectChannel, 10)
	projectCounter, gitCloneChannel := filterCloneNeeded(checkCloneChannel)
	gitlabCloneRatePerSecond := 7
	var cloneChannelRateLimited <-chan gitrepo.Repository = pipe.RateLimit[gitrepo.Repository](gitCloneChannel, gitlabCloneRatePerSecond, 10)
	cloneCount := cloneGitRepos(cloneChannelRateLimited)

	Log.Infof(color.FgGreen("%d repos, %d cloned. %.2f seconds"), projectCounter.Count(), cloneCount, time.Since(startTime).Seconds())
}

func forwardChannels[T any](projectChannels []<-chan T, gitlabProjectChannel chan T, channelBufferCap int) {
	combinedProjectChannel := lo.FanIn(channelBufferCap, projectChannels...)
	go func() {
		for {
			project, ok := <-combinedProjectChannel
			if !ok {
				break
			}
			gitlabProjectChannel <- project
		}
		close(gitlabProjectChannel)
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

func convertProjectsToRepos(gitlabProjectChannel chan gitlab.ProjectMetadata, gitRepoChannel chan gitrepo.Repository) func() {
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
