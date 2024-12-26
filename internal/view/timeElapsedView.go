package view

import (
	"fmt"
	"gcm/internal/color"
	"io"
	"strings"
	"time"
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

func (t *TimeElapsedView) Render() int {
	elapsed := t.since(t.startTime).Seconds()
	out := fmt.Sprintf("%s seconds\n", color.FgGreen(fmt.Sprintf("%.2f", elapsed)))
	_, err := fmt.Fprint(t.stdout, out)
	if err != nil {
		return 0
	}
	return strings.Count(out, "\n")
}
