package gitrepo

import (
	"fmt"
	"gcm/internal/gitremote"
	. "gcm/internal/log"
	"gcm/internal/sh"
	"github.com/sirupsen/logrus"
	"os"
	"path"
)

type GitRepository struct {
	Name              string
	SSHURLToRepo      string
	PathWithNamespace string
	Archived          bool
	CloneOptions      CloneOptions
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

func (repo *GitRepository) Clone() error {
	needsCloning, checkErr := repo.CheckNeedsCloning()
	if !needsCloning {
		return checkErr
	}

	projectPath := repo.getWorkingCopyPath(repo.CloneOptions.CloneRootDirectory())
	Log.Infof("Cloning %s to %s", repo.Name, projectPath)
	err := os.MkdirAll(projectPath, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create directory %s: %v", projectPath, err)
	}
	cloneCmd := fmt.Sprintf("git clone %s .", repo.SSHURLToRepo)
	_, err = sh.ExecuteShellCommand(sh.DirectoryPath(projectPath), sh.ShellCommand(cloneCmd))

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
	gitDir, err := os.Stat(path.Join(projectPath, ".git"))
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return gitDir.IsDir(), nil
}

func (repo *GitRepository) getWorkingCopyPath(cloneDirectory string) string {
	projectPath := path.Join(cloneDirectory, repo.PathWithNamespace)
	return projectPath
}

// WriteArchivedMarker creates an "ARCHIVED.txt" file in the root directory of the archived project
func (repo *GitRepository) WriteArchivedMarker(projectPath string) error {
	// Define the path for the ARCHIVED.txt marker file
	markerFilePath := path.Join(projectPath, "ARCHIVED.txt")

	// Create the marker file
	file, err := os.Create(markerFilePath)
	if err != nil {
		// To publish to errorChannel or not...that is the question.
		Log.Errorf("failed to create marker file: %v", err)
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			// To publish to errorChannel or not...that is the question.
			Log.Errorf("failed to close marker file: %v", err)
		}
	}(file)

	// Write a message indicating the repo is archived
	_, err = file.WriteString("This repo is archived and not active.\n")
	if err != nil {
		return fmt.Errorf("failed to write to marker file: %w", err)
	}
	if Log.GetLevel() >= logrus.DebugLevel {
		Log.Debugf("ARCHIVED.txt marker file created at %s\n", markerFilePath)
	}
	return nil
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

func CreateFromGitRemoteConfig(
	project gitremote.GitRemoteProjectConfig,
	hostName string,
	cloneDirectory string,
) *GitRepository {
	opts := RemoteCloneOptions{cloneDirectory: cloneDirectory}

	var gitRepo = GitRepository{
		Name:              project.Name,
		PathWithNamespace: project.FullPath,
		SSHURLToRepo:      fmt.Sprintf("git@%s:%s", hostName, project.FullPath),
		CloneOptions:      opts,
	}
	return &gitRepo
}
