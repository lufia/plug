package plug

import (
	"os"
	"testing"

	"rsc.io/quote/v3"
)

func TestFunc_replacingThirdPartyPackage(t *testing.T) {
	scope := CurrentScopeFor(t)
	key := Func("rsc.io/quote/v3.HelloV3", quote.HelloV3)
	Set(scope, key, func() string {
		return "quote"
	})
	f := Get(scope, key, func() string {
		return "default"
	})
	if s := f(); s != "quote" {
		t.Errorf("%s = %v; want quote", key, s)
	}
}

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
