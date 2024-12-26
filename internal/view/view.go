package view

import (
	"context"
	"fmt"
	"io"
	"time"
)

type View interface {
	Render() (lines int)
}

func StartTTYRenderLoop(r View, out io.Writer, ctx context.Context) {
	// Initial placeholder rendering to create space for counters
	lineCount := r.Render()

	for {
		select {
		case <-ctx.Done():
			return // Exit the Render loop when the context is canceled
		default:
			_, err := fmt.Fprint(out, ansiLineOffset(lineCount))
			if err != nil {
				return
			}
			r.Render()
			time.Sleep(100 * time.Millisecond) // Refresh rate
		}
	}
}

type CompositeView struct {
	views []View
}

func NewCompositeView(views []View) *CompositeView {
	return &CompositeView{views: views}
}

func (cv *CompositeView) Render() int {
	totalLines := 0
	for _, view := range cv.views {
		lines := view.Render()
		totalLines += lines
	}
	return totalLines
}

func ansiLineOffset(lines int) string {
	return fmt.Sprintf("\033[%dA", lines)
}
