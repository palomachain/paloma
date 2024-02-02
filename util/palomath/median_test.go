package palomath_test

import (
	"testing"

	"github.com/palomachain/paloma/util/palomath"
)

func TestMedianFloats(t *testing.T) {
	type args struct {
		s []float64
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{name: "empty slice", args: args{s: nil}, want: 0},
		{name: "single element", args: args{s: []float64{1}}, want: 1},
		{name: "sorted slice", args: args{s: []float64{2, 3, 4, 5}}, want: 3.5},
		{name: "unsorted slice", args: args{s: []float64{4, 2, 3, 5}}, want: 3.5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := palomath.Median(tt.args.s)
			if got != tt.want {
				t.Errorf("Median(%v) = %v, want %v", tt.args.s, got, tt.want)
			}
		})
	}
}

func TestMedianIntegers(t *testing.T) {
	type args struct {
		s []int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{name: "empty slice", args: args{s: nil}, want: 0},
		{name: "single element", args: args{s: []int{1}}, want: 1},
		{name: "sorted slice", args: args{s: []int{2, 3, 4, 5}}, want: 3},
		{name: "unsorted slice", args: args{s: []int{4, 2, 3, 5}}, want: 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := palomath.Median(tt.args.s)
			if got != tt.want {
				t.Errorf("Median(%v) = %v, want %v", tt.args.s, got, tt.want)
			}
		})
	}
}
