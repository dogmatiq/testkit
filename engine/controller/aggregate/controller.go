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
	config     configkit.RichAggregate
	messageIDs *envelope.MessageIDGenerator
	produced   message.TypeCollection
	history    map[string][]*envelope.Envelope
}

// NewController returns a new controller for the given handler.
func NewController(
	c configkit.RichAggregate,
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

// Type returns configkit.AggregateHandlerType.
func (c *Controller) Type() configkit.HandlerType {
	return configkit.AggregateHandlerType
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

	ident := c.config.Identity()
	handler := c.config.Handler()

	var id string
	controller.ConvertUnexpectedMessagePanic(
		c.config,
		"AggregateMessageHandler",
		"RouteCommandToInstance",
		env.Message,
		func() {
			id = handler.RouteCommandToInstance(env.Message)
		},
	)

	if id == "" {
		panic(fmt.Sprintf(
			"the '%s' aggregate message handler attempted to route a %s command to an empty instance ID",
			ident.Name,
			message.TypeOf(env.Message),
		))
	}

	history, exists := c.history[id]
	r := handler.New()

	if exists {
		for _, env := range history {
			controller.ConvertUnexpectedMessagePanic(
				c.config,
				"AggregateRoot",
				"ApplyEvent",
				env.Message,
				func() {
					r.ApplyEvent(env.Message)
				},
			)
		}

		obs.Notify(fact.AggregateInstanceLoaded{
			HandlerName: ident.Name,
			Handler:     handler,
			InstanceID:  id,
			Root:        r,
			Envelope:    env,
		})
	} else {
		obs.Notify(fact.AggregateInstanceNotFound{
			HandlerName: ident.Name,
			Handler:     handler,
			InstanceID:  id,
			Envelope:    env,
		})

		r = handler.New()

		if r == nil {
			panic(fmt.Sprintf(
				"the '%s' aggregate message handler returned a nil root from New()",
				ident.Name,
			))
		}
	}

	s := &scope{
		instanceID: id,
		config:     c.config,
		messageIDs: c.messageIDs,
		observer:   obs,
		now:        now,
		root:       r,
		exists:     exists,
		produced:   c.produced,
		command:    env,
	}

	controller.ConvertUnexpectedMessagePanic(
		c.config,
		"AggregateMessageHandler",
		"HandleCommand",
		env.Message,
		func() {
			handler.HandleCommand(s, env.Message)
		},
	)

	if len(s.events) == 0 && s.destroyed {
		panic(fmt.Sprintf(
			"the '%s' aggregate message handler destroyed the '%s' instance without recording an event while handling a %s command",
			ident.Name,
			id,
			message.TypeOf(env.Message),
		))
	}

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
