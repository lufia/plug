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
	"github.com/lufia/plug/plugcore"
)

// Symbol represents an object that will be replaced.
type Symbol[T any] plugcore.Symbol[T]

func (s *Symbol[T]) core() *plugcore.Symbol[T] {
	return (*plugcore.Symbol[T])(s)
}

// Func returns a symbol constructed with both the name and the function is referenced to.
// The name syntax must be either $package.$function or $package.$type.$method.
//
// For example:
//
//   - math/rand/v2.N
//   - net/http.Client.Do
func Func[F any](name string, f F) *Symbol[F] {
	return (*Symbol[F])(plugcore.Func(name, f))
}

type constraints struct {
	recv   any
	params plugcore.Params
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

	return (*Object)(plugcore.Set(scope.core(), s.core(), v, c.recv, c.params))
}

// Get returns an object that is bound to s, or dflt if s is bound nothing.
func Get[T any](scope *Scope, s *Symbol[T], dflt T, opts ...Option) T {
	var c constraints
	applyOptions(&c, opts...)
	return plugcore.Get(scope.core(), s.core(), dflt, c.recv, c.params)
}

// CurrentScope returns the scope object that is strongly related to current calling stacks on the goroutine.
//
// When the scope becomes unnecessary the scope should be released through [Scope.Delete] method.
// Otherwise the scope and its objects will not be garbage collection because the package continues to kept them in the internal state.
func CurrentScope() *Scope {
	return (*Scope)(plugcore.NewScope(1))
}

type testingTB interface {
	Cleanup(func())
}

// CurrentScopeFor is similar to [CurrentScope] except the scope will be automatically deleted on cleanup of t.
func CurrentScopeFor(t testingTB) *Scope {
	scope := plugcore.NewScope(1)
	t.Cleanup(scope.Delete)
	return (*Scope)(scope)
}
