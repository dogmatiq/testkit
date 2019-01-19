package fact

import (
	"context"
)

// Fact is an interface for internal engine events that occur while handling
// Dogma messages.
type Fact interface {
	// Accept calls the appropriate method on v for this event type.
	Accept(ctx context.Context, v Visitor) error
}

// Visitor is an interface for visiting engine events.
type Visitor interface {
	// VisitUnroutableMessageDispatched is called by UnroutableMessageDispatched.Accept().
	VisitUnroutableMessageDispatched(context.Context, UnroutableMessageDispatched) error

	// VisitMessageDispatchBegun is called by MessageDispatchBegun.Accept().
	VisitMessageDispatchBegun(context.Context, MessageDispatchBegun) error

	// VisitMessageDispatchCompleted is called by MessageDispatchCompleted.Accept().
	VisitMessageDispatchCompleted(context.Context, MessageDispatchCompleted) error

	// VisitMessageHandled is called by MessageHandled.Accept().
	VisitMessageHandled(context.Context, MessageHandled) error

	// VisitMessageProduced is called by MessageProduced.Accept().
	VisitMessageProduced(context.Context, MessageProduced) error
}
