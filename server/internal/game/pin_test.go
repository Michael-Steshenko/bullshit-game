package game

import (
	"testing"
)

func TestIntToPin(t *testing.T) {
	tests := []struct {
		expected string
		input    int64
	}{
		{"AAAA", 0},
		{"AAAB", 1},
		{"AAAZ", 25},
		{"AABA", 26},
		{"AAZZ", 675},
		{"ABAA", 676},
	}
	for _, tt := range tests {
		got := intToPin(tt.input)
		if got != tt.expected {
			t.Errorf("intToPin(%d) = %s, want %s", tt.input, got, tt.expected)
		}
	}
}

func TestPinGeneratorSequential(t *testing.T) {
	gen := NewPinGenerator()
	first := gen.Next()
	second := gen.Next()

	if first != "AAAA" {
		t.Errorf("first pin = %s, want AAAA", first)
	}
	if second != "AAAB" {
		t.Errorf("second pin = %s, want AAAB", second)
	}
}

func TestPinGeneratorSetCounter(t *testing.T) {
	gen := NewPinGenerator()
	gen.SetCounter(26)
	pin := gen.Next()
	if pin != "AABA" {
		t.Errorf("pin = %s, want AABA", pin)
	}
}
