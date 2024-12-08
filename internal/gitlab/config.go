package gitlab

import (
	"os"
	"tools/internal/gitremote"
)

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
