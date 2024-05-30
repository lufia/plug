package plug

import (
	"math/rand/v2"
	"reflect"
	"testing"
)

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
