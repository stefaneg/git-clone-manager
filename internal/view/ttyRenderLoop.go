package view

import (
	"context"
	"fmt"
	"golang.org/x/term"
	"io"
	"os"
	"time"
)

func StartTTYRenderLoop(r View, out io.Writer, ctx context.Context) {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		panic(err)
	}
	lineCount := r.Render(width)

	for {
		width, _, err := term.GetSize(int(os.Stdout.Fd()))
		if err != nil {
			panic(err)
		}
		select {
		case <-ctx.Done():
			return // Exit the Render loop when the context is canceled
		default:
			_, err := fmt.Fprint(out, ansiLineOffset(lineCount))
			if err != nil {
				return
			}
			lineCount = r.Render(width)
			time.Sleep(100 * time.Millisecond) // Refresh rate
		}
	}
}