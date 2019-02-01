package integration

import (
	"context"
	"time"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/dogmatest/engine/envelope"
	"github.com/dogmatiq/dogmatest/engine/fact"
	"github.com/dogmatiq/enginekit/handler"
	"github.com/dogmatiq/enginekit/message"
)

// Controller is an implementation of engine.Controller for
// dogma.IntegrationMessageHandler implementations.
type Controller struct {
	name       string
	handler    dogma.IntegrationMessageHandler
	messageIDs *envelope.MessageIDGenerator
}

// NewController returns a new controller for the given handler.
func NewController(
	n string,
	h dogma.IntegrationMessageHandler,
	g *envelope.MessageIDGenerator,
) *Controller {
	return &Controller{
		name:       n,
		handler:    h,
		messageIDs: g,
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
	_ time.Time,
	env *envelope.Envelope,
) ([]*envelope.Envelope, error) {
	env.Role.MustBe(message.CommandRole)

	s := &scope{
		name:       c.name,
		handler:    c.handler,
		messageIDs: c.messageIDs,
		observer:   obs,
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
