package rrange

import (
	"sort"
)

type IntRangeReducer struct{}

func (IntRangeReducer) Merge(i, j int) (int, bool) {
	return 1, true
}
func (IntRangeReducer) Less(i, j int) bool {
	return true
}

var _ RangedEventReducer[int] = IntRangeReducer{}

type RangedEventReducer[T any] interface {
	Merge(a, b T) (T, bool)
	Less(a, b T) bool
}

func SimplifyRangeEvents[T any](events []T, reducer RangedEventReducer[T]) []T {
	slice := events[:]
	sort.Slice(slice, func(i, j int) bool { return reducer.Less(events[i], events[j]) })
	events = slice

	newEvents := []T{}
	newEvents = append(newEvents, events[0])
	for i := 1; i < len(events); i++ {
		evt := newEvents[len(newEvents)-1]

		newEvt, ok := reducer.Merge(evt, events[i])
		if !ok {
			newEvents = append(newEvents, events[i])
			continue
		}

		newEvents[len(newEvents)-1] = newEvt
	}

	return newEvents
}
