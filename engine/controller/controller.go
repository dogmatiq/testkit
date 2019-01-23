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
	// now is the current time, according to the engine, which may not match the
	// system time.
	//
	// nt is the time that this controller next requires Tick() to be called. If it
	// is non-nil the engine must call Tick() again at that time. This value
	// replaces any next-tick time returned by Tick() or Handle(), even if it is
	// nil.
	Tick(
		ctx context.Context,
		now time.Time,
	) (nt *time.Time, err error)

	// Handle handles a message.
	//
	// It returns the messages produced by the handler that need to be dispatched
	// by the engine.
	//
	// nt is the time that this controller next requires Tick() to be called. If it
	// is non-nil the engine must call Tick() again at that time. This value
	// replaces any next-tick time returned by Tick() or Handle(), even if it is
	// nil.
	Handle(
		ctx context.Context,
		obs fact.Observer,
		env *envelope.Envelope,
	) (
		nt *time.Time,
		out []*envelope.Envelope,
		err error,
	)

	// Reset clears the state of the controller.
	Reset()
}
