package gitrepo

import (
	"gcm/internal/fs"
	"gcm/internal/sh"
)

type FakeGitRepository struct {
	GitRepository
	FakeFS *fs.FakeFileSystem
}

func (fgr *FakeGitRepository) Clone(_ sh.CommandRunner) error {
	// Use the fake filesystem to simulate cloning behavior
	projectPath := fs.DirectoryPath(fgr.getWorkingCopyPath(fgr.CloneOptions.CloneRootDirectory()))
	if _, err := fgr.FakeFS.DirectoryExists(fs.DirectoryPath(projectPath)); err == nil {
		return nil // Simulate already cloned
	}

	// Simulate creating the directory and ARCHIVED.txt if needed
	if err := fgr.FakeFS.MkDir(projectPath); err != nil {
		return err
	}
	if fgr.Archived {
		return fgr.WriteArchivedMarker(fs.DirectoryPath(projectPath))
	}
	return nil
}
