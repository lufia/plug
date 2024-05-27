package plug

import (
	"reflect"
	"unicode"
)

// Recorder is the interface that wraps the Record method.
type Recorder interface {
	Record(params map[string]any)
}

// FuncRecorder records its function callings for later inspection in tests.
type FuncRecorder[T any] struct {
	calls []T
}

func (r *FuncRecorder[T]) Count() int {
	return len(r.calls)
}

func (r *FuncRecorder[T]) At(i int) T {
	return r.calls[i]
}

func (r *FuncRecorder[T]) Record(params map[string]any) {
	var call T
	v := reflect.ValueOf(&call)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	for _, f := range reflect.VisibleFields(v.Type()) {
		tag := plugTag(f)
		if tag == "-" {
			continue
		}
		param, ok := params[tag]
		if !ok {
			continue
		}
		p := v.FieldByIndex(f.Index)
		p.Set(reflect.ValueOf(param))
	}
	r.calls = append(r.calls, call)
}

func plugTag(f reflect.StructField) string {
	tag := f.Tag.Get("plug")
	if tag == "" && len(f.Name) > 0 {
		s := []rune(f.Name)
		tag = string(unicode.ToLower(s[0])) + string(s[1:])
	}
	if tag == "" || tag == "-" {
		return "-"
	}
	return tag
}

type nullRecorder struct{}

func (nullRecorder) Record(params map[string]any) {
}
