package message

import (
	"reflect"
	"sync"

	"github.com/dogmatiq/dogma"
)

// Type is a value that identifies the type of a message.
type Type interface {
	// ReflectType returns the reflect.Type for this message type.
	ReflectType() reflect.Type

	// String returns a human-readable name for the message type.
	// Note that this representation may not be globally unique.
	String() string
}

// TypeOf returns the message type of m.
func TypeOf(m dogma.Message) Type {
	rt := reflect.TypeOf(m)

	// try to load first, to avoid building the string if it's already stored
	v, loaded := mtypes.Load(rt)

	if !loaded {
		mt := newmtype(rt)

		// try to store the new mt, but if another goroutine has stored it since, use
		// that so that we get the same pointer value.
		v, loaded = mtypes.LoadOrStore(rt, mt)
		if !loaded {
			// if we stored out mt, create the reverse mapping as well
			rtypes.Store(mt, rt)
		}
	}

	return v.(*mtype)
}

// TypeSet is a collection of distinct message types.
type TypeSet map[Type]struct{}

// TypesOf returns a type set containing the types of the given messages.
func TypesOf(messages ...dogma.Message) TypeSet {
	s := TypeSet{}

	for _, m := range messages {
		s[TypeOf(m)] = struct{}{}
	}

	return s
}

// Has returns true if s contains t.
func (s TypeSet) Has(t Type) bool {
	_, ok := s[t]
	return ok
}

// Add adds t to s.
func (s TypeSet) Add(t Type) {
	s[t] = struct{}{}
}

// Remove removes t from s.
func (s TypeSet) Remove(t Type) {
	delete(s, t)
}

// AddM adds TypeOf(m) to s.
func (s TypeSet) AddM(m dogma.Message) {
	s[TypeOf(m)] = struct{}{}
}

// RemoveM removes TypeOf(m) from s.
func (s TypeSet) RemoveM(m dogma.Message) {
	delete(s, TypeOf(m))
}

var mtypes, rtypes sync.Map

type mtype string

func newmtype(rt reflect.Type) *mtype {
	var n string

	if rt.Name() == "" {
		n = "<anonymous>"
	} else {
		n = rt.String()
	}

	mt := mtype(n)

	return &mt
}

func (mt *mtype) ReflectType() reflect.Type {
	v, _ := rtypes.Load(mt)
	return v.(reflect.Type)
}

func (mt *mtype) String() string {
	return string(*mt)
}
