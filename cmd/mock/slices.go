package main

func Map[S ~[]T, T, R any](a S, f func(v T) R) []R {
	p := make([]R, len(a))
	for i, v := range a {
		p[i] = f(v)
	}
	return p
}
