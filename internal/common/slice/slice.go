package slice

func FindFirst[A any](as []A, f func(A) bool) *A {
	for _, e := range as {
		if f(e) {
			return &e
		}
	}

	return nil
}

func Map[A, B any](as []A, fn func(A) B) []B {
	bs := make([]B, len(as))

	for i, e := range as {
		bs[i] = fn(e)
	}

	return bs
}
