package fs

import (
	"fmt"
	"os"
)

type MkDirFn = func(string) error

func MkDir(projectPath string) error {
	err := os.MkdirAll(projectPath, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create directory %s: %v", projectPath, err)
	}
	return nil
}
