package main

func Map[S ~[]T, T, R any](a S, f func(v T) R) []R {
	p := make([]R, len(a))
	for i, v := range a {
		p[i] = f(v)
	}
	return p
}

func MapValues[M ~map[K]V, K comparable, V any](m M) []V {
	a := make([]V, 0, len(m))
	for _, v := range m {
		a = append(a, v)
	}
	return a
}
