package fact

import (
	"github.com/dogmatiq/dogmatest/engine/envelope"
)

// MessageHandlingBegun indicates that a message is about to be handled by a
// specific handler.
type MessageHandlingBegun struct {
	HandlerName string
	Envelope    *envelope.Envelope
}

// MessageHandlingCompleted indicates that a message has been handled by a
// specific handler, either successfully or unsucessfully.
type MessageHandlingCompleted struct {
	HandlerName string
	Envelope    *envelope.Envelope
	Error       error
}
