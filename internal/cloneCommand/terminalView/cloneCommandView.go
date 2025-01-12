package terminalView

import (
	logger "gcm/internal/log"
	"gcm/internal/view"
	"os"
	"time"
)

type CloneCommandViewModel struct {
	GitLabCloneViewModels []*GitLabCloneViewModel
	ClonedNowViewModel    *ClonedNowViewModel
	ErrorViewModel        *view.ErrorViewModel
}

func NewCloneCommandViewModel() *CloneCommandViewModel {
	return &CloneCommandViewModel{
		GitLabCloneViewModels: make([]*GitLabCloneViewModel, 0),
		ClonedNowViewModel:    NewClonedNowViewModel(),
		ErrorViewModel:        view.NewErrorViewModel(logger.GetLogFilePath()), // TODO GetLogFilePath is not good here
	}
}

func (vm *CloneCommandViewModel) AddGitLabCloneVM(model *GitLabCloneViewModel) {
	vm.GitLabCloneViewModels = append(vm.GitLabCloneViewModels, model)
}

func (vm *CloneCommandViewModel) getGitLabCloneViewModels() []*GitLabCloneViewModel {
	return vm.GitLabCloneViewModels
}

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
