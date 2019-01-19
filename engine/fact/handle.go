package fact

import (
	"context"

	"github.com/dogmatiq/dogmatest/engine/envelope"
)

// MessageHandled indicates that a message has been handled by a specific
// handler, either successfully or unsucessfully.
type MessageHandled struct {
	Envelope *envelope.Envelope
	Handler  string
	Error    error
}

// Accept calls v.VisitMessageHandled(ctx, ev.)
func (ev MessageHandled) Accept(ctx context.Context, v Visitor) error {
	return v.VisitMessageHandled(ctx, ev)
}

// MessageProduced indicates that a message was produced by a handler.
type MessageProduced struct {
	Envelope *envelope.Envelope
	Handler  string
}

// Accept calls v.VisitMessageProduced(ctx, ev.)
func (ev MessageProduced) Accept(ctx context.Context, v Visitor) error {
	return v.VisitMessageProduced(ctx, ev)
}
