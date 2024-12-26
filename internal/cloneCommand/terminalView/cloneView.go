package terminalView

import (
	"context"
	"fmt"
	"io"
	"time"
	"tools/internal/color"
	"tools/internal/counter"
	"tools/internal/view"
)

type CloneViewModel struct {
	GroupCount           *counter.Counter
	ProjectCount         *counter.Counter
	CloneCount           *counter.Counter
	ClonedNowCount       *counter.Counter
	ArchivedCloneCounter *counter.Counter
}

func NewCloneViewModel() *CloneViewModel {
	return &CloneViewModel{
		ClonedNowCount:       counter.NewCounter(),
		ProjectCount:         counter.NewCounter(),
		CloneCount:           counter.NewCounter(),
		GroupCount:           counter.NewCounter(),
		ArchivedCloneCounter: counter.NewCounter(),
	}
}

// CloneView handles rendering counters in different modes
type CloneView struct {
	viewModel       *CloneViewModel
	isTTY           bool
	stdout          io.Writer
	timeElapsedView view.View
}

func NewCloneView(store *CloneViewModel, isTTY bool, stdout io.Writer, timeElapsedView view.View) *CloneView {
	return &CloneView{
		viewModel:       store,
		isTTY:           isTTY,
		stdout:          stdout,
		timeElapsedView: timeElapsedView,
	}
}

func (r *CloneView) StartTTYRenderLoop(ctx context.Context) {
	// Initial placeholder rendering to create space for counters
	r.render()

	for {
		select {
		case <-ctx.Done():
			return // Exit the render loop when the context is canceled
		default:
			_, err := fmt.Fprintf(r.stdout, "\033[%dA", 4)
			if err != nil {
				return
			}
			r.render()
			time.Sleep(100 * time.Millisecond) // Refresh rate
		}
	}
}
func (r *CloneView) render() {
	_, err := fmt.Fprintf(r.stdout, "%s projects in %s groups\n%s git clones (%s archived) \n%s cloned now\n",
		color.FgMagenta(fmt.Sprintf("%d", r.viewModel.ProjectCount.Count())),
		color.FgMagenta(fmt.Sprintf("%d", r.viewModel.GroupCount.Count())),
		color.FgMagenta(fmt.Sprintf("%d", r.viewModel.CloneCount.Count())),
		color.FgMagenta(fmt.Sprintf("%d", r.viewModel.ArchivedCloneCounter.Count())),
		color.FgMagenta(fmt.Sprintf("%d", r.viewModel.ClonedNowCount.Count())))
	if err != nil {
		return
	}
	r.timeElapsedView.Render()
}

func (r *CloneView) RenderNonTTY() {
	_, err := fmt.Fprintln(r.stdout, "Cloning done")
	if err != nil {
		return
	}
	r.render()
}
