package terminalView

import (
	"gcm/internal/log"
	"gcm/internal/view"
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
		ErrorViewModel:        view.NewErrorViewModel(logger.GetLogFilePath()),
	}
}

func (vm *CloneCommandViewModel) AddGitLabCloneVM(hostName, absPath string) *GitLabCloneViewModel {
	cloneViewModel := NewGitLabCloneViewModel(hostName, absPath)
	vm.GitLabCloneViewModels = append(vm.GitLabCloneViewModels, cloneViewModel)
	return cloneViewModel
}

func (vm *CloneCommandViewModel) getGitLabCloneViewModels() []*GitLabCloneViewModel {
	return vm.GitLabCloneViewModels
}
