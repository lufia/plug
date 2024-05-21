package mock

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
