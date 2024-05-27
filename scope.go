package plug

import (
	"reflect"
	"runtime"
	"slices"
)

type Scope struct {
	entry  uintptr
	parent *Scope
	refers map[uintptr]*Scope
	mocks  map[symbolKey]*Object
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
}

var root Scope

func init() {
	root.entry = 0
	root.parent = &root
	root.refers = make(map[uintptr]*Scope)
}

func newScope(skip int) *Scope {
	frames := getFrames(skip + 1)
	slices.Reverse(frames)
	return lookupScope(&root, frames)
}

func getFrames(skip int) []*frame {
	pc := make([]uintptr, 100)       // TODO: grow
	n := runtime.Callers(skip+2, pc) // +1: Callers, +1: getFrames
	pc = pc[:n]
	frames := runtime.CallersFrames(pc)

	a := make([]*frame, 0, len(pc))
	for {
		f, more := frames.Next()
		a = append(a, &frame{
			file:  f.File,
			line:  f.Line,
			entry: f.Entry,
		})
		if !more {
			break
		}
	}
	return a
}

func lookupScope(s *Scope, frames []*frame) *Scope {
	if len(frames) == 0 {
		return s
	}
	frame, frames := frames[0], frames[1:]
	if p := s.refers[frame.entry]; p != nil {
		return lookupScope(p, frames)
	}
	p := &Scope{
		entry:  frame.entry,
		parent: s,
		refers: make(map[uintptr]*Scope),
		mocks:  make(map[symbolKey]*Object),
	}
	s.refers[frame.entry] = p
	return lookupScope(p, frames)
}

// Delete deletes all objects that were bound by Set from s then deletes s itself from the internal state.
func (s *Scope) Delete() {
	clear(s.mocks)
	for _, p := range s.refers {
		p.Delete()
	}
	delete(s.parent.refers, s.entry)
	s.parent = nil
}

type constraints struct {
	recv   any
	params map[string]any
}

func (s *Scope) set(key symbolKey, v any) *Object {
	mustFunc(v)
	obj := &Object{v, nullRecorder{}}
	s.mocks[key] = obj
	return obj
}

func (s *Scope) get(key symbolKey, dflt any, c *constraints) any {
	mustFunc(dflt)
	for s != &root {
		obj := s.mocks[key]
		if obj != nil {
			obj.r.Record(c.params)
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
