package plugcore

import (
	"os"
	"testing"
)

func TestFuncRecorder(t *testing.T) {
	scope := NewScope(1)
	t.Cleanup(scope.Delete)

	var r FuncRecorder[struct {
		Key string `plug:"key"`
	}]
	key := Func("dummy", os.Getenv)
	Set(scope, key, func(string) string {
		return "/bin:/usr/bin"
	}, nil, nil).SetRecorder(&r)

	defaultGetenv := func(string) string {
		return ""
	}
	Get(scope, key, defaultGetenv, nil, Params{
		"key": "PATH",
	})("PATH")
	if n, w := r.Count(), 1; n != w {
		t.Fatalf("Count = %d; want %d", n, w)
	}
	params := r.At(0)
	if w := "PATH"; params.Key != w {
		t.Errorf("Key = %v; want %v", params.Key, w)
	}
}
