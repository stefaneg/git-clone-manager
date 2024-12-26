package view

import (
	"fmt"
	"io"
	"time"
	"tools/internal/color"
)

type TimeElapsedView struct {
	startTime time.Time
	stdout    io.Writer
	since     func(time.Time) time.Duration // Custom Since function
}

func NewTimeElapsedView(startTime time.Time, stdout io.Writer, since func(time.Time) time.Duration) *TimeElapsedView {
	return &TimeElapsedView{
		startTime: startTime,
		stdout:    stdout,
		since:     since,
	}
}

func (t *TimeElapsedView) Render() {
	elapsed := t.since(t.startTime).Seconds()
	_, err := fmt.Fprintf(t.stdout, "%s seconds\n", color.FgGreen(fmt.Sprintf("%.2f", elapsed)))
	if err != nil {
		return
	}
}
