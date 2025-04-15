package view

import (
	"fmt"
	"gcm/internal/appConfig"
	"gcm/internal/color"
	"gcm/internal/counter"
	"gcm/internal/fs"
	logger "gcm/internal/log"
	"io"
	"strings"
)

type ErrorViewModel struct {
	ErrorChannel chan error
	ErrorCount   *counter.Counter
	LatestError  string
	logFilePath  fs.FilePath
}

func NewErrorViewModel(logFilePath fs.FilePath) *ErrorViewModel {
	viewModel := ErrorViewModel{
		ErrorCount:   counter.NewCounter(),
		ErrorChannel: make(chan error, appConfig.DefaultChannelBufferLength),
		logFilePath:  logFilePath,
	}
	go func() {
		for err := range viewModel.ErrorChannel {
			viewModel.ErrorCount.Add(1)
			viewModel.LatestError = err.Error()
			logger.Log.Errorf("%v", err)
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
	if v.viewModel.ErrorCount.Count() > 0 {
		out := fmt.Sprintf(
			("--- %s errors ---\nSee log file:\n%s\n"),
			color.FgRed(fmt.Sprintf("%d", v.viewModel.ErrorCount.Count())),
			color.FgMagenta(fs.ReplaceHomeDirWithTilde(fs.Path(v.viewModel.logFilePath))),
		)

		_, err := fmt.Fprint(v.stdout, out)
		if err != nil {
			panic(err)
		}
		return strings.Count(out, "\n")
	} else {
		return 0
	}
}
