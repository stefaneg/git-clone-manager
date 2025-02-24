package fs

import "os"

type DirectoryExistsCheckFn func(string) (bool, error)

func DirectoryExists(directoryName string) (bool, error) {
	gitDir, err := os.Stat(directoryName)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return gitDir.IsDir(), nil
}
