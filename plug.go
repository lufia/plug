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
	"unicode"
)

type Recorder interface {
	record(params map[string]any)
}

type nullRecorder struct{}

func (nullRecorder) record(params map[string]any) {}

type symbolKey struct {
	name string
	t    reflect.Type
}

// Symbol represents an object that will be replaced.
type Symbol[T any] struct {
	key symbolKey
}

// FuncRecorder records its function callings for later inspection in tests.
type FuncRecorder[T any] struct {
	calls []T
}

func (r *FuncRecorder[T]) Count() int {
	return len(r.calls)
}

func (r *FuncRecorder[T]) At(i int) T {
	return r.calls[i]
}

func (r *FuncRecorder[T]) record(params map[string]any) {
	var call T
	v := reflect.ValueOf(&call)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	for _, f := range reflect.VisibleFields(v.Type()) {
		tag := plugTag(f)
		if tag == "-" {
			continue
		}
		param, ok := params[tag]
		if !ok {
			continue
		}
		p := v.FieldByIndex(f.Index)
		p.Set(reflect.ValueOf(param))
	}
	r.calls = append(r.calls, call)
}

func plugTag(f reflect.StructField) string {
	tag := f.Tag.Get("plug")
	if tag == "" && len(f.Name) > 0 {
		s := []rune(f.Name)
		tag = string(unicode.ToLower(s[0])) + string(s[1:])
	}
	if tag == "" || tag == "-" {
		return "-"
	}
	return tag
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
