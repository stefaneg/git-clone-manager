package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"
	"tools/gitlab"
	"tools/internal/color"
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
	}

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
			l.Log.Infof("Cloning projects in group  %s\n", color.FgCyan(group.ID))
			projects, err := gitLabConfig.GetGroupProjects(token, group)

			for _, project := range projects {
				err := project.CloneProject(gitLabConfig.CloneDirectory, group.CloneArchived)
				if err != nil {
					l.Log.Printf("Failed to clone project %s: %v", project.Name, err)
				}
			}

			if err != nil {
				l.Log.Printf("Failed to clone projects for group %s: %v", group.ID, err)
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
