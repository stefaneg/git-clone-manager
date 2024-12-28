package view

import (
	"fmt"
)

type View interface {
	Render(width int) (lines int)
}

type CompositeView struct {
	views []View
}

func NewCompositeView(views []View) *CompositeView {
	return &CompositeView{views: views}
}

func (cv *CompositeView) Render(w int) int {
	totalLines := 0
	for _, view := range cv.views {
		lines := view.Render(w)
		totalLines += lines
	}
	return totalLines
}

func ansiLineOffset(lines int) string {
	return fmt.Sprintf("\033[%dA", lines)
}
