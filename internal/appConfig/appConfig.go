package appConfig

import "gcm/internal/gitlab"

const DefaultChannelBufferLength = 10

type AppConfig struct {
	GitLab []gitlab.GitLabConfig `yaml:"gitlab"`
}
