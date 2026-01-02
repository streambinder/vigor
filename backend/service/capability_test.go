package service

import "testing"

func TestProgressiveMargin(t *testing.T) {
	tests := []struct {
		completed int
		expected  float64
	}{
		{0, 45.0},
		{1, 35.0},
		{2, 35.0},
		{3, 25.0},
		{4, 25.0},
		{5, 15.0},
		{10, 15.0},
		{100, 15.0},
	}

	for _, tt := range tests {
		got := ProgressiveMargin(tt.completed)
		if got != tt.expected {
			t.Errorf("ProgressiveMargin(%d) = %v, want %v", tt.completed, got, tt.expected)
		}
	}
}
