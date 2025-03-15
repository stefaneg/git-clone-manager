package ext

import (
	"testing"
)

func TestDefaultValue(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		fallback interface{}
		expected interface{}
	}{
		{"Non-zero int", 5, 10, 5},
		{"Zero int", 0, 10, 10},
		{"Non-zero string", "hello", "world", "hello"},
		{"Empty string", "", "world", "world"},
		{"Non-zero float", 3.14, 2.71, 3.14},
		{"Zero float", 0.0, 2.71, 2.71},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				result := DefaultValue(tt.value, tt.fallback)
				if result != tt.expected {
					t.Errorf("DefaultValue(%v, %v) = %v; want %v", tt.value, tt.fallback, result, tt.expected)
				}
			},
		)
	}
}
