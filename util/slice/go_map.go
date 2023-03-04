package slice

import (
	"fmt"
	"sort"

	"golang.org/x/exp/constraints"
)

func FromMapValues[K constraints.Ordered, V any](mm map[K]V) []V {
	res := make([]V, 0, len(mm))
	for _, v := range FromMapKeys(mm) {
		res = append(res, mm[v])
	}
	return res
}

func FromMapKeys[K constraints.Ordered, V any](mm map[K]V) []K {
	res := make([]K, 0, len(mm))
	for k := range mm {
		res = append(res, k)
	}
	sort.Slice(res, func(i, j int) bool { return res[i] < res[j] })
	return res
}

// MakeMapKeys makes a map of provided slice and a function which
// returns a key value for a map given an item from a slice.
// If key already exists, it overrides it.
func MakeMapKeys[K comparable, V any](slice []V, getKey func(V) K) map[K]V {
	m := make(map[K]V, len(slice))
	for _, item := range slice {
		key := getKey(item)
		m[key] = item
	}
	return m
}

// MustMakeMapKeys makes a map of provided slice and a function which
// returns a key value for a map given an item from a slice.
// If key already exists, it panics.
func MustMakeMapKeys[K comparable, V any](slice []V, getKey func(V) K) map[K]V {
	m := make(map[K]V, len(slice))
	for _, item := range slice {
		key := getKey(item)
		if _, ok := m[key]; ok {
			panic(fmt.Sprintf("key %s already exists", any(key)))
		}
		m[key] = item
	}
	return m
}
