package plugcore

import (
	"reflect"
	"runtime"
	"slices"
)

type Scope struct {
	entry  uintptr
	name   string
	parent *Scope
	refers map[uintptr]*Scope
	mocks  map[symbolKey]*Object
	ref    int64
}

func (s *Scope) incref() int64 {
	// TODO: atomic; add up to root?
	s.ref++
	return s.ref
}

func (s *Scope) decref() int64 {
	s.ref--
	return s.ref
}

type Object struct {
	f any
	r Recorder
}

func (obj *Object) SetRecorder(r Recorder) {
	obj.r = r
}

type frame struct {
	file  string
	line  int
	entry uintptr
	name  string
}

var root Scope

func init() {
	root.entry = 0
	root.parent = &root
	root.refers = make(map[uintptr]*Scope)
}

func NewScope(skip int) *Scope {
	return NewScopeFrom(&root, skip+1)
}

func NewScopeFrom(parent *Scope, skip int) *Scope {
	frames := getFrames(skip + 1)
	slices.Reverse(frames)
	return lookupScope(parent, frames)
}

func getFrames(skip int) []*frame {
	pc := getCallers(skip + 1)
	frames := runtime.CallersFrames(pc)

	a := make([]*frame, 0, len(pc))
	for {
		f, more := frames.Next()
		a = append(a, &frame{
			file:  f.File,
			line:  f.Line,
			entry: f.Entry,
			name:  funcName(f),
		})
		if !more {
			break
		}
	}
	return a
}

func getCallers(skip int) []uintptr {
	pc := make([]uintptr, 1)
	for {
		n := runtime.Callers(skip+2, pc) // +1: Callers, +1: getCallers
		if n < len(pc) {
			return pc[:n]
		}
		pc = make([]uintptr, len(pc)*2)
	}
}

func funcName(f runtime.Frame) string {
	if f.Func == nil {
		return "(anonymous)"
	}
	return f.Func.Name()
}

func lookupScope(s *Scope, frames []*frame) *Scope {
	if len(frames) == 0 {
		s.incref()
		return s
	}
	frame, frames := frames[0], frames[1:]
	if p := s.refers[frame.entry]; p != nil {
		return lookupScope(p, frames)
	}
	p := &Scope{
		entry:  frame.entry,
		name:   frame.name,
		parent: s,
		refers: make(map[uintptr]*Scope),
		mocks:  make(map[symbolKey]*Object),
		ref:    0,
	}
	s.refers[frame.entry] = p
	return lookupScope(p, frames)
}

const doubleDeleteMessage = "double delete or corruption"

// Delete deletes all objects that were bound by Set from s then deletes s itself from the internal state.
func (s *Scope) Delete() {
	if s.ref == 0 {
		panic(doubleDeleteMessage)
	}
	if n := s.decref(); n > 0 {
		return
	}
	s.destroy()
}

func (s *Scope) destroy() {
	clear(s.mocks)
	for _, p := range s.refers {
		p.destroy()
	}
	delete(s.parent.refers, s.entry)
	s.parent = nil
}

func (s *Scope) set(key symbolKey, v any) *Object {
	mustFunc(v)
	obj := &Object{v, nullRecorder{}}
	s.mocks[key] = obj
	return obj
}

func (s *Scope) get(key symbolKey, dflt any, recv any, params Params) any {
	mustFunc(dflt)
	for s != &root {
		obj := s.mocks[key]
		if obj != nil {
			obj.r.Record(params)
			return obj.f
		}
		s = s.parent
	}
	return dflt
}

func mustFunc(f any) reflect.Value {
	v := reflect.ValueOf(f)
	if v.Type().Kind() != reflect.Func {
		panic("not a function")
	}
	return v
}
