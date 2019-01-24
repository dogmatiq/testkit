package controller

import (
	"context"
	"time"

	"github.com/dogmatiq/dogmatest/engine/envelope"
	"github.com/dogmatiq/dogmatest/engine/fact"
	"github.com/dogmatiq/dogmatest/internal/enginekit/handler"
)

// Controller orchestrates the handling of a message by Dogma message handler.
type Controller interface {
	// Name returns the name of the handler that is managed by this controller.
	Name() string

	// Type returns the name of the handler that is managed by this controller.
	Type() handler.Type

	// Tick instructs the controller to perform an implementation-defined "tick".
	//
	// It returns the messages produced by the handler that need to be dispatched
	// by the engine.
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
	// It returns the messages produced by the handler that need to be dispatched
	// by the engine.
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
