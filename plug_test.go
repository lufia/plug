package plug

import (
	"os"
	"testing"

	"github.com/lufia/plug"
)

func TestFuncRecorder(t *testing.T) {
	scope := CurrentScopeFor(t)

	var r FuncRecorder[struct {
		Key string `plug:"key"`
	}]
	key := Func("dummy", os.Getenv)
	Set(scope, key, func(string) string {
		return "/bin:/usr/bin"
	}).SetRecorder(&r)

	defaultGetenv := func(string) string {
		return ""
	}
	Get(scope, key, defaultGetenv, WithParams(map[string]any{
		"key": "PATH",
	}))("PATH")
	if n, w := r.Count(), 1; n != w {
		t.Fatalf("Count = %d; want %d", n, w)
	}
	params := r.At(0)
	if w := "PATH"; params.Key != w {
		t.Errorf("Key = %v; want %v", params.Key, w)
	}
}
