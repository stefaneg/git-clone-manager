package gitlab

import (
	"os"
	"time"
	"tools/internal/color"
	"tools/internal/ext"
	"tools/internal/gitremote"
	"tools/internal/gitrepo"
	logger "tools/internal/log"
)

// This rate is tested to minimise error rate on cloning 250 repositories.
const DefaultGitlabRateLimit = 7

type GitLabConfig struct {
	EnvTokenVariableName string                             `yaml:"tokenEnvVar"`    // The environment variable name for the GitLab token
	HostName             string                             `yaml:"hostName"`       // Gitlab host name
	CloneDirectory       string                             `yaml:"cloneDirectory"` // Where to clone projects in local directory structure
	Groups               []GitLabGroupConfig                `yaml:"groups"`
	Projects             []gitremote.GitRemoteProjectConfig `yaml:"projects"`
	RateLimitPerSecond   int                                `yaml:"rateLimitPerSecond"` // 0 is interpreted as no limit
}

type GitLabGroupConfig struct {
	Name          string `yaml:"name"`
	CloneArchived bool   `yaml:"cloneArchived"`
}

func (gitLabConfig GitLabConfig) RetrieveTokenFromEnv() string {
	token := os.Getenv(gitLabConfig.EnvTokenVariableName)
	return token
}

func (gitLabConfig GitLabConfig) GetConfiguredCloneRate() int {
	return ext.DefaultValue(gitLabConfig.RateLimitPerSecond, DefaultGitlabRateLimit)
}

func ScheduleRemoteProjects(gitLabConfig GitLabConfig) chan *gitrepo.Repository {
	repoChannel := make(chan *gitrepo.Repository, GroupChannelBufferSize)
	go func() {
		startTime := time.Now()
		for _, prj := range gitLabConfig.Projects {
			repo := gitrepo.CreateFromGitRemoteConfig(prj, gitLabConfig.HostName, gitLabConfig.CloneDirectory)
			repoChannel <- repo
		}
		logger.Log.Infof(color.FgGreen("Repos scheduled in %.2f seconds"), time.Since(startTime).Seconds())

		close(repoChannel)
	}()
	return repoChannel
}
