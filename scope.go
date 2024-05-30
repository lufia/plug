package plug

import (
	"github.com/lufia/plug/plugcore"
)

type Object plugcore.Object

func (obj *Object) core() *plugcore.Object {
	return (*plugcore.Object)(obj)
}

func (obj *Object) SetRecorder(r Recorder) {
	obj.core().SetRecorder(r)
}

type Scope plugcore.Scope

func (s *Scope) core() *plugcore.Scope {
	return (*plugcore.Scope)(s)
}

// Delete deletes all objects that were bound by Set from s then deletes s itself from the internal state.
func (s *Scope) Delete() {
	s.core().Delete()
}
