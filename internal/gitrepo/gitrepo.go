package gitrepo

import (
	"fmt"
	"gcm/internal/fs"
	"gcm/internal/gitremote"
	. "gcm/internal/log"
	"gcm/internal/sh"
	"path"
)

type GitRepository struct {
	Name                   string
	SSHURLToRepo           string
	PathWithNamespace      string
	Archived               bool
	CloneOptions           CloneOptions
	DirectoryExistsCheckFn fs.DirectoryExistsCheckFn
	MkDirFn                fs.MkDirFn
	CreateSmallTextFileFn  fs.CreateSmallTextFileFn
}

func NewGitRepositoryFromRemoteConfig(
	project gitremote.ProjectConfig,
	hostName string,
	cloneDirectory string,
) *GitRepository {
	opts := RemoteCloneOptions{cloneDirectory: cloneDirectory}

	name := project.Name
	fullPath := project.FullPath
	sprintf := fmt.Sprintf("git@%s:%s", hostName, fullPath)
	return NewGitRepository(name, fullPath, sprintf, opts)
}

func NewGitRepository(name string, fullPath string, sprintf string, opts RemoteCloneOptions) *GitRepository {
	var gitRepo = GitRepository{
		Name:                   name,
		PathWithNamespace:      fullPath,
		SSHURLToRepo:           sprintf,
		CloneOptions:           opts,
		DirectoryExistsCheckFn: fs.DirectoryExists,
		MkDirFn:                fs.MkDir,
		CreateSmallTextFileFn:  fs.CreateSmallTextFile,
	}
	return &gitRepo
}

func (repo *GitRepository) GetName() string {
	return repo.Name
}

func (repo *GitRepository) IsArchived() bool {
	return repo.Archived
}

func (repo *GitRepository) GetCloneOptions() CloneOptions {
	return repo.CloneOptions
}

func (repo *GitRepository) Clone(cmdRunner sh.CommandRunner) error {
	needsCloning, checkErr := repo.CheckNeedsCloning()
	if !needsCloning {
		return checkErr
	}

	projectPath := fs.DirectoryPath(repo.getWorkingCopyPath(repo.CloneOptions.CloneRootDirectory()))
	Log.Infof("Cloning %s to %s", repo.Name, projectPath)
	err := repo.MkDirFn(projectPath)
	if err != nil {
		return err
	}

	cloneCmd := fmt.Sprintf("git clone %s .", repo.SSHURLToRepo)
	_, err = cmdRunner.ExecuteShellCommand(fs.DirectoryPath(projectPath), sh.ShellCommand(cloneCmd))

	if err != nil {
		return fmt.Errorf("in %s, %s failed: %s", projectPath, cloneCmd, err)
	}

	if repo.Archived {
		err := repo.WriteArchivedMarker(projectPath)
		if err != nil {
			return err
		}
	}

	return nil
}

func (repo *GitRepository) CheckNeedsCloning() (bool, error) {
	cloned, err := repo.IsCloned()
	if err != nil {
		return false, err
	}
	if cloned {
		return false, nil
	}
	if !repo.cloneArchived() && repo.Archived {
		return false, nil
	}
	return true, nil
}

func (repo *GitRepository) IsCloned() (bool, error) {
	projectPath := repo.getWorkingCopyPath(repo.CloneOptions.CloneRootDirectory())
	return repo.DirectoryExistsCheckFn(fs.DirectoryPath(path.Join(projectPath, ".git")))
}

func (repo *GitRepository) getWorkingCopyPath(cloneDirectory string) string {
	return path.Join(cloneDirectory, repo.PathWithNamespace)
}

// WriteArchivedMarker creates an "ARCHIVED.txt" file in the root directory of the archived project
func (repo *GitRepository) WriteArchivedMarker(projectPath fs.DirectoryPath) error {
	// Define the path for the ARCHIVED.txt marker file
	fileName := fs.FileName("ARCHIVED.txt")
	fileContent := "This repo is archived and not active.\n"
	return repo.CreateSmallTextFileFn(projectPath, fileName, fileContent)
}

func (repo *GitRepository) cloneArchived() bool {
	return repo.CloneOptions.CloneArchived()
}

type RemoteCloneOptions struct {
	cloneDirectory string
}

func (rco RemoteCloneOptions) CloneRootDirectory() string {
	return rco.cloneDirectory
}

func (_ RemoteCloneOptions) CloneArchived() bool {
	return true
}
