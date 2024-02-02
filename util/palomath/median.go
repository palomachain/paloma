package palomath

import (
	"slices"

	"golang.org/x/exp/constraints"
)

type Number interface {
	constraints.Float | constraints.Integer
}

// Median returns the median value of the given slice.
// The slice does not need to be sorted.
func Median[E Number](s []E) E {
	if len(s) < 1 {
		return 0
	}

	w := make([]E, len(s))
	copy(w, s)

	slices.Sort(w)

	c := len(w) / 2
	if len(w)%2 == 0 {
		return (w[c-1] + w[c]) / 2.0
	} else {
		return w[c]
	}
}
