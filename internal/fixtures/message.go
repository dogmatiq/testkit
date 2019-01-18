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

// MessageE is a test implementation of dogma.Message.
type MessageE struct {
	Value interface{}
}

// MessageF is a test implementation of dogma.Message.
type MessageF struct {
	Value interface{}
}

// MessageG is a test implementation of dogma.Message.
type MessageG struct {
	Value interface{}
}

var (
	_ dogma.Message = MessageA{}
	_ dogma.Message = MessageB{}
	_ dogma.Message = MessageC{}
	_ dogma.Message = MessageD{}
	_ dogma.Message = MessageE{}
	_ dogma.Message = MessageF{}
	_ dogma.Message = MessageG{}
)
