package plug

import (
	"math/rand/v2"
	"reflect"
	"testing"
)

func TestScopeGet(t *testing.T) {
	scope := CurrentScopeFor(t)
	p := reflect.ValueOf(TestScopeGet).Pointer()
	if scope.entry != p {
		t.Errorf("Scope.entry = %v; want %v", scope.entry, p)
	}
}

func TestScopeGeneric(t *testing.T) {
	scope := CurrentScopeFor(t)

	key := Func("math/rand/v2.N", rand.N[int])
	Set(scope, key, func(int) int {
		return 0
	})
	f := Get(scope, key, func(int) int {
		return 10
	})
	if v := reflect.ValueOf(f); v.IsZero() {
		t.Errorf("got %v", v)
	}
}
