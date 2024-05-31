package plugcore

import (
	"reflect"
	"runtime"
	"testing"
)

func newScopeFor(t testing.TB) *Scope {
	scope := NewScope(1)
	t.Cleanup(scope.Delete)
	return scope
}

func newScopeParentFor(t testing.TB, parent *Scope) *Scope {
	scope := NewScopeFrom(parent, 1)
	t.Cleanup(scope.Delete)
	return scope
}

// NewScope should return a scope related to current function stack.
func TestNewScope_relatedToCurrentFunc(t *testing.T) {
	f := reflect.ValueOf(TestNewScope_relatedToCurrentFunc)
	name := runtime.FuncForPC(f.Pointer()).Name()

	scope := newScopeFor(t)
	if scope.name != name {
		t.Errorf("Scope.name = %s; want %s", scope.name, name)
	}
	if scope.entry != f.Pointer() {
		t.Errorf("Scope.entry = %v; want %v", scope.entry, f.Pointer())
	}
}

// NewScope should return the same scope if it is called in the same function stack.
func TestNewScope_sameScopeIfSameStack(t *testing.T) {
	scope1 := newScopeFor(t)
	scope2 := newScopeFor(t)
	if scope1 != scope2 {
		t.Errorf("scope1(%p) != scope2(%p)", scope1, scope2)
	}
}

func TestScopeDelete_reachedToZeroRef(t *testing.T) {
	scope1 := NewScope(0)
	scope2 := NewScope(0)
	parent := scope1.parent
	scope1.Delete()
	if _, ok := parent.refers[scope1.entry]; !ok {
		t.Errorf("Delete decreases the ref but it should not delete the scope here")
	}
	scope2.Delete()
	if _, ok := parent.refers[scope2.entry]; ok {
		t.Errorf("Delete should delete the scope")
	}
}

func TestScopeDelete_doubleDelete(t *testing.T) {
	defer func() {
		e := recover()
		if e == nil {
			t.Errorf("Delete should panic if it called two or more times")
		}
		if s := e.(string); s != doubleDeleteMessage {
			t.Errorf("Delete panics with %s; want %s", s, doubleDeleteMessage)
		}
	}()
	scope := NewScope(0)
	scope.Delete()
	scope.Delete()
}

func TestNewScopeFrom(t *testing.T) {
	parent := newScopeFor(t)
	t.Run("child", func(t *testing.T) {
		scope := NewScopeFrom(parent, 0)
		scope.Delete()
	})
}
