package palomath

// Clamp returns n clamped to [min, max]
func Clamp[E Number](n, min, max E) E {
	switch true {
	case n < min:
		return min
	case n > max:
		return max
	default:
		return n
	}
}
