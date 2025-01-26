package view

import (
	"fmt"
	"strings"
)

// TruncateTextToWidth Cuts off front of text and adds ellipsis to indicate that text was shortened. Fills lines with spaces.
func TruncateTextToWidth(width int, out string) string {
	lines := strings.Split(out, "\n")
	for i, line := range lines {
		if len(line) > width {
			if width > 3 {
				lines[i] = "..." + line[len(line)-width+3:]
			} else {
				lines[i] = line[len(line)-width:]
			}
		} else {
			lines[i] = fmt.Sprintf("%-*s", width, line)
		}
	}
	out = strings.Join(lines, "\n")
	return out
}

// TrimTextToWidth Cuts off end of every line if longer than width. Fills lines to width with spaces.
func TrimTextToWidth(width int, out string) string {
	lines := strings.Split(out, "\n")
	for i, line := range lines {
		if len(line) > width {
			lines[i] = line[:width]
		} else {
			lines[i] = fmt.Sprintf("%-*s", width, line)
		}
	}
	out = strings.Join(lines, "\n")
	return out
}
