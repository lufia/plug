package plugcore

import (
	"math/rand/v2"
	"reflect"
	"testing"
)

func TestScopeGet(t *testing.T) {
	scope := NewScope(1)
	t.Cleanup(scope.Delete)
	p := reflect.ValueOf(TestScopeGet).Pointer()
	if scope.entry != p {
		t.Errorf("Scope.entry = %v; want %v", scope.entry, p)
	}
}

func TestScopeGeneric(t *testing.T) {
	scope := NewScope(1)
	t.Cleanup(scope.Delete)

	key := Func("math/rand/v2.N", rand.N[int])
	Set(scope, key, func(int) int {
		return 0
	}, nil, nil)
	f := Get(scope, key, func(int) int {
		return 10
	}, nil, nil)
	if v := reflect.ValueOf(f); v.IsZero() {
		t.Errorf("got %v", v)
	}
}
