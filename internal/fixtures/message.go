package fixtures

import "github.com/dogmatiq/dogma"

// MessageA is a test implementation of dogma.Message.
type MessageA struct {
	Value interface{}
}

// MessageB is a test implementation of dogma.Message.
type MessageB struct {
	Value interface{}
}

// MessageC is a test implementation of dogma.Message.
type MessageC struct {
	Value interface{}
}

// MessageD is a test implementation of dogma.Message.
type MessageD struct {
	Value interface{}
}

var (
	_ dogma.Message = MessageA{}
	_ dogma.Message = MessageB{}
	_ dogma.Message = MessageC{}
	_ dogma.Message = MessageD{}
)
