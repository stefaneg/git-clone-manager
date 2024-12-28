package terminalView

import (
	"bytes"
	"fmt"
	"gcm/internal/color"
	"gcm/internal/view"
	"io"
	"strings"
	"testing"
)

type FakeView struct {
	Output string
	stdout io.Writer
}

func NewFakeView(stdout io.Writer, output string) view.View {
	return &FakeView{
		Output: output,
		stdout: stdout,
	}
}

func (m *FakeView) Render(int) int {
	var lines int
	_, err := fmt.Fprint(m.stdout, m.Output)
	if err != nil {
		return lines
	}
	return strings.Count(m.Output, "\n")
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
	viewModel := NewCloneViewModel("testing.123", "localtest")
	addSomeFakeCounts(viewModel)

	var buf bytes.Buffer
	cloneView := NewCloneView(viewModel, false, &buf)

	// Call RenderNonTTY
	lineCount := cloneView.Render(11)

	// Expected output
	expected := fmt.Sprintf("localtest  \n  <- testi:\n    %s projects in %s groups\n    %s direct projects\n    %s git clones (%s archived)\n",
		color.FgMagenta("20"),
		color.FgMagenta("10"),
		color.FgMagenta("1"),
		color.FgMagenta("30"),
		color.FgMagenta("5"))

	// Assert output
	if buf.String() != expected {
		t.Errorf("Render() output mismatch.\nExpected:\n%s\nGot:\n%s", escapeNonPrintable(expected), escapeNonPrintable(buf.String()))
	}
	if lineCount != 5 {
		t.Errorf("Render() line count.\nExpected: %d\nGot: %d", 5, lineCount)
	}
}

func addSomeFakeCounts(mockModel *CloneViewModel) {
	mockModel.GroupProjectCount.Add(20)
	mockModel.GroupCount.Add(10)
	mockModel.DirectProjectCount.Add(1)
	mockModel.CloneCount.Add(30)
	mockModel.ArchivedCloneCounter.Add(5)
}
