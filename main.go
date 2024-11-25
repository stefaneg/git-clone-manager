package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
	"tools/gitlab"
	"tools/internal/color"
	"tools/internal/gitrepo"
	l "tools/internal/log"
	typex "tools/type"

	"gopkg.in/yaml.v2"
)

type Config struct {
	GitLab []gitlab.GitLabConfig `yaml:"gitlab"`
}

func main() {
	var verbose = typex.NullableBool{}
	startTime := time.Now()

	flag.Var(&verbose, "verbose", "Print verbose output")
	flag.Parse()
	l.InitLogger(verbose.Val(false))

	config, err := loadConfig("workingCopies.yaml")

	if err != nil {
		l.Log.Fatalf("Failed to load configuration: %v", err)
		os.Exit(1)
	}

	fetchWaitGroup := sync.WaitGroup{}

	for _, gitLabConfig := range config.GitLab {
		l.Log.Infof("Cloning groups/projects in %s into %s", color.FgCyan(gitLabConfig.HostName), color.FgCyan(gitLabConfig.CloneDirectory))
		err = os.MkdirAll(gitLabConfig.GetCloneDirectory(), os.ModePerm)
		if err != nil {
			l.Log.Fatalf("Failed to create clone directory: %v", err)
		}

		token := os.Getenv(gitLabConfig.EnvTokenVariableName)
		if token == "" {
			l.Log.Printf("Environment variable %s not set; skipping", gitLabConfig.EnvTokenVariableName)
			continue
		}

		for _, group := range gitLabConfig.Groups {

			groupRepoChannel := make(chan gitrepo.GitRepoSpec, 10)

			go func() {
				for {
					receivedRepo, ok := <-groupRepoChannel
					if !ok {
						l.Log.Debugf("%s %s \n", color.FgRed("Channel close, breaking"), group.ID)
						break
					}
					l.Log.Debugf("Channel receive, cloning %s %t \n", color.FgCyan(receivedRepo.PathWithNamespace), group.CloneArchived)
					go func() {
						fetchWaitGroup.Add(1)
						defer fetchWaitGroup.Done()
						err := receivedRepo.CloneProject(gitLabConfig.CloneDirectory, group.CloneArchived)
						if err != nil {
							l.Log.Printf("Failed to clone project %s: %v", receivedRepo.Name, err)
						}

					}()
				}
			}()

			err := gitLabConfig.GetGroupProjects(token, group, groupRepoChannel, &fetchWaitGroup)

			if err != nil {
				l.Log.Printf("Failed to get projects for group %s: %v", group.ID, err)
			}
		}

		for _, prj := range gitLabConfig.Projects {
			repo := prj.AsGitRepoSpec(gitLabConfig)
			if err != nil {
				l.Log.Printf("Failed to fetch repo %s: %v", prj.FullPath, err)
				continue
			}
			err = repo.CloneProject(gitLabConfig.CloneDirectory, true)
			if err != nil {
				l.Log.Printf("Failed to clone repo %s: %v", repo.Name, err)
			}
		}
	}
	fetchWaitGroup.Wait()
	l.Log.Infof(color.FgGreen("Done in %.2f seconds!"), time.Since(startTime).Seconds())
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
