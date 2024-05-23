package plug

import "reflect"

type symbolKey struct {
	name string
	t    reflect.Type
}

type Symbol[T any] symbolKey

func (s *Symbol[T]) key() symbolKey {
	return symbolKey(*s)
}

func Func[F any](name string, f F) *Symbol[F] {
	var zero F
	return &Symbol[F]{name, reflect.TypeOf(zero)}
}

func Set[T any](s *Symbol[T], v T) {
	newScope(1).set(s.key(), v)
}

func Get[T any](s *Symbol[T], dflt T) T {
	return newScope(1).get(s.key(), dflt).(T)
}

func CurrentScope() *Scope {
	return newScope(1)
}
