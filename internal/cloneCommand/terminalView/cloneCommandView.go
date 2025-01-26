package terminalView

import (
	"gcm/internal/view"
	"os"
	"time"
)

type CloneCommandView struct {
	compositeView *view.CompositeView
}

func NewCloneCommandView(vm *CloneCommandViewModel) *CloneCommandView {
	startTime := time.Now()
	out := os.Stdout
	gitLabCloneView := NewGitLabCloneView(out, vm.getGitLabCloneViewModels)

	compositeView := view.NewCompositeView(make([]view.View, 0))
	compositeView.AddView(gitLabCloneView)

	compositeView.AddFooter(view.NewErrorView(vm.ErrorViewModel, out))
	compositeView.AddFooter(NewClonedNowView(vm.ClonedNowViewModel, out))
	compositeView.AddFooter(view.NewTimeElapsedView(startTime, out, time.Since))

	return &CloneCommandView{
		compositeView: compositeView,
	}
}

func (c CloneCommandView) Render(width int) (lines int) {
	return c.compositeView.Render(width)
}
