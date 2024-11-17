package sh

import (
	"os"
	"os/exec"
	"strings"
)

type DirectoryPath string
type ShellCommand string

func ExecuteShellCommand(cwd DirectoryPath, command ShellCommand) (string, error) {
	cmd := exec.Command("sh", "-c", string(command))
	cmd.Dir = string(cwd)
	cmd.Env = os.Environ()
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
