package fact

import (
	"github.com/dogmatiq/dogmatest/engine/envelope"
	"github.com/dogmatiq/enginekit/handler"
)

// MessageHandlingBegun indicates that a message is about to be handled by a
// specific handler.
type MessageHandlingBegun struct {
	HandlerName string
	HandlerType handler.Type
	Envelope    *envelope.Envelope
}

// MessageHandlingSkipped indicates that a message has been not been handled by
// a specific handler, because handlers of that type are disabled.
type MessageHandlingSkipped struct {
	HandlerName string
	HandlerType handler.Type
	Envelope    *envelope.Envelope
}

// MessageHandlingCompleted indicates that a message has been handled by a
// specific handler, either successfully or unsucessfully.
type MessageHandlingCompleted struct {
	HandlerName string
	HandlerType handler.Type
	Envelope    *envelope.Envelope
	Error       error
}
