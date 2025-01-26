package ext

import (
	"os"
	"strings"
)

// ReplaceHomeDirWithTilde replaces the home directory in an absolute path with ~
func ReplaceHomeDirWithTilde(path string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return path // If there's an error, return the original path
	}

	if strings.HasPrefix(path, homeDir) {
		return "~" + strings.TrimPrefix(path, homeDir)
	}
	return path
}
