package terminalView

import (
	"bytes"
	"fmt"
	"gcm/internal/counter"
	"testing"
)

func TestErrorView_Render(t *testing.T) {
	// Create a new ErrorViewModel and ErrorView
	vm := &ErrorViewModel{
		errorCount:  counter.NewCounter(),
		latestError: "This is a very long error message that should be truncated",
		logFilePath: "somePath.log",
	}
	vm.errorCount.Add(1)

	var buf bytes.Buffer
	view := NewErrorView(vm, &buf)

	// Render the view with a mock terminal width of 20
	view.Render(20)

	// Expected output
	expectedOutput := "--- 1 errors ---\nSee log file:\nsomePath.log\n"

	// Check the output
	if buf.String() != expectedOutput {
		fmt.Println("Actual:")
		fmt.Println(buf.String())
		t.Errorf("\nexpected %q\n"+
			"     got %q", expectedOutput, buf.String())
	}
}
