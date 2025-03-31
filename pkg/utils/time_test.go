package utils

import "testing"

func TestParseZfsTimeUnit(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"500us", 500},
		{"1ms", 1000},
		{"2s", 2000000},
		{"2.5s", 2500000},
		{"100", 100},
		{"-", 0},
		{"bad", 0},
		{"123xyz", 123},
		{"3.75ms", 3750},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ParseZfsTimeUnit(tt.input)
			if result != tt.expected {
				t.Errorf("ParseZfsTimeUnit(%q) = %d; want %d", tt.input, result, tt.expected)
			}
		})
	}
}
