package fact

import (
	"github.com/dogmatiq/dogmatest/engine/envelope"
)

// MessageHandled indicates that a message has been handled by a specific
// handler, either successfully or unsucessfully.
type MessageHandled struct {
	Envelope *envelope.Envelope
	Handler  string
	Error    error
}

// MessageProduced indicates that a message was produced by a handler.
type MessageProduced struct {
	Envelope *envelope.Envelope
	Handler  string
}
