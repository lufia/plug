package mock

import (
	"reflect"
	"runtime"
	"slices"
)

var mem map[uintptr]map[uintptr]reflect.Value

func Set[F any](f, m F) {
	v := reflect.ValueOf(f)
	if v.Type().Kind() != reflect.Func {
		panic("not function")
	}
	pc, _, _, _ := runtime.Caller(1)
	frames := runtime.CallersFrames([]uintptr{pc})
	frame, _ := frames.Next()

	if mem == nil {
		mem = make(map[uintptr]map[uintptr]reflect.Value)
	}
	if mem[v.Pointer()] == nil {
		mem[v.Pointer()] = make(map[uintptr]reflect.Value)
	}
	mem[v.Pointer()][frame.Entry] = reflect.ValueOf(m)
}

func Get[F any](dflt F) F {
	v := reflect.ValueOf(dflt)
	if v.Type().Kind() != reflect.Func {
		panic("not function")
	}
	pcs := make([]uintptr, 100)
	n := runtime.Callers(1, pcs)
	pcs = pcs[:n]

	callers := make([]uintptr, 0, len(pcs))
	frames := runtime.CallersFrames(pcs)
	for {
		frame, more := frames.Next()
		callers = append(callers, frame.Entry)
		if !more {
			break
		}
	}
	slices.Reverse(callers)

	m := mem[v.Pointer()]
	for _, c := range callers {
		v, ok := m[c]
		if !ok {
			continue
		}
		if f, ok := v.Interface().(F); ok {
			return f
		}
	}
	return dflt
}
