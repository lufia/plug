package plug

import (
	"testing"
)

func TestFuncRecorder(t *testing.T) {
	scope := CurrentScope()
	t.Cleanup(scope.Delete)
	fake := func() {}

	type Params struct {
		Lat float64 `plug:"lat"`
		Lng float64 `plug:"lng"`
	}
	var r FuncRecorder[Params]
	key := Func("dummy", fake)
	Set(scope, key, fake).SetRecorder(&r)

	Get(scope, key, fake, WithParams(map[string]any{
		"lat": 32.1,
		"lng": -18.8,
	}))
	if n, w := r.Count(), 1; n != w {
		t.Errorf("Count = %d; want %d", n, w)
	}
	params := r.At(0)
	if w := 32.1; params.Lat != w {
		t.Errorf("Lat = %v; want %v", params.Lat, w)
	}
	if w := -18.8; params.Lng != w {
		t.Errorf("Lng = %v; want %v", params.Lng, w)
	}
}
