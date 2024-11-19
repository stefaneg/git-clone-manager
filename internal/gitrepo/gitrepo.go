package gitrepo

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"path"
	"tools/internal/color"
	l "tools/internal/log"
	"tools/internal/sh"
)

type GitRepoSpec struct {
	Name              string `json:"name"`
	SSHURLToRepo      string `json:"ssh_url_to_repo"`
	PathWithNamespace string `json:"path_with_namespace"`
	Archived          bool   `json:"archived"`
}

func (project *GitRepoSpec) CloneProject(cloneDirectory string, cloneArchived bool) error {
	projectPath := project.getProjectPath(cloneDirectory)

	// Check if the project directory already exists
	if _, err := os.Stat(projectPath); !os.IsNotExist(err) {
		if l.Log.GetLevel() >= logrus.DebugLevel {
			l.Log.Debugf("GitRepoSpec %s already exists at %s, skipping clone\n", project.Name, projectPath)
		}
		return nil // Skip cloning if directory already exists
	}
	if !cloneArchived && project.Archived {
		fmt.Printf("Skipping archived project %s %s\n", project.Name, projectPath)
		return nil
	}

	l.Log.Infof("Cloning project to %s\n", projectPath)
	err := os.MkdirAll(projectPath, 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory %s: %v", projectPath, err)
	}
	output, err := sh.ExecuteShellCommand(sh.DirectoryPath(projectPath), sh.ShellCommand(fmt.Sprintf("git clone %s", project.SSHURLToRepo)))

	if err != nil {
		return fmt.Errorf("git clone failed: %s", output)
	}

	if project.Archived {
		err := project.WriteArchivedMarker(projectPath)
		if err != nil {
			return err
		}
	}

	return nil
}

func (project *GitRepoSpec) getProjectPath(cloneDirectory string) string {
	projectPath := path.Join(cloneDirectory, project.PathWithNamespace)
	return projectPath
}

// WriteArchivedMarker creates an "ARCHIVED.txt" file in the root directory of the archived project
func (project *GitRepoSpec) WriteArchivedMarker(projectPath string) error {
	// Define the path for the ARCHIVED.txt marker file
	markerFilePath := path.Join(projectPath, "ARCHIVED.txt")

	// Create the marker file
	file, err := os.Create(markerFilePath)
	if err != nil {
		l.Log.Errorf("failed to create marker file: %w", err)
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			l.Log.Errorf("failed to close marker file: %w", err)
		}
	}(file)

	// Write a message indicating the project is archived
	_, err = file.WriteString("This project is archived and not active.\n")
	if err != nil {
		return fmt.Errorf("failed to write to marker file: %w", err)
	}
	if l.Log.GetLevel() >= logrus.DebugLevel {
		l.Log.Debugf("ARCHIVED.txt marker file created at %s\n", color.FgCyan(markerFilePath))
	}
	return nil
}
