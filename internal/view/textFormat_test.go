package view

import (
	"testing"
)

func TestTruncateTextToWidth(t *testing.T) {
	tests := []struct {
		name     string
		width    int
		input    string
		expected string
	}{
		{
			name:     "Short text, no truncation",
			width:    10,
			input:    "short",
			expected: "short     ",
		},
		{
			name:     "Exact width text",
			width:    5,
			input:    "exact",
			expected: "exact",
		},
		{
			name:     "Long text, truncation with ellipsis",
			width:    10,
			input:    "This is a very long text",
			expected: "...ng text",
		},
		{
			name:     "Width less than 3, no ellipsis",
			width:    2,
			input:    "long text",
			expected: "xt",
		},
		{
			name:     "Multiple lines, mixed lengths",
			width:    10,
			input:    "short\nThis is a very long text\nexact",
			expected: "short     \n...ng text\nexact     ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TruncateTextToWidth(tt.width, tt.input)
			if result != tt.expected {
				t.Errorf("Test expected %q, got %q", tt.expected, result)
			}
		})
	}
}
