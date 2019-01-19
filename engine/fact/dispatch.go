package fact

import (
	"context"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/dogmatest/engine/envelope"
)

// UnroutableMessageDispatched indicates that Engine.Dispatch() has been called
// with a message that is not routed to any handlers.
//
// Note that when dispatch is called with an unroutable message, it is unknown
// whether it was intended to be a command or an event.
type UnroutableMessageDispatched struct {
	// Message is the message that was dispatched.
	Message dogma.Message
}

// Accept calls v.VisitMessageDispatchBegun(ctx, ev.)
func (ev UnroutableMessageDispatched) Accept(ctx context.Context, v Visitor) error {
	return v.VisitUnroutableMessageDispatched(ctx, ev)
}

// MessageDispatchBegun indicates that Engine.Dispatch() has been called with a
// message that is able to be routed to at least one handler.
type MessageDispatchBegun struct {
	Envelope *envelope.Envelope
}

// Accept calls v.VisitMessageDispatchBegun(ctx, ev.)
func (ev MessageDispatchBegun) Accept(ctx context.Context, v Visitor) error {
	return v.VisitMessageDispatchBegun(ctx, ev)
}

// MessageDispatchCompleted indicates that a call Engine.Dispatch() has completed.
type MessageDispatchCompleted struct {
	Envelope *envelope.Envelope
	Error    error
}

// Accept calls v.VisitMessageDispatchComplete(ctx, ev.)
func (ev MessageDispatchCompleted) Accept(ctx context.Context, v Visitor) error {
	return v.VisitMessageDispatchCompleted(ctx, ev)
}
