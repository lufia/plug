package mock

func Set[F any](f, m F) {
	NewScope(1).Set(f, m)
}

func Get[F any](dflt F) F {
	return NewScope(1).Get(dflt).(F)
}

func Cleanup() {
	NewScope(1).Delete()
}
