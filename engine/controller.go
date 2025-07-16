package engine

import (
	"context"
	"time"

	"github.com/dogmatiq/enginekit/config"
	"github.com/dogmatiq/testkit/engine/internal/aggregate"
	"github.com/dogmatiq/testkit/engine/internal/integration"
	"github.com/dogmatiq/testkit/engine/internal/process"
	"github.com/dogmatiq/testkit/engine/internal/projection"
	"github.com/dogmatiq/testkit/envelope"
	"github.com/dogmatiq/testkit/fact"
)

// A controller provides Dispatch(), Tick() and Reset() functionality to the
// engine for a single Dogma message handler.
type controller interface {
	// HandlerConfig returns the config of the handler that is managed by this
	// controller.
	HandlerConfig() config.Handler

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

func registerControllers(
	e *Engine,
	opts *engineOptions,
	app *config.Application,
) {
	for _, h := range app.Handlers() {
		config.SwitchByHandlerTypeOf(
			h,
			func(h *config.Aggregate) {
				registerController(
					e,
					&aggregate.Controller{
						Config:     h,
						MessageIDs: &e.messageIDs,
					},
				)
			},
			func(h *config.Process) {
				registerController(
					e,
					&process.Controller{
						Config:     h,
						MessageIDs: &e.messageIDs,
					},
				)
			},
			func(h *config.Integration) {
				registerController(
					e,
					&integration.Controller{
						Config:     h,
						MessageIDs: &e.messageIDs,
					},
				)
			},
			func(h *config.Projection) {
				registerController(
					e,
					&projection.Controller{
						Config:                h,
						CompactDuringHandling: opts.compactDuringHandling,
					},
				)
			},
		)
	}
}

func registerController(
	e *Engine,
	ctrl controller,
) {
	cfg := ctrl.HandlerConfig()

	e.controllers[cfg.Identity().Name] = ctrl

	types := cfg.
		RouteSet().
		Filter(config.FilterByMessageDirection(config.InboundDirection)).
		MessageTypes()

	for t := range types {
		e.routes[t] = append(e.routes[t], ctrl)
	}
}
