package mock

func Set[F any](f, m F) {
	NewScope(1).set(f, m)
}

func Get[F any](f, dflt F) F {
	return NewScope(1).get(f, dflt).(F)
}

func Cleanup() {
	NewScope(1).Delete()
}
