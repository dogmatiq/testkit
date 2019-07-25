package integration

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
// dogma.IntegrationMessageHandler implementations.
type Controller struct {
	name       string
	handler    dogma.IntegrationMessageHandler
	messageIDs *envelope.MessageIDGenerator
	produced   message.TypeContainer
}

// NewController returns a new controller for the given handler.
func NewController(
	n string,
	h dogma.IntegrationMessageHandler,
	g *envelope.MessageIDGenerator,
	t message.TypeContainer,
) *Controller {
	return &Controller{
		name:       n,
		handler:    h,
		messageIDs: g,
		produced:   t,
	}
}

// Name returns the name of the handler that is managed by this controller.
func (c *Controller) Name() string {
	return c.name
}

// Type returns handler.IntegrationType.
func (c *Controller) Type() handler.Type {
	return handler.IntegrationType
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
	now time.Time,
	env *envelope.Envelope,
) ([]*envelope.Envelope, error) {
	env.Role.MustBe(message.CommandRole)

	s := &scope{
		name:       c.name,
		handler:    c.handler,
		messageIDs: c.messageIDs,
		observer:   obs,
		now:        now,
		produced:   c.produced,
		command:    env,
	}

	if t := c.handler.TimeoutHint(env.Message); t != 0 {
		var cancel func()
		ctx, cancel = context.WithTimeout(ctx, t)
		defer cancel()
	}

	if err := c.handler.HandleCommand(ctx, s, env.Message); err != nil {
		return nil, err
	}

	return s.events, nil
}

// Reset does nothing.
func (c *Controller) Reset() {
}
