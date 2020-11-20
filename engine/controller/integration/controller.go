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
	Config     configkit.RichIntegration
	MessageIDs *envelope.MessageIDGenerator
}

// Identity returns the identity of the handler that is managed by this
// controller.
func (c *Controller) Identity() configkit.Identity {
	return c.Config.Identity()
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

	var t time.Duration
	controller.ConvertUnexpectedMessagePanic(
		c.Config,
		"IntegrationMessageHandler",
		"TimeoutHint",
		env.Message,
		func() {
			t = c.Config.Handler().TimeoutHint(env.Message)
		},
	)

	if t != 0 {
		var cancel func()
		ctx, cancel = context.WithTimeout(ctx, t)
		defer cancel()
	}

	s := &scope{
		config:     c.Config,
		messageIDs: c.MessageIDs,
		observer:   obs,
		now:        now,
		command:    env,
	}

	var err error
	controller.ConvertUnexpectedMessagePanic(
		c.Config,
		"IntegrationMessageHandler",
		"HandleCommand",
		env.Message,
		func() {
			err = c.Config.Handler().HandleCommand(ctx, s, env.Message)
		},
	)

	return s.events, err
}

// Reset does nothing.
func (c *Controller) Reset() {
}
