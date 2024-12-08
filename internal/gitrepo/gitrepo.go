package gitrepo

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"path"
	"tools/internal/color"
	"tools/internal/gitremote"
	. "tools/internal/log"
	"tools/internal/sh"
)

type Repository struct {
	Name              string
	SSHURLToRepo      string
	PathWithNamespace string
	Archived          bool
	CloneOptions      CloneOptions
}

type CloneOptions interface {
	CloneArchived() bool
	CloneRootDirectory() string
}

func (repo *Repository) Clone() error {
	needsCloning, checkErr := repo.CheckNeedsCloning()
	if !needsCloning {
		return checkErr
	}

	projectPath := repo.getWorkingCopyPath(repo.CloneOptions.CloneRootDirectory())
	Log.Infof("Cloning %s to %s", color.FgMagenta(repo.Name), color.FgMagenta(projectPath))
	err := os.MkdirAll(projectPath, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create directory %s: %v", color.FgRed(projectPath), err)
	}
	cloneCmd := fmt.Sprintf("git clone %s .", repo.SSHURLToRepo)
	_, err = sh.ExecuteShellCommand(sh.DirectoryPath(projectPath), sh.ShellCommand(cloneCmd))

	if err != nil {
		return fmt.Errorf("in %s, %s failed: %s", color.FgRed(projectPath), cloneCmd, err)
	}

	if repo.Archived {
		err := repo.WriteArchivedMarker(projectPath)
		if err != nil {
			return err
		}
	}

	return nil
}

func (repo *Repository) CheckNeedsCloning() (bool, error) {
	projectPath := repo.getWorkingCopyPath(repo.CloneOptions.CloneRootDirectory())

	if _, err := os.Stat(path.Join(projectPath, ".git")); !os.IsNotExist(err) {
		if Log.GetLevel() >= logrus.DebugLevel {
			Log.Debugf("Git repository %s already exists at %s, skipping clone\n", color.FgMagenta(repo.Name), color.FgMagenta(projectPath))
		}
		return false, nil // Skip cloning if directory already exists
	}
	if !repo.cloneArchived() && repo.Archived {
		if Log.GetLevel() >= logrus.DebugLevel {
			Log.Debugf("Skipping archived repo %s %s", color.FgMagenta(repo.Name), color.FgMagenta(projectPath))
		}
		return false, nil
	}
	return true, nil
}

func (repo *Repository) getWorkingCopyPath(cloneDirectory string) string {
	projectPath := path.Join(cloneDirectory, repo.PathWithNamespace)
	return projectPath
}

// WriteArchivedMarker creates an "ARCHIVED.txt" file in the root directory of the archived project
func (repo *Repository) WriteArchivedMarker(projectPath string) error {
	// Define the path for the ARCHIVED.txt marker file
	markerFilePath := path.Join(projectPath, "ARCHIVED.txt")

	// Create the marker file
	file, err := os.Create(markerFilePath)
	if err != nil {
		Log.Errorf("failed to create marker file: %w", err)
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			Log.Errorf("failed to close marker file: %w", err)
		}
	}(file)

	// Write a message indicating the repo is archived
	_, err = file.WriteString("This repo is archived and not active.\n")
	if err != nil {
		return fmt.Errorf("failed to write to marker file: %w", err)
	}
	if Log.GetLevel() >= logrus.DebugLevel {
		Log.Debugf("ARCHIVED.txt marker file created at %s\n", color.FgCyan(markerFilePath))
	}
	return nil
}

func (repo *Repository) cloneArchived() bool {
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

func CreateFromGitRemoteConfig(project gitremote.GitRemoteProjectConfig, hostName string, cloneDirectory string) *Repository {
	opts := RemoteCloneOptions{cloneDirectory: cloneDirectory}

	var gitRepo = Repository{
		Name:              project.Name,
		PathWithNamespace: project.FullPath,
		SSHURLToRepo:      fmt.Sprintf("git@%s:%s", hostName, project.FullPath),
		CloneOptions:      opts,
	}
	return &gitRepo
}
