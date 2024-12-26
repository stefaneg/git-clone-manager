package terminalView

import (
	"fmt"
	"gcm/internal/color"
	"gcm/internal/counter"
	"io"
	"strings"
)

type ClonedNowViewModel struct {
	ClonedNowCount *counter.Counter
}

func NewClonedNowViewModel() *ClonedNowViewModel {
	return &ClonedNowViewModel{
		ClonedNowCount: counter.NewCounter(),
	}
}

type ClonedNowView struct {
	viewModel *ClonedNowViewModel
	stdout    io.Writer
}

func NewClonedNowView(vm *ClonedNowViewModel, stdout io.Writer) *ClonedNowView {
	return &ClonedNowView{
		viewModel: vm,
		stdout:    stdout,
	}
}

func (v ClonedNowView) Render() int {
	out := fmt.Sprintf("%s cloned now\n", color.FgMagenta(fmt.Sprintf("%d", v.viewModel.ClonedNowCount.Count())))
	_, err := fmt.Fprint(v.stdout, out)
	if err != nil {
		return 0
	}
	return strings.Count(out, "\n")
}
