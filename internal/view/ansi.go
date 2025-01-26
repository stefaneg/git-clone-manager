package view

import "fmt"

func ansiLineOffset(lines int) string {
	return fmt.Sprintf("\033[%dA", lines)
}
