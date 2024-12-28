package terminalView

import (
	"fmt"
	"gcm/internal/color"
	"gcm/internal/counter"
	"io"
	"os"
	"strings"
)

type ErrorViewModel struct {
	errorCount   *counter.Counter
	latestError  string
	ErrorChannel chan error
	logFilePath  string
}

func NewErrorViewModel(logFilePath string) *ErrorViewModel {
	viewModel := ErrorViewModel{
		errorCount:   counter.NewCounter(),
		ErrorChannel: make(chan error),
		logFilePath:  logFilePath,
	}
	go func() {
		for err := range viewModel.ErrorChannel {
			viewModel.errorCount.Add(1)
			viewModel.latestError = err.Error()
		}
	}()
	return &viewModel
}

type ErrorView struct {
	viewModel *ErrorViewModel
	stdout    io.Writer
}

func NewErrorView(vm *ErrorViewModel, stdout io.Writer) *ErrorView {
	return &ErrorView{
		viewModel: vm,
		stdout:    stdout,
	}
}

func (v ErrorView) Render(width int) int {
	if v.viewModel.errorCount.Count() > 0 {
		out := fmt.Sprintf("--- %s errors ---\nSee log file:\n%s\n",
			color.FgRed(fmt.Sprintf("%d", v.viewModel.errorCount.Count())),
			color.FgMagenta(ReplaceHomeDirWithTilde(v.viewModel.logFilePath)))

		_, err := fmt.Fprint(v.stdout, out)
		if err != nil {
			panic(err)
		}
		return strings.Count(out, "\n")
	} else {
		return 0
	}
}

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

func FitOutputToWidthUsingCut(width int, out string) string {
	// Truncate or pad the string to fit the terminal width
	lines := strings.Split(out, "\n")
	for i, line := range lines {
		if len(line) > width {
			if width > 3 {
				lines[i] = "..." + line[len(line)-width+3:]
			} else {
				lines[i] = line[len(line)-width:]
			}
		} else {
			lines[i] = fmt.Sprintf("%-*s", width, line)
		}
	}
	out = strings.Join(lines, "\n")
	return out
}

func FitOutputToWidth(width int, out string) string {
	// Truncate or pad the string to fit the terminal width
	lines := strings.Split(out, "\n")
	for i, line := range lines {
		if len(line) > width {
			lines[i] = line[:width]
		} else {
			lines[i] = fmt.Sprintf("%-*s", width, line)
		}
	}
	out = strings.Join(lines, "\n")
	return out
}

func WordWrap(text string, width int) string {
	if width <= 0 {
		return text
	}

	var wrapped string
	var line string

	words := strings.Fields(text)
	for _, word := range words {
		if len(line)+len(word)+1 > width {
			if len(line) > 0 {
				wrapped += line + "\n"
			}
			line = word
		} else {
			if len(line) > 0 {
				line += " " + word
			} else {
				line = word
			}
		}
	}

	if len(line) > 0 {
		wrapped += line
	}

	// Handle forced wrapping for long words without spaces
	lines := strings.Split(wrapped, "\n")
	for i, line := range lines {
		if len(line) > width {
			lines[i] = ""
			for len(line) > width {
				lines[i] += line[:width] + "\n"
				line = line[width:]
			}
			lines[i] += line
		}
	}

	return strings.Join(lines, "\n")
}
