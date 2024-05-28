package test

import (
	"os"
	"testing"

	"github.com/lufia/plug"
)

func TestSameFunc(t *testing.T) {
	scope := plug.CurrentScopeFor(t)
	key1 := plug.Func("os.Getenv", os.Getenv)
	key2 := plug.Func("os.Getenv", os.Getenv)
	plug.Set(scope, key1, func(_ string) string {
		return "dummy1"
	})
	plug.Set(scope, key2, func(_ string) string {
		return "dummy2"
	})
}

func TestSameFile(t *testing.T) {
	scope := plug.CurrentScopeFor(t)
	keyGetenv := plug.Func("os.Getenv", os.Getenv)
	keySetenv := plug.Func("os.Setenv", os.Setenv)
	plug.Set(scope, keyGetenv, func(_ string) string {
		return "dummy"
	})
	plug.Set(scope, keySetenv, func(_, _ string) error {
		return nil
	})
}
