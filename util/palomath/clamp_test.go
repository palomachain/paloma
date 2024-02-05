package palomath_test

import (
	"testing"

	"github.com/palomachain/paloma/util/palomath"
)

func TestClampInt(t *testing.T) {
	tests := []struct {
		n    int
		min  int
		max  int
		want int
	}{
		{-2, 0, 1, 0},
		{0, -1, 2, 0},
		{1, -2, 3, 1},
		{10, 5, 15, 10},
	}

	for _, tt := range tests {
		got := palomath.Clamp[int](tt.n, tt.min, tt.max)
		if got != tt.want {
			t.Errorf("Clamp(n=%d, min=%d, max=%d) = %d, want %d", tt.n, tt.min, tt.max, got, tt.want)
		}
	}
}

func TestClampFloat64(t *testing.T) {
	tests := []struct {
		n    float64
		min  float64
		max  float64
		want float64
	}{
		{-2.5, 0, 1, 0},
		{0, -1, 2, 0},
		{1.5, -2, 3, 1.5},
		{10.2, 5, 15, 10.2},
	}

	for _, tt := range tests {
		got := palomath.Clamp[float64](tt.n, tt.min, tt.max)
		if got != tt.want {
			t.Errorf("Clamp(n=%f, min=%f, max=%f) = %f, want %f", tt.n, tt.min, tt.max, got, tt.want)
		}
	}
}
