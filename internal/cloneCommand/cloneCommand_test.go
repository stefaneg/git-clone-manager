package cloneCommand

import (
	"fmt"
	"gcm/internal/appConfig"
	"gcm/internal/cloneCommand/terminalView"
	"gcm/internal/counter"
	"gcm/internal/gitlab"
	"gcm/internal/gitremote"
	"gcm/internal/view"
	"testing"
)

func NewFakeCloneCommandViewModel() *terminalView.CloneCommandViewModel {
	return &terminalView.CloneCommandViewModel{
		GitLabCloneViewModels: make([]*terminalView.GitLabCloneViewModel, 0),
		ClonedNowViewModel:    terminalView.NewClonedNowViewModel(),
		ErrorViewModel: &view.ErrorViewModel{
			ErrorCount:   counter.NewCounter(),
			ErrorChannel: make(chan error, appConfig.DefaultChannelBufferLength),
		},
	}
}

func TestCloneCommandConfigErrorHandling(t *testing.T) {
	appCfg := &appConfig.AppConfig{
		GitLab: []gitlab.GitLabConfig{
			{
				EnvTokenVariableName: "GITLAB_TOKEN_ENV_VAR_NOT_SET",
				HostName:             "missing-token.gitlab.com",
				CloneDirectory:       "/path/to/clone",
				Groups: []gitlab.GroupConfig{
					{
						Name:          "example-group",
						CloneArchived: false,
					},
				},
				Projects: []gitremote.ProjectConfig{
					{
						Name: "example-project",
					},
				},
				RateLimitPerSecond: 10,
			},
		},
	}

	testView := NewFakeCloneCommandViewModel()

	errorChannel := testView.ErrorViewModel.ErrorChannel
	ExecuteCloneCommand(appCfg, testView)
	// Use appCfg in your test

	expectedError := "Gitlab token env variable GITLAB_TOKEN_ENV_VAR_NOT_SET not set for missing-token.gitlab.com; skipping"
	err := <-errorChannel
	if fmt.Sprintf("%v", err) != expectedError {
		t.Errorf("expecting error: %v, got %v", expectedError, err)
	}

}
