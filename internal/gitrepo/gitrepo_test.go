package gitrepo

import (
	"gcm/internal/sh"
	"testing"
)

func TestClone(t *testing.T) {
	mockRunner := &sh.MockCommandRunner{
		Output: "mock output",
		Err:    nil,
	}

	gr := GitRepository{
		Name:              "gitRepoName",
		SSHURLToRepo:      "git:somewhere.else",
		PathWithNamespace: "",
		Archived:          false,
		CloneOptions:      nil,
	}
	err := gr.Clone(mockRunner)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if mockRunner.ExecutedCommands[0] != "git clone" {
		t.Fatalf("expected 'git clone', got %v", mockRunner.ExecutedCommands[0])
	}
}
