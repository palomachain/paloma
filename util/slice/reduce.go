package slice

func Reduce[R any, V any](slice []V, reducer func(prev R, val V) (next R)) R {
	var reduced R
	for _, item := range slice {
		reduced = reducer(reduced, item)
	}
	return reduced
}
