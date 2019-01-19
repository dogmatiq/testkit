package integration

import (
	"context"

	"github.com/dogmatiq/dogmatest/engine/envelope"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/dogmatest/engine/controller"
	"github.com/dogmatiq/dogmatest/render"
)

// Controller is an implementation of engine.Controller for
// dogma.IntegrationMessageHandler implementations.
type Controller struct {
	name     string
	handler  dogma.IntegrationMessageHandler
	renderer render.Renderer
}

// NewController returns a new controller for the given handler.
func NewController(
	n string,
	h dogma.IntegrationMessageHandler,
	r render.Renderer,
) *Controller {
	return &Controller{
		name:     n,
		handler:  h,
		renderer: r,
	}
}

// Name returns the name of the handler that managed by this controller.
func (c *Controller) Name() string {
	return c.name
}

// Handle handles a message.
func (c *Controller) Handle(ctx context.Context, s controller.Scope) ([]*envelope.Envelope, error) {
	panic("not implemented")
}

// Reset clears the state of the controller.
func (c *Controller) Reset() {
}
