package fixtures

import "github.com/dogmatiq/dogma"

// Message is a test implementation of dogma.Message.
type Message struct {
	Value interface{}
}

var _ dogma.Message = Message{}
