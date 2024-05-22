package plug

func Set[F any](f, m F) {
	newScope(1).set(f, m)
}

func SetT1[F, T1 any](f, m F, _ T1) {
	newScope(1).set(f, m)
}

func Get[F any](f, dflt F) F {
	return newScope(1).get(f, dflt).(F)
}

func GetT1[F, T1 any](f, dflt F, _ T1) F {
	return newScope(1).get(f, dflt).(F)
}

func CurrentScope() *Scope {
	return newScope(1)
}
