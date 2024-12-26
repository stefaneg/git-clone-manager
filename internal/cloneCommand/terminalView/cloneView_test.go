package terminalView

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"
	"tools/internal/color"
	"tools/internal/counter"
	"tools/internal/view"
)

type MockTimeElapsedView struct {
	Output string
	stdout io.Writer
}

func NewFakeView(stdout io.Writer, output string) view.View {
	return &MockTimeElapsedView{
		Output: output,
		stdout: stdout,
	}
}

func (m *MockTimeElapsedView) Render() {
	_, err := fmt.Fprint(m.stdout, m.Output)
	if err != nil {
		return
	}
}

func escapeNonPrintable(input string) string {
	// Replace ANSI escape sequences with readable placeholders
	replacer := strings.NewReplacer(
		"\033[4A", "\\033[4A",
		"\033[5A", "\\033[5A",
		"\033", "\\033",
	)
	return replacer.Replace(input)
}

func TestCloneView_Render(t *testing.T) {
	viewModel := NewCloneViewModel()
	addSomeFakeCounts(viewModel)

	var buf bytes.Buffer
	mockTimeElapsedView := NewFakeView(&buf, "5.00 seconds\n")

	cloneView := NewCloneView(viewModel, false, &buf, mockTimeElapsedView)

	// Call RenderNonTTY
	cloneView.RenderNonTTY()

	// Expected output
	expected := fmt.Sprintf("Cloning done\n%s projects in %s groups\n%s git clones (%s archived) \n%s cloned now\n%s seconds\n",
		color.FgMagenta("20"),
		color.FgMagenta("10"),
		color.FgMagenta("30"),
		color.FgMagenta("5"),
		color.FgMagenta("2"),
		color.FgGreen("5.00"))

	// Assert output
	if buf.String() != expected {
		t.Errorf("RenderNonTTY() output mismatch.\nExpected:\n%s\nGot:\n%s", expected, buf.String())
	}
}

func TestCloneView_StartTTYRenderLoop(t *testing.T) {
	// Mock CloneViewModel
	cloneViewModel := &CloneViewModel{
		GroupCount:           counter.NewCounter(),
		ProjectCount:         counter.NewCounter(),
		CloneCount:           counter.NewCounter(),
		ClonedNowCount:       counter.NewCounter(),
		ArchivedCloneCounter: counter.NewCounter(),
	}

	// Use a buffer to capture stdout
	var buf bytes.Buffer
	mockTimeElapsedView := NewFakeView(&buf, "5.00 seconds\n")

	cloneView := NewCloneView(cloneViewModel, true, &buf, mockTimeElapsedView)

	ctx, cancel := context.WithCancel(context.Background())

	addSomeFakeCounts(cloneViewModel)

	go cloneView.StartTTYRenderLoop(ctx)

	// Let the loop run for a short duration
	time.Sleep(1 * time.Millisecond)

	cancel()

	// Expected render calls
	singleRender := fmt.Sprintf("%s projects in %s groups\n%s git clones (%s archived) \n%s cloned now\n%s seconds\n",
		color.FgMagenta("20"),
		color.FgMagenta("10"),
		color.FgMagenta("30"),
		color.FgMagenta("5"),
		color.FgMagenta("2"),
		color.FgGreen("5.00"))

	// Combine expected output, including ANSI escape for moving the cursor up
	expected := singleRender + ansiLineOffset(4) + singleRender

	// Assert output
	if buf.String() != expected {
		t.Errorf("StartTTYRenderLoop() output mismatch.\nExpected:\n%s\nGot:\n%s",
			escapeNonPrintable(expected),
			escapeNonPrintable(buf.String()))
	}
}

func addSomeFakeCounts(mockModel *CloneViewModel) {
	mockModel.ProjectCount.Add(20)
	mockModel.GroupCount.Add(10)
	mockModel.CloneCount.Add(30)
	mockModel.ArchivedCloneCounter.Add(5)
	mockModel.ClonedNowCount.Add(2)
}

func ansiLineOffset(lines int) string {
	return fmt.Sprintf("\033[%dA", lines)
}
