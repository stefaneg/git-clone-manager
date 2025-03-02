package fs

import (
	"fmt"
	"gcm/internal/log"
	"github.com/sirupsen/logrus"
	"os"
	"path"
)

func DirectoryExists(directoryName DirectoryPath) (bool, error) {
	gitDir, err := os.Stat(string(directoryName))
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return gitDir.IsDir(), nil
}

func MkDir(projectPath DirectoryPath) error {
	err := os.MkdirAll(string(projectPath), os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create directory %s: %v", projectPath, err)
	}
	return nil
}

// CreateSmallTextFile creates a small text file with the specified content.
// FilePath: the directory path where the file will be created.
// FileName: the name of the file to be created.
// fileContent: the content to be written to the file.
// Returns an error if the file creation or writing fails.
func CreateSmallTextFile(filePath DirectoryPath, fileName string, fileContent string) error {
	markerFilePath := path.Join(string(filePath), fileName)
	// Create the marker file
	file, err := os.Create(markerFilePath)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			// To publish to errorChannel or not...that is the question.
			logger.Log.Errorf("failed to close file: %v", err)
		}
	}(file)

	_, err = file.WriteString(fileContent)
	if err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}
	if logger.Log.GetLevel() >= logrus.DebugLevel {
		logger.Log.Debugf("file created at %s\n", markerFilePath)
	}
	return nil
}
