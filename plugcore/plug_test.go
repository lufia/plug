package plugcore

import (
	"math/rand/v2"
	"testing"
)

func TestWorkflow_combinedFromAnotherScope(t *testing.T) {
	parent := newScopeFor(t)
	key := Func("dummy", func() {})
	var r FuncRecorder[struct{}]
	Set(parent, key, func() {}, nil, nil).SetRecorder(&r)

	// spawn a goroutine
	t.Run("get function after set", func(t *testing.T) {
		scope := newScopeParentFor(t, parent)
		key := Func("dummy", func() {})
		f := Get(scope, key, func() {
			t.Errorf("should not call fallback function")
		}, nil, nil)
		f()
	})

	if n := r.Count(); n != 1 {
		t.Errorf("Recorder.Count = %d; want 1", n)
	}
}

func TestWorkflow_defaultFunction(t *testing.T) {
	scope := newScopeFor(t)

	key := Func("dummy", func() int {
		return 0
	})
	f := Get(scope, key, func() int {
		return 2
	}, nil, nil)
	if n := f(); n != 2 {
		t.Errorf("%v: got %v; want %v", key, n, 2)
	}
}

// Scope manages same name functions by its types.
func TestWorkflow_allowsSameNameFunction(t *testing.T) {
	scope := newScopeFor(t)

	key1 := Func("math/rand/v2.N", rand.N[int])
	key2 := Func("math/rand/v2.N", rand.N[int64])
	Set(scope, key1, func(int) int {
		return 10
	}, nil, nil)
	Set(scope, key2, func(int64) int64 {
		return 20
	}, nil, nil)

	f := Get(scope, key1, func(int) int {
		return 0
	}, nil, nil)
	if n := f(100); n != 10 {
		t.Errorf("%s: got %v; want %v", key1, n, 10)
	}
}

func TestFuncRecorder_bindByParamName(t *testing.T) {
	scope := newScopeFor(t)

	t.Run("explicit parameter name", func(t *testing.T) {
		r := testFuncRecorder[struct {
			KeyPath string `plug:"keyPath"`
		}](t, scope, map[string]any{
			"keyPath": "PATH",
		})
		if r.KeyPath != "PATH" {
			t.Errorf("KeyPath = %s; want PATH", r.KeyPath)
		}
	})
	t.Run("implicit parameter name", func(t *testing.T) {
		r := testFuncRecorder[struct {
			KeyPath string
		}](t, scope, map[string]any{
			"keyPath": "PATH",
		})
		if r.KeyPath != "PATH" {
			t.Errorf("KeyPath = %s; want PATH", r.KeyPath)
		}
	})
	t.Run("ignoring parameter name", func(t *testing.T) {
		r := testFuncRecorder[struct {
			KeyPath string `plug:"-"`
		}](t, scope, map[string]any{
			"keyPath": "PATH",
		})
		if r.KeyPath != "" {
			t.Errorf("KeyPath should be ignored; got %s", r.KeyPath)
		}
	})
	t.Run("no matched fields", func(t *testing.T) {
		r := testFuncRecorder[struct {
			FilePath string
		}](t, scope, map[string]any{
			"keyPath": "PATH",
		})
		if r.FilePath != "" {
			t.Errorf("FilePath should be ignored; got %s", r.FilePath)
		}
	})
}

func testFuncRecorder[T any](t *testing.T, scope *Scope, params Params) T {
	var r FuncRecorder[T]
	f := func(keyPath string) {
	}
	key := Func("func", f)
	Set(scope, key, f, nil, nil).SetRecorder(&r)
	Get(scope, key, f, nil, params)
	return r.At(0)
}
