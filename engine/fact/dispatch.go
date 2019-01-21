package fact

import (
	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/dogmatest/engine/envelope"
	"github.com/dogmatiq/dogmatest/enginekit/handler"
)

// UnroutableMessageDispatched indicates that Engine.Dispatch() has been called
// with a message that is not routed to any handlers.
//
// Note that when dispatch is called with an unroutable message, it is unknown
// whether it was intended to be a command or an event.
type UnroutableMessageDispatched struct {
	// Message is the message that was dispatched.
	Message         dogma.Message
	EnabledHandlers map[handler.Type]bool
}

// MessageDispatchBegun indicates that Engine.Dispatch() has been called with a
// message that is able to be routed to at least one handler.
type MessageDispatchBegun struct {
	Envelope        *envelope.Envelope
	EnabledHandlers map[handler.Type]bool
}

// MessageDispatchCompleted indicates that a call Engine.Dispatch() has completed.
type MessageDispatchCompleted struct {
	Envelope        *envelope.Envelope
	Error           error
	EnabledHandlers map[handler.Type]bool
}
