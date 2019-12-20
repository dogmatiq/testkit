package controller

import (
	"context"
	"time"

	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/testkit/engine/envelope"
	"github.com/dogmatiq/testkit/engine/fact"
)

// Controller orchestrates the handling of a message by Dogma message handler.
type Controller interface {
	// Identity returns the identity of the handler that is managed by this
	// controller.
	Identity() configkit.Identity

	// Type returns the name of the handler that is managed by this controller.
	Type() configkit.HandlerType

	// Tick instructs the controller to perform an implementation-defined
	// "tick".
	//
	// It returns the messages produced by the handler that need to be
	// dispatched by the engine.
	//
	// now is the current time, according to the engine, which may not match the
	// system time.
	Tick(
		ctx context.Context,
		obs fact.Observer,
		now time.Time,
	) ([]*envelope.Envelope, error)

	// Handle handles a message.
	//
	// It returns the messages produced by the handler that need to be
	// dispatched by the engine.
	//
	// now is the current time, according to the engine, which may not match the
	// system time.
	Handle(
		ctx context.Context,
		obs fact.Observer,
		now time.Time,
		env *envelope.Envelope,
	) ([]*envelope.Envelope, error)

	// Reset clears the state of the controller.
	Reset()
}
