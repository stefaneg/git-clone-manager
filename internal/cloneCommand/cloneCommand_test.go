package cloneCommand

import (
	"fmt"
	"gcm/internal/appConfig"
	"gcm/internal/cloneCommand/terminalView"
	"gcm/internal/counter"
	"gcm/internal/fs"
	"gcm/internal/gitlab"
	"gcm/internal/gitremote"
	"gcm/internal/view"
	"math"
	"sync"
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
	fakeFs := &fs.FakeFileSystem{}
	cmd := NewCloneCommand(
		testView, fakeFs, gitlab.FakeApiFactory, func(s string) string {
			return ""
		}, // Empty string indicates env variable not set
	)
	cmd.Execute(appCfg)
	// Use appCfg in your test

	expectedError := "Gitlab token env variable GITLAB_TOKEN_ENV_VAR_NOT_SET not set for missing-token.gitlab.com; skipping"
	err := <-errorChannel
	if fmt.Sprintf("%v", err) != expectedError {
		t.Errorf("expecting error: %v, got %v", expectedError, err)
	}

}

type TokenHostPair struct {
	Token string
	Host  string
}

func TestCloneCommandWithFakeImplementations(t *testing.T) {
	// Setup fake dependencies
	fakeFS := &fs.FakeFileSystem{}
	createdApis := []TokenHostPair{}

	fakeAPIFactory := func(token string, host string) gitlab.API {
		createdApis = append(createdApis, TokenHostPair{token, host})
		return gitlab.FakeSetupNoSubgroupsOneProject()
	}
	testView := NewFakeCloneCommandViewModel()

	// Create a sample app config
	appCfg := &appConfig.AppConfig{
		GitLab: []gitlab.GitLabConfig{
			{
				EnvTokenVariableName: "GITLAB_TEST_TOKEN_ENV_VAR",
				HostName:             "gitlab.somehost.com",
				CloneDirectory:       "/path/for/gitlab.somehost.com",
				Groups: []gitlab.GroupConfig{
					{
						Name:          "root1",
						CloneArchived: false,
					},
				},
				Projects:           []gitremote.ProjectConfig{},
				RateLimitPerSecond: math.MaxInt,
			},
		},
	}

	envVariableName := "NONE"
	fakeGetEnv := func(s string) string {
		envVariableName = s
		return "FAKE_TOKEN"
	}
	cmd := NewCloneCommand(
		testView, fakeFS, fakeAPIFactory, fakeGetEnv,
	)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		cmd.Execute(appCfg)
	}()
	wg.Wait()

	expectedDir := fs.DirectoryPath("/path/for/gitlab.somehost.com")
	if len(fakeFS.CreatedDirs) < 1 || fakeFS.CreatedDirs[0] != expectedDir {
		t.Errorf("expected directory %v to be created, but got %v", expectedDir, fakeFS.CreatedDirs[0])
	}
	if len(createdApis) != 1 {
		t.Errorf("expected 1 created api, found %v", len(createdApis))
	}
	if envVariableName != appCfg.GitLab[0].EnvTokenVariableName {
		t.Errorf(
			"Expected env variable %v to be used to retrieve token env variable %v",
			envVariableName,
			appCfg.GitLab[0].EnvTokenVariableName,
		)
	}
	if createdApis[0].Token != fakeGetEnv("") {
		t.Errorf(
			"Expected retrieved env token %v to be used to create API %v",
			fakeGetEnv(""),
			createdApis[0].Token,
		)
	}
	if createdApis[0].Host != appCfg.GitLab[0].HostName {
		t.Errorf("Expected host to be used to create API")
	}
	if testView.ClonedNowViewModel.ClonedNowCount.Count() != 1 {
		t.Errorf("Expected one project to be cloned now, got %v", testView.ClonedNowViewModel.ClonedNowCount.Count())
	}
}
