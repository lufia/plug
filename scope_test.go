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

	entry := reflect.ValueOf(put[int]).Pointer()
	mock := func(int){}
	p := reflect.ValueOf(mock)
	Set(put[int], mock)
	if scope.mocks[entry] == nil {
		t.Errorf("scope[%v] = nil; but want %v", entry, p)
	}
	f := Get(put[int], _put[int])
	if v := reflect.ValueOf(f); v.Pointer() != p.Pointer() {
		t.Errorf("got %v; want %v", v, p)
	}
}
