package mock

import (
	"reflect"
	"testing"
)

func TestScopeGet(t *testing.T) {
	s := NewScope(1)
	p := reflect.ValueOf(TestScopeGet).Pointer()
	if s.entry != p {
		t.Errorf("Scope.entry = %v; want %v", s.entry, p)
	}
}
