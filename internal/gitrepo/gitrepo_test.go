package gitrepo

import (
	"gcm/internal/sh"
	"testing"
)

func TestCloneTableDriven(t *testing.T) {
	tests := []struct {
		name string
		// Arrange
		cloneExists bool
		archived    bool

		// Act
		cwd           sh.DirectoryPath
		cloneArchived bool

		//  Assert
		expectedCmd sh.ShellCommand
		expectedCwd sh.DirectoryPath
		expectError bool
	}{
		{
			name:        "Clone when .git does not exist",
			cloneExists: false,

			cwd:         "/faking/it/somewhere",
			expectedCmd: "git clone git:somewhere.else .",
			expectedCwd: "/faking/it/somewhere",
			expectError: false,
		},
		{
			name:          "Clone when archived",
			cloneExists:   false,
			cloneArchived: true,

			cwd:         "/faking/it/somewhere/else",
			expectedCmd: "git clone git:somewhere.else .",
			expectedCwd: "/faking/it/somewhere/else",
			expectError: false,
			//expectArchiveMarker: true // AM HERE...keep improving those tests
		},
		{
			name:        "Clone when directory exists",
			cloneExists: true,
			expectedCmd: "",
			expectedCwd: "",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				mockRunner := &sh.MockCommandRunner{
					Output: "mock output",
					Err:    nil,
				}
				fakeCloneOptions := MockCloneOptions{cloneArchived: false, rootDirectory: tt.cwd}
				gr := GitRepository{
					Name:              "gitRepoName",
					SSHURLToRepo:      "git:somewhere.else",
					PathWithNamespace: "",
					Archived:          false,
					CloneOptions:      fakeCloneOptions,
					DirectoryExistsCheckFn: func(s string) (bool, error) {
						return tt.cloneExists, nil
					},
					MkDirFn: func(dir string) error {
						if dir != string(tt.cwd) {
							t.Fatalf("Making the wrong dir, expected %s, got %s", tt.cwd, dir)
						}
						return nil
					},
				}
				err := gr.Clone(mockRunner)

				if (err != nil) != tt.expectError {
					t.Fatalf("expected error: %v, got: %v", tt.expectError, err)
				}
				if len(mockRunner.ExecutedCommands) == 0 && tt.expectedCmd != "" {
					t.Fatalf("No clone commands executed")
				}
				if len(mockRunner.ExecutedCommands) > 0 && mockRunner.ExecutedCommands[0] != tt.expectedCmd {
					t.Fatalf("expected '%v', got '%v'", tt.expectedCmd, mockRunner.ExecutedCommands[0])
				}
				if len(mockRunner.ExecutionCwds) > 0 && mockRunner.ExecutionCwds[0] != tt.expectedCwd {
					t.Fatalf("expected '%v', got '%v'", tt.expectedCwd, mockRunner.ExecutionCwds[0])
				}
			},
		)
	}
}
