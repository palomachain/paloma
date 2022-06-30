package slice

func All[A any, B any](in []A, f func(A)) {
	for _, el := range in {
		f(el)
	}
}
