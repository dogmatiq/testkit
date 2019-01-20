package fact

import (
	"github.com/dogmatiq/dogmatest/engine/envelope"
)

// MessageHandlingBegun indicates that a message is about to be handled by a
// specific handler.
type MessageHandlingBegun struct {
	Envelope *envelope.Envelope
	Handler  string
}

// MessageHandlingCompleted indicates that a message has been handled by a
// specific handler, either successfully or unsucessfully.
type MessageHandlingCompleted struct {
	Envelope *envelope.Envelope
	Handler  string
	Error    error
}

// MessageProduced indicates that a message was produced by a handler.
type MessageProduced struct {
	Envelope *envelope.Envelope
	Handler  string
}
