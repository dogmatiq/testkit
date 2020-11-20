package aggregate

import (
	"context"
	"fmt"
	"time"

	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/configkit/message"
	"github.com/dogmatiq/testkit/engine/controller"
	"github.com/dogmatiq/testkit/engine/envelope"
	"github.com/dogmatiq/testkit/engine/fact"
)

// Controller is an implementation of engine.Controller for
// dogma.AggregateMessageHandler implementations.
type Controller struct {
	Config     configkit.RichAggregate
	MessageIDs *envelope.MessageIDGenerator

	history map[string][]*envelope.Envelope
}

// HandlerConfig returns the config of the handler that is managed by this
// controller.
func (c *Controller) HandlerConfig() configkit.RichHandler {
	return c.Config
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

	var id string
	controller.ConvertUnexpectedMessagePanic(
		c.Config,
		"AggregateMessageHandler",
		"RouteCommandToInstance",
		env.Message,
		func() {
			id = c.Config.Handler().RouteCommandToInstance(env.Message)
		},
	)

	if id == "" {
		panic(fmt.Sprintf(
			"the '%s' aggregate message handler attempted to route a %s command to an empty instance ID",
			c.Config.Identity().Name,
			message.TypeOf(env.Message),
		))
	}

	history, exists := c.history[id]
	r := c.Config.Handler().New()
	if r == nil {
		panic(fmt.Sprintf(
			"the '%s' aggregate message handler returned a nil root from New()",
			c.Config.Identity().Name,
		))
	}

	if exists {
		for _, env := range history {
			controller.ConvertUnexpectedMessagePanic(
				c.Config,
				"AggregateRoot",
				"ApplyEvent",
				env.Message,
				func() {
					r.ApplyEvent(env.Message)
				},
			)
		}

		obs.Notify(fact.AggregateInstanceLoaded{
			Handler:    c.Config,
			InstanceID: id,
			Root:       r,
			Envelope:   env,
		})
	} else {
		obs.Notify(fact.AggregateInstanceNotFound{
			Handler:    c.Config,
			InstanceID: id,
			Envelope:   env,
		})
	}

	s := &scope{
		instanceID: id,
		config:     c.Config,
		messageIDs: c.MessageIDs,
		observer:   obs,
		now:        now,
		root:       r,
		exists:     exists,
		command:    env,
	}

	controller.ConvertUnexpectedMessagePanic(
		c.Config,
		"AggregateMessageHandler",
		"HandleCommand",
		env.Message,
		func() {
			c.Config.Handler().HandleCommand(r, s, env.Message)
		},
	)

	if s.exists {
		if c.history == nil {
			c.history = map[string][]*envelope.Envelope{}
		}
		c.history[id] = append(c.history[id], s.events...)
	} else {
		delete(c.history, id)
	}

	return s.events, nil
}

// Reset clears the state of the controller.
func (c *Controller) Reset() {
	c.history = nil
}
