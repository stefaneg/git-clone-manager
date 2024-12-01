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
	"tools/internal/gitlab"
	"tools/internal/gitrepo"
	. "tools/internal/log"
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

	gitlabProjectChannel := make(chan gitlab.ProjectMetadata, 20)
	gitCloneChannel := make(chan gitrepo.Repository, 20)

	// Receive channelled projects, instantiate git repository for each project
	go pipeProjectsToRepos(gitlabProjectChannel, gitCloneChannel)()

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
			// Get subgroups and projects
			gitlabProjectChannel := make(chan gitlab.ProjectMetadata, 100)
			projectChannels = append(projectChannels, gitlabProjectChannel)
			go gitLabConfig.ChannelProjects(token, group, gitlabProjectChannel)

		}

		// Iterate through directly referred projects
		for _, prj := range gitLabConfig.Projects {
			repo := gitrepo.CreateFromGitRemoteConfig(prj, gitLabConfig.HostName, gitLabConfig.CloneDirectory)
			gitCloneChannel <- *repo
		}
	}
	combinedProjectChannel := lo.FanIn(10, projectChannels...)
	go func() {
		for {
			project, ok := <-combinedProjectChannel
			if !ok {
				Log.Tracef("%s \n", color.FgRed("All projects processed, breaking!!!!"))
				break
			}
			gitlabProjectChannel <- project
		}
		Log.Tracef("Closing combined project channel")
		close(gitlabProjectChannel)
	}()

	cloneWaitGroup := sync.WaitGroup{}
	cloneCount := 0
	for {
		receivedRepo, ok := <-gitCloneChannel
		if !ok {
			Log.Tracef("%s \n", "Clone channel close, wait for last clone to finish, then breaking")
			cloneWaitGroup.Wait()
			break
		}
		cloneCount++
		cloneWaitGroup.Add(1)
		go func() {
			defer cloneWaitGroup.Done()
			// NEXT: Deal with rate limit on git clone operations somehow.
			// Add tests....
			err := receivedRepo.CloneProject()
			if err != nil {
				Log.Errorf("Failed to clone project %s: %v", color.FgRed(receivedRepo.Name), err)
			}
		}()
	}

	Log.Infof(color.FgGreen("%d repos, took %.2f seconds"), cloneCount, time.Since(startTime).Seconds())
}

func pipeProjectsToRepos(gitlabProjectChannel chan gitlab.ProjectMetadata, gitRepoChannel chan gitrepo.Repository) func() {
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
