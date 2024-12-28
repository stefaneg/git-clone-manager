package terminalView

import (
	"fmt"
	"gcm/internal/color"
	"gcm/internal/counter"
	"gcm/internal/ext"
	"io"
	"strings"
)

type CloneViewModel struct {
	CloneRoot            string
	RemoteHostName       string
	GroupCount           *counter.Counter
	GroupProjectCount    *counter.Counter
	DirectProjectCount   *counter.Counter
	CloneCount           *counter.Counter
	ArchivedCloneCounter *counter.Counter
}

func NewCloneViewModel(remoteHostName string, cloneRoot string) *CloneViewModel {
	return &CloneViewModel{
		CloneRoot:            cloneRoot,
		RemoteHostName:       remoteHostName,
		GroupProjectCount:    counter.NewCounter(),
		DirectProjectCount:   counter.NewCounter(),
		CloneCount:           counter.NewCounter(),
		GroupCount:           counter.NewCounter(),
		ArchivedCloneCounter: counter.NewCounter(),
	}
}

// CloneView handles rendering counters in different modes
type CloneView struct {
	viewModel *CloneViewModel
	isTTY     bool
	stdout    io.Writer
}

func NewCloneView(store *CloneViewModel, isTTY bool, stdout io.Writer) *CloneView {
	return &CloneView{
		viewModel: store,
		isTTY:     isTTY,
		stdout:    stdout,
	}
}

func (r *CloneView) Render(width int) int {
	out := fmt.Sprintf("%s\n  <- %s:\n    %s projects in %s groups\n    %s direct projects\n    %s git clones (%s archived)\n",
		color.FgCyan(FitOutputToWidthUsingCut(width, ReplaceHomeDirWithTilde(r.viewModel.CloneRoot))),
		color.FgCyan(FitOutputToWidth(ext.Max(width-6, 1), r.viewModel.RemoteHostName)),
		color.FgMagenta(fmt.Sprintf("%d", r.viewModel.GroupProjectCount.Count())),
		color.FgMagenta(fmt.Sprintf("%d", r.viewModel.GroupCount.Count())),
		color.FgMagenta(fmt.Sprintf("%d", r.viewModel.DirectProjectCount.Count())),
		color.FgMagenta(fmt.Sprintf("%d", r.viewModel.CloneCount.Count())),
		color.FgMagenta(fmt.Sprintf("%d", r.viewModel.ArchivedCloneCounter.Count())),
	)
	_, err := fmt.Fprint(r.stdout, out)
	if err != nil {
		return 0
	}
	return strings.Count(out, "\n")
}
