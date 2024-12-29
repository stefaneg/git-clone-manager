package view

import (
	"testing"
)

type MockView struct {
	lines int
}

func (mv *MockView) Render(int) int {
	return mv.lines
}

func TestCompositeView_Render(t *testing.T) {
	view1 := &MockView{lines: 3}
	view2 := &MockView{lines: 5}
	view3 := &MockView{lines: 2}

	compositeView := NewCompositeView([]View{view1, view2, view3})

	totalLines := compositeView.Render(80)
	expectedLines := 10

	if totalLines != expectedLines {
		t.Errorf("expected %d lines, got %d", expectedLines, totalLines)
	}
}
