package plug

import (
	"github.com/lufia/plug/plugcore"
)

// Recorder is the interface that wraps the Record method.
type Recorder interface {
	Record(params map[string]any)
}

// FuncRecorder records its function callings for later inspection in tests.
type FuncRecorder[T any] plugcore.FuncRecorder[T]

func (r *FuncRecorder[T]) Count() int {
	return (*plugcore.FuncRecorder[T])(r).Count()
}

func (r *FuncRecorder[T]) At(i int) T {
	return (*plugcore.FuncRecorder[T])(r).At(i)
}

func (r *FuncRecorder[T]) Record(params map[string]any) {
	(*plugcore.FuncRecorder[T])(r).Record(params)
}
