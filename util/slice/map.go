package slice

func Map[A any, B any](in []A, f func(A) B) []B {
	res := make([]B, 0, len(in))
	for _, el := range in {
		res = append(res, f(el))
	}
	return res
}

func MapErr[A any, B any](in []A, f func(A) (B, error)) ([]B, error) {
	res := make([]B, 0, len(in))
	for _, el := range in {
		newVal, err := f(el)
		if err != nil {
			return nil, err
		}
		res = append(res, newVal)
	}
	return res, nil
}
