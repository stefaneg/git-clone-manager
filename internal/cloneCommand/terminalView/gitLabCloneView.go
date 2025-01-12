package terminalView

import (
	"fmt"
	"gcm/internal/color"
	"gcm/internal/counter"
	"gcm/internal/ext"
	"gcm/internal/view"
	"io"
	"strings"
)

type GitLabCloneViewModel struct {
	CloneRoot            string
	RemoteHostName       string
	GroupCount           *counter.Counter
	GroupProjectCount    *counter.Counter
	DirectProjectCount   *counter.Counter
	CloneCount           *counter.Counter
	ArchivedCloneCounter *counter.Counter
}

func NewGitLabCloneViewModel(remoteHostName string, cloneRoot string) *GitLabCloneViewModel {
	return &GitLabCloneViewModel{
		CloneRoot:            cloneRoot,
		RemoteHostName:       remoteHostName,
		GroupProjectCount:    counter.NewCounter(),
		DirectProjectCount:   counter.NewCounter(),
		CloneCount:           counter.NewCounter(),
		GroupCount:           counter.NewCounter(),
		ArchivedCloneCounter: counter.NewCounter(),
	}
}

// GitLabCloneView handles rendering counters in different modes
type GitLabCloneView struct {
	viewModels []*GitLabCloneViewModel
	stdout     io.Writer
}

func NewGitLabCloneView(stdout io.Writer) *GitLabCloneView {
	return &GitLabCloneView{
		viewModels: []*GitLabCloneViewModel{},
		stdout:     stdout,
	}
}

func (r *GitLabCloneView) AddViewModel(viewModel *GitLabCloneViewModel) {
	r.viewModels = append(r.viewModels, viewModel)
}

func (r *GitLabCloneView) Render(width int) (lines int) {
	var out strings.Builder
	for _, vm := range r.viewModels {
		out.WriteString(
			fmt.Sprintf(
				"%s\n  <- %s:\n    %s projects in %s groups\n    %s direct projects\n    %s git clones (%s archived)\n",
				color.FgCyan(view.TruncateTextToWidth(width, ext.ReplaceHomeDirWithTilde(vm.CloneRoot))),
				color.FgCyan(view.TrimTextToWidth(ext.Max(width-6, 1), vm.RemoteHostName)),
				color.FgMagenta(fmt.Sprintf("%d", vm.GroupProjectCount.Count())),
				color.FgMagenta(fmt.Sprintf("%d", vm.GroupCount.Count())),
				color.FgMagenta(fmt.Sprintf("%d", vm.DirectProjectCount.Count())),
				color.FgMagenta(fmt.Sprintf("%d", vm.CloneCount.Count())),
				color.FgMagenta(fmt.Sprintf("%d", vm.ArchivedCloneCounter.Count())),
			),
		)
	}
	_, err := fmt.Fprint(r.stdout, out.String())
	if err != nil {
		return 0
	}
	return strings.Count(out.String(), "\n")
}
