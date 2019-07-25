package projection

import (
	"context"
	"time"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/enginekit/handler"
	"github.com/dogmatiq/enginekit/message"
	"github.com/dogmatiq/testkit/engine/envelope"
	"github.com/dogmatiq/testkit/engine/fact"
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

// Tick does nothing.
func (c *Controller) Tick(
	context.Context,
	fact.Observer,
	time.Time,
) ([]*envelope.Envelope, error) {
	return nil, nil
}

// Handle handles a message.
func (c *Controller) Handle(
	ctx context.Context,
	obs fact.Observer,
	_ time.Time,
	env *envelope.Envelope,
) ([]*envelope.Envelope, error) {
	env.Role.MustBe(message.EventRole)

	s := &scope{
		name:     c.name,
		handler:  c.handler,
		observer: obs,
		event:    env,
	}

	k := []byte(env.MessageID)

	var err error

	_, ok, err := c.handler.Recover(ctx, k)
	if err != nil {
		return nil, err
	}

	if !ok {
		err = c.handler.HandleEvent(
			ctx,
			s,
			env.Message,
			k,
			nil,
		)
	}

	if err == nil {
		err = c.handler.Discard(ctx, k)
	}

	return nil, err
}

// Reset does nothing.
func (c *Controller) Reset() {
}
