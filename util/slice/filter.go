package slice

func Filter[V any](slice []V, filterIn func(val V) bool) []V {
	var filtered []V
	for _, item := range slice {
		if filterIn(item) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}
