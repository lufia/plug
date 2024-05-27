package plug

import (
	"reflect"
	"testing"
)

func TestScopeGet(t *testing.T) {
	scope := CurrentScope()
	p := reflect.ValueOf(TestScopeGet).Pointer()
	if scope.entry != p {
		t.Errorf("Scope.entry = %v; want %v", scope.entry, p)
	}
}

func put[T any](v T) {}

func _put[T any](v T) {}

func TestScopeGeneric(t *testing.T) {
	scope := CurrentScope()
	defer scope.Delete()

	key := Func("put", put[int])
	Set(scope, key, func(int) {})
	if scope.mocks[key.key] == nil {
		t.Errorf("scope[%v] = nil; but want non-nil", key)
	}
	f := Get(scope, key, _put[int])
	if v := reflect.ValueOf(f); v.IsZero() {
		t.Errorf("got %v", v)
	}
}
