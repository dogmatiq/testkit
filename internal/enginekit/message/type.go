package message

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/dogmatiq/dogma"
)

// Type is a value that identifies the type of a message.
type Type struct {
	id   uint64
	name string
}

func (t Type) String() string {
	return t.name
}

// TypeOf returns the message type of m.
func TypeOf(m dogma.Message) *Type {
	t := reflect.TypeOf(m)

	// try to load first, to avoid building the string if it's already stored
	mt, ok := messageTypes.Load(t)

	if !ok {
		mt, _ = messageTypes.LoadOrStore(
			t,
			newMessageType(t),
		)
	}

	return mt.(*Type)
}

var messageTypes sync.Map

func newMessageType(t reflect.Type) *Type {
	var n string
	p := reflect.ValueOf(t).Pointer()

	if t.Name() == "" {
		n = fmt.Sprintf("<anonymous %d>", p)
	} else {
		n = t.String()
	}

	return &Type{
		uint64(p),
		n,
	}
}
