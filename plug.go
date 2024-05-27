// Package plug replaces functions, defined in the other packages, on testing.
//
// # Scope
//
// TBD.
//
// # Generics
//
// TBD.
package plug

import (
	"reflect"
)

type symbolKey struct {
	name string
	t    reflect.Type
}

// Symbol represents an object that will be replaced.
type Symbol[T any] struct {
	key symbolKey
}

// Func returns a symbol constructed with both the name and the function is referenced to.
// The name syntax must be either $package.$function or $package.$type.$method.
//
// For example:
//
//   - math/rand/v2.N
//   - net/http.Client.Do
func Func[F any](name string, f F) *Symbol[F] {
	key := symbolKey{name, reflect.TypeOf(f)}
	return &Symbol[F]{key}
}

// Set binds s to v. If s does already bound to another object, it will unbound.
func Set[T any](s *Symbol[T], v T) *Object {
	return newScope(1).set(s.key, v)
}

// Get returns an object that is bound to s, or dflt if s is bound nothing.
func Get[T any](s *Symbol[T], dflt T, recv any, params map[string]any) T {
	return newScope(1).get(s.key, dflt, recv, params).(T)
}

// CurrentScope returns the scope object that is strongly related to current calling stacks on the goroutine.
//
// When the scope becomes unnecessary the scope should be released through its Delete method.
// Otherwise the scope and its objects will not be garbage collection because the package continues to kept them in the internal state.
func CurrentScope() *Scope {
	return newScope(1)
}
