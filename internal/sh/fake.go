package sh

import "gcm/internal/fs"

// MockCommandRunner is a mock implementation for testing
type MockCommandRunner struct {
	Output           string
	Err              error
	ExecutionCwds    []fs.DirectoryPath
	ExecutedCommands []ShellCommand
}

func (m *MockCommandRunner) ExecuteShellCommand(cwd fs.DirectoryPath, command ShellCommand) (string, error) {

	m.ExecutionCwds = append(m.ExecutionCwds, cwd)
	m.ExecutedCommands = append(m.ExecutedCommands, command)
	return m.Output, m.Err
}
