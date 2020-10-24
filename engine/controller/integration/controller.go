package integration

import (
	"context"
	"time"

	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/configkit/message"
	"github.com/dogmatiq/testkit/engine/controller"
	"github.com/dogmatiq/testkit/engine/envelope"
	"github.com/dogmatiq/testkit/engine/fact"
)

// Controller is an implementation of engine.Controller for
// dogma.IntegrationMessageHandler implementations.
type Controller struct {
	config     configkit.RichIntegration
	messageIDs *envelope.MessageIDGenerator
	produced   message.TypeCollection
}

// NewController returns a new controller for the given handler.
func NewController(
	c configkit.RichIntegration,
	g *envelope.MessageIDGenerator,
	t message.TypeCollection,
) *Controller {
	return &Controller{
		config:     c,
		messageIDs: g,
		produced:   t,
	}
}

// Identity returns the identity of the handler that is managed by this
// controller.
func (c *Controller) Identity() configkit.Identity {
	return c.config.Identity()
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

	handler := c.config.Handler()

	var t time.Duration
	controller.ConvertUnexpectedMessagePanic(
		c.config,
		"IntegrationMessageHandler",
		"TimeoutHint",
		env.Message,
		func() {
			t = handler.TimeoutHint(env.Message)
		},
	)

	if t != 0 {
		var cancel func()
		ctx, cancel = context.WithTimeout(ctx, t)
		defer cancel()
	}

	s := &scope{
		config:     c.config,
		messageIDs: c.messageIDs,
		observer:   obs,
		now:        now,
		produced:   c.produced,
		command:    env,
	}

	var err error
	controller.ConvertUnexpectedMessagePanic(
		c.config,
		"IntegrationMessageHandler",
		"HandleCommand",
		env.Message,
		func() {
			err = handler.HandleCommand(ctx, s, env.Message)
		},
	)

	return s.events, err
}

// Reset does nothing.
func (c *Controller) Reset() {
}
