// Package plugcore is designed to be imported only from the plug package.
// All packages except plug or its artifacts must not import this package directly.
package plugcore

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

type Params = map[string]any

// Set binds s to v. If s does already bound to another object, it will unbound.
func Set[T any](scope *Scope, s *Symbol[T], v T, recv any, params Params) *Object {
	return scope.set(s.key, v)
}

// Get returns an object that is bound to s, or dflt if s is bound nothing.
func Get[T any](scope *Scope, s *Symbol[T], dflt T, recv any, params Params) T {
	return scope.get(s.key, dflt, recv, params).(T)
}
