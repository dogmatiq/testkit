package aggregate

import (
	"context"
	"fmt"
	"time"

	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/configkit/message"
	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/testkit/engine/envelope"
	"github.com/dogmatiq/testkit/engine/fact"
)

// Controller is an implementation of engine.Controller for
// dogma.AggregateMessageHandler implementations.
type Controller struct {
	identity   configkit.Identity
	handler    dogma.AggregateMessageHandler
	messageIDs *envelope.MessageIDGenerator
	produced   message.TypeCollection
	history    map[string][]*envelope.Envelope
}

// NewController returns a new controller for the given handler.
func NewController(
	i configkit.Identity,
	h dogma.AggregateMessageHandler,
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

	id := c.handler.RouteCommandToInstance(env.Message)
	if id == "" {
		panic(fmt.Sprintf(
			"the '%s' aggregate message handler attempted to route a %s command to an empty instance ID",
			c.identity.Name,
			message.TypeOf(env.Message),
		))
	}

	history, exists := c.history[id]
	r := c.handler.New()

	if exists {
		for _, env := range history {
			r.ApplyEvent(env.Message)
		}

		obs.Notify(fact.AggregateInstanceLoaded{
			HandlerName: c.identity.Name,
			Handler:     c.handler,
			InstanceID:  id,
			Root:        r,
			Envelope:    env,
		})
	} else {
		obs.Notify(fact.AggregateInstanceNotFound{
			HandlerName: c.identity.Name,
			Handler:     c.handler,
			InstanceID:  id,
			Envelope:    env,
		})

		r = c.handler.New()

		if r == nil {
			panic(fmt.Sprintf(
				"the '%s' aggregate message handler returned a nil root from New()",
				c.identity.Name,
			))
		}
	}

	s := &scope{
		instanceID: id,
		identity:   c.identity,
		handler:    c.handler,
		messageIDs: c.messageIDs,
		observer:   obs,
		now:        now,
		root:       r,
		exists:     exists,
		produced:   c.produced,
		command:    env,
	}

	c.handler.HandleCommand(s, env.Message)

	if len(s.events) == 0 {
		if s.created {
			panic(fmt.Sprintf(
				"the '%s' aggregate message handler created the '%s' instance without recording an event while handling a %s command",
				c.identity.Name,
				id,
				message.TypeOf(env.Message),
			))
		}

		if s.destroyed {
			panic(fmt.Sprintf(
				"the '%s' aggregate message handler destroyed the '%s' instance without recording an event while handling a %s command",
				c.identity.Name,
				id,
				message.TypeOf(env.Message),
			))
		}
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
