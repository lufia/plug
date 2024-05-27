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
	"testing"
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

// Option represents the constraints for [Get] or [Set].
type Option func(*constraints)

func applyOptions(c *constraints, opts ...Option) {
	for _, opt := range opts {
		opt(c)
	}
}

func WithParams(params map[string]any) Option {
	return func(c *constraints) {
		c.params = params
	}
}

// Set binds s to v. If s does already bound to another object, it will unbound.
func Set[T any](scope *Scope, s *Symbol[T], v T, opts ...Option) *Object {
	var c constraints
	applyOptions(&c, opts...)
	// TODO(lufia): scope.set will become to receive c.

	return scope.set(s.key, v)
}

// Get returns an object that is bound to s, or dflt if s is bound nothing.
func Get[T any](scope *Scope, s *Symbol[T], dflt T, opts ...Option) T {
	var c constraints
	applyOptions(&c, opts...)
	return scope.get(s.key, dflt, &c).(T)
}

// CurrentScope returns the scope object that is strongly related to current calling stacks on the goroutine.
//
// When the scope becomes unnecessary the scope should be released through [Scope.Delete] method.
// Otherwise the scope and its objects will not be garbage collection because the package continues to kept them in the internal state.
func CurrentScope() *Scope {
	return newScope(1)
}

// CurrentScopeFor is similar to [CurrentScope] except the scope will be automatically deleted on cleanup of t.
func CurrentScopeFor(t testing.TB) *Scope {
	scope := newScope(1)
	t.Cleanup(scope.Delete)
	return scope
}
