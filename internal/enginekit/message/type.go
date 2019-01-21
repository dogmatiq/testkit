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

// NewTypeSet returns a TypeSet containing the given types.
func NewTypeSet(types ...Type) TypeSet {
	s := TypeSet{}

	for _, t := range types {
		s[t] = struct{}{}
	}

	return s
}

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

// HasM returns true if s contains TypeOf(m).
func (s TypeSet) HasM(m dogma.Message) bool {
	return s.Has(TypeOf(m))
}

// Add adds t to s.
//
// It returns true if the type was added, or false if the set already contained
// the type.
func (s TypeSet) Add(t Type) bool {
	if _, ok := s[t]; ok {
		return false
	}

	s[t] = struct{}{}
	return true
}

// AddM adds TypeOf(m) to s.
//
// It returns true if the type was added, or false if the set already contained
// the type.
func (s TypeSet) AddM(m dogma.Message) bool {
	return s.Add(TypeOf(m))
}

// Remove removes t from s.
//
// It returns true if the type was removed, or false if the set did not contain
// the type.
func (s TypeSet) Remove(t Type) bool {
	if _, ok := s[t]; ok {
		delete(s, t)
		return true
	}

	return false
}

// RemoveM removes TypeOf(m) from s.
//
// It returns true if the type was removed, or false if the set did not contain
// the type.
func (s TypeSet) RemoveM(m dogma.Message) bool {
	return s.Remove(TypeOf(m))
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
