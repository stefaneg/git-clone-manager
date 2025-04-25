package gitrepo

import (
	"gcm/internal/fs"
	"gcm/internal/sh"
	"testing"
)

func TestCloneTableDriven(t *testing.T) {
	tests := []struct {
		name string
		// Arrange
		archived bool

		// Act
		cwd           fs.DirectoryPath
		cloneArchived bool

		//  Assert
		expectedCmd         sh.ShellCommand
		expectedCwd         fs.DirectoryPath
		expectError         bool
		expectArchiveMarker bool
	}{
		{
			name: "Clone when .git does not exist",

			cwd:         "/faking/it/somewhere",
			expectedCmd: "git clone git:somewhere.else .",
			expectedCwd: "/faking/it/somewhere",
			expectError: false,
		},
		{
			name:          "Clone when archived",
			cloneArchived: true,
			archived:      true,

			cwd:                 "/faking/it/somewhere/else",
			expectedCmd:         "git clone git:somewhere.else .",
			expectedCwd:         "/faking/it/somewhere/else",
			expectError:         false,
			expectArchiveMarker: true,
		},
		{
			name:                "Clone when directory exists",
			cwd:                 "/this/one/exists",
			expectedCmd:         "",
			expectedCwd:         "",
			expectError:         false,
			expectArchiveMarker: false,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				mockRunner := &sh.FakeCommandRunner{
					Output: "mock output",
					Err:    nil,
				}
				fakeCloneOptions := MockCloneOptions{cloneArchived: tt.cloneArchived, rootDirectory: tt.cwd}
				ffs := &fs.FakeFileSystem{
					ExistingDirs: map[fs.DirectoryPath]bool{
						"/this/one/exists/.git": true,
					},
				}

				gr := GitRepository{
					Name:              "gitRepoName",
					SSHURLToRepo:      "git:somewhere.else",
					PathWithNamespace: "",
					Archived:          tt.archived,
					CloneOptions:      fakeCloneOptions,
					Fs:                ffs,
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
				if tt.expectArchiveMarker && len(ffs.CreatedFiles) == 0 {
					t.Fatalf("expected a archive marker file to be created")
				}
				if tt.expectArchiveMarker && len(ffs.CreatedFiles) > 0 {
					if ffs.CreatedFiles[0].FilePath != tt.cwd {
						t.Fatalf("Expected file to be created in %v, got %v", tt.cwd, ffs.CreatedFiles[0].FilePath)
					}
					if ffs.CreatedFiles[0].FileName != "ARCHIVED.txt" {
						t.Fatalf("Expected filename to be %v, got %v", "ARCHIVED.txt", ffs.CreatedFiles[0].FileName)
					}
				}
			},
		)
	}
}
