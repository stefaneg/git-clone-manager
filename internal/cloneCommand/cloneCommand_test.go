package cloneCommand

import (
	"fmt"
	"gcm/internal/appConfig"
	"gcm/internal/cloneCommand/terminalView"
	"gcm/internal/counter"
	"gcm/internal/fs"
	"gcm/internal/gitlab"
	"gcm/internal/gitremote"
	"gcm/internal/sh"
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
		}, &sh.FakeCommandRunner{}, // Empty string indicates env variable not set
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

// TODO GSE: AM HERE, next is designing the arrangement nicely for next command
// Perhaps add tests for interesting corner cases.
func TestCloneCommandWithFakeImplementationsTableDriven(t *testing.T) {
	type testCase struct {
		name                 string
		appCfg               *appConfig.AppConfig
		gitlabApiArrangement gitlab.API

		expectedDir                fs.DirectoryPath
		expectedApis               []TokenHostPair
		expectedCommands           []string
		expectedCwds               []fs.DirectoryPath
		expectedCloneCount         int
		expectedError              string
		expectedEnvVariableLookups []string
	}

	testCases := []testCase{
		{
			name: "Single project clone",
			appCfg: &appConfig.AppConfig{
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
			},
			gitlabApiArrangement:       gitlab.FakeSetupNoSubgroupsOneProject(),
			expectedDir:                fs.DirectoryPath("/path/for/gitlab.somehost.com"),
			expectedApis:               []TokenHostPair{{Token: "FAKE_TOKEN", Host: "gitlab.somehost.com"}},
			expectedCommands:           []string{"git clone git@somewhere:project1.git ."},
			expectedCwds:               []fs.DirectoryPath{fs.DirectoryPath("/path/for/gitlab.somehost.com")},
			expectedCloneCount:         1,
			expectedEnvVariableLookups: []string{"GITLAB_TEST_TOKEN_ENV_VAR"},
		},
	}

	for _, tc := range testCases {
		t.Run(
			tc.name, func(t *testing.T) {
				// Setup fake dependencies
				fakeFS := &fs.FakeFileSystem{}
				createdApis := []TokenHostPair{}

				fakeAPIFactory := func(token string, host string) gitlab.API {
					createdApis = append(createdApis, TokenHostPair{token, host})
					apiArrangement := tc.gitlabApiArrangement
					return apiArrangement // Needs to go into test arrangement
				}
				testView := NewFakeCloneCommandViewModel()

				envVariableLookups := []string{}
				fakeGetEnv := func(s string) string {
					envVariableLookups = append(envVariableLookups, s)
					return "FAKE_TOKEN"
				}
				mockCommandRunner := sh.FakeCommandRunner{}
				cmd := NewCloneCommand(
					testView, fakeFS, fakeAPIFactory, fakeGetEnv, &mockCommandRunner,
				)

				wg := sync.WaitGroup{}
				wg.Add(1)
				go func() {
					defer wg.Done()
					cmd.Execute(tc.appCfg)
				}()
				wg.Wait()

				if len(tc.expectedEnvVariableLookups) != len(envVariableLookups) {
					t.Errorf(
						"expected env variable lookups %v, got %v",
						tc.expectedEnvVariableLookups,
						envVariableLookups,
					)
				} else {
					for i, expectedEnvVariableLookup := range tc.expectedEnvVariableLookups {
						if envVariableLookups[i] != expectedEnvVariableLookup {
							t.Errorf(
								"expected env variable lookup %v, got %v",
								expectedEnvVariableLookup,
								envVariableLookups[i],
							)
						}
					}
				}
				// Assertions
				if len(fakeFS.CreatedDirs) < 1 || fakeFS.CreatedDirs[0] != tc.expectedDir {
					t.Errorf("expected directory %v to be created, but got %v", tc.expectedDir, fakeFS.CreatedDirs[0])
				}
				if len(createdApis) != len(tc.expectedApis) {
					t.Errorf("expected %v created APIs, found %v", len(tc.expectedApis), len(createdApis))
				}
				for i, expectedApi := range tc.expectedApis {
					if createdApis[i] != expectedApi {
						t.Errorf("expected API %v, got %v", expectedApi, createdApis[i])
					}
				}
				if testView.ClonedNowViewModel.ClonedNowCount.Count() != tc.expectedCloneCount {
					t.Errorf(
						"expected %v projects to be cloned now, got %v",
						tc.expectedCloneCount,
						testView.ClonedNowViewModel.ClonedNowCount.Count(),
					)
				}
				if len(mockCommandRunner.ExecutedCommands) != len(tc.expectedCommands) {
					t.Errorf(
						"expected %v commands to be executed, got %v",
						len(tc.expectedCommands),
						len(mockCommandRunner.ExecutedCommands),
					)
				}
				for i, expectedCommand := range tc.expectedCommands {
					if mockCommandRunner.ExecutedCommands[i] != sh.ShellCommand(expectedCommand) {
						t.Errorf(
							"expected command %v to be executed, got %v",
							expectedCommand,
							mockCommandRunner.ExecutedCommands[i],
						)
					}
				}
				for i, expectedCwd := range tc.expectedCwds {
					if mockCommandRunner.ExecutionCwds[i] != expectedCwd {
						t.Errorf(
							"expected current working dir %v, got %v",
							expectedCwd,
							mockCommandRunner.ExecutionCwds[i],
						)
					}
				}
			},
		)
	}
}
