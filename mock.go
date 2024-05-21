package plug

func Set[F any](f, m F) {
	newScope(1).set(f, m)
}

func Get[F any](f, dflt F) F {
	return newScope(1).get(f, dflt).(F)
}

func CurrentScope() *Scope {
	return newScope(1)
}
