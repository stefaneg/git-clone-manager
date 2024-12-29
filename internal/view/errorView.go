package view

import (
	"fmt"
	"gcm/internal/color"
	"gcm/internal/counter"
	"gcm/internal/ext"
	"io"
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

func (v ErrorView) Render(int) int {
	if v.viewModel.errorCount.Count() > 0 {
		out := fmt.Sprintf(("--- %s errors ---\nSee log file:\n%s\n"),
			color.FgRed(fmt.Sprintf("%d", v.viewModel.errorCount.Count())),
			color.FgMagenta(ext.ReplaceHomeDirWithTilde(v.viewModel.logFilePath)))

		_, err := fmt.Fprint(v.stdout, out)
		if err != nil {
			panic(err)
		}
		return strings.Count(out, "\n")
	} else {
		return 0
	}
}
