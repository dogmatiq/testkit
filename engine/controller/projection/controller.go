package projection

import (
	"context"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/dogmatest/engine/envelope"
	"github.com/dogmatiq/dogmatest/engine/fact"
	"github.com/dogmatiq/dogmatest/internal/enginekit/handler"
)

// Controller is an implementation of engine.Controller for
// dogma.ProjectionMessageHandler implementations.
type Controller struct {
	name    string
	handler dogma.ProjectionMessageHandler
}

// NewController returns a new controller for the given handler.
func NewController(
	n string,
	h dogma.ProjectionMessageHandler,
) *Controller {
	return &Controller{
		name:    n,
		handler: h,
	}
}

// Name returns the name of the handler that is managed by this controller.
func (c *Controller) Name() string {
	return c.name
}

// Type returns handler.ProjectionType.
func (c *Controller) Type() handler.Type {
	return handler.ProjectionType
}

// Handle handles a message.
func (c *Controller) Handle(
	ctx context.Context,
	obs fact.ObserverSet,
	env *envelope.Envelope,
) ([]*envelope.Envelope, error) {
	s := &eventScope{
		name:      c.name,
		observers: obs,
		event:     env,
	}

	return nil, c.handler.HandleEvent(ctx, s, env.Message)
}

// Reset clears the state of the controller.
func (c *Controller) Reset() {
}
