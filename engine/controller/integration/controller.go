package integration

import (
	"context"
	"time"

	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/configkit/message"
	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/testkit/engine/envelope"
	"github.com/dogmatiq/testkit/engine/fact"
)

// Controller is an implementation of engine.Controller for
// dogma.IntegrationMessageHandler implementations.
type Controller struct {
	identity   configkit.Identity
	handler    dogma.IntegrationMessageHandler
	messageIDs *envelope.MessageIDGenerator
	produced   message.TypeCollection
}

// NewController returns a new controller for the given handler.
func NewController(
	i configkit.Identity,
	h dogma.IntegrationMessageHandler,
	g *envelope.MessageIDGenerator,
	t message.TypeCollection,
) *Controller {
	return &Controller{
		identity:   i,
		handler:    h,
		messageIDs: g,
		produced:   t,
	}
}

// Identity returns the identity of the handler that is managed by this
// controller.
func (c *Controller) Identity() configkit.Identity {
	return c.identity
}

// Type returns configkit.IntegrationHandlerType.
func (c *Controller) Type() configkit.HandlerType {
	return configkit.IntegrationHandlerType
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

	if t := c.handler.TimeoutHint(env.Message); t != 0 {
		var cancel func()
		ctx, cancel = context.WithTimeout(ctx, t)
		defer cancel()
	}

	s := &scope{
		identity:   c.identity,
		handler:    c.handler,
		messageIDs: c.messageIDs,
		observer:   obs,
		now:        now,
		produced:   c.produced,
		command:    env,
	}

	if err := c.handler.HandleCommand(ctx, s, env.Message); err != nil {
		return nil, err
	}

	return s.events, nil
}

// Reset does nothing.
func (c *Controller) Reset() {
}
