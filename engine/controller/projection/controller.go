package projection

import (
	"context"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/dogmatest/engine/controller"
	"github.com/dogmatiq/dogmatest/engine/envelope"
	"github.com/dogmatiq/dogmatest/render"
)

// Controller is an implementation of engine.Controller for
// dogma.ProjectionMessageHandler implementations.
type Controller struct {
	name     string
	handler  dogma.ProjectionMessageHandler
	renderer render.Renderer
}

// NewController returns a new controller for the given handler.
func NewController(
	n string,
	h dogma.ProjectionMessageHandler,
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
func (c *Controller) Handle(ctx context.Context, cs controller.Scope) ([]*envelope.Envelope, error) {
	panic("not implemented")
}

// Reset clears the state of the controller.
func (c *Controller) Reset() {
}