package sh

import (
	"gcm/internal/fs"
	"os"
	"os/exec"
	"strings"
)

type ShellCommand string

type CommandRunner interface {
	ExecuteShellCommand(cwd fs.DirectoryPath, command ShellCommand) (string, error)
}

type ShellCommandRunner struct{}

func (r *ShellCommandRunner) ExecuteShellCommand(cwd fs.DirectoryPath, command ShellCommand) (string, error) {
	cmd := exec.Command("sh", "-c", string(command))
	cmd.Dir = string(cwd)
	cmd.Env = os.Environ()
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
