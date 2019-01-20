package integration

import (
	"context"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/dogmatest/engine/controller"
	"github.com/dogmatiq/dogmatest/engine/envelope"
)

// Controller is an implementation of engine.Controller for
// dogma.IntegrationMessageHandler implementations.
type Controller struct {
	name    string
	handler dogma.IntegrationMessageHandler
}

// NewController returns a new controller for the given handler.
func NewController(
	n string,
	h dogma.IntegrationMessageHandler,
) *Controller {
	return &Controller{
		name:    n,
		handler: h,
	}
}

// Name returns the name of the handler that managed by this controller.
func (c *Controller) Name() string {
	return c.name
}

// Handle handles a message.
func (c *Controller) Handle(ctx context.Context, cs controller.Scope) ([]*envelope.Envelope, error) {
	env := cs.Envelope()

	s := &commandScope{
		name:    c.name,
		parent:  cs,
		command: env,
	}

	if err := c.handler.HandleCommand(ctx, s, env.Message); err != nil {
		return nil, err
	}

	return s.children, nil
}

// Reset clears the state of the controller.
func (c *Controller) Reset() {
}
