package aggregate

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
// dogma.AggregateMessageHandler implementations.
type Controller struct {
	name       string
	handler    dogma.AggregateMessageHandler
	messageIDs *envelope.MessageIDGenerator
	produced   message.TypeContainer
	instances  map[string]dogma.AggregateRoot
}

// NewController returns a new controller for the given handler.
func NewController(
	n string,
	h dogma.AggregateMessageHandler,
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

// Type returns handler.AggregateType.
func (c *Controller) Type() handler.Type {
	return handler.AggregateType
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
		panic(handler.EmptyInstanceIDError{
			HandlerName: c.name,
			HandlerType: c.Type(),
		})
	}

	r, exists := c.instances[id]

	if exists {
		obs.Notify(fact.AggregateInstanceLoaded{
			HandlerName: c.name,
			Handler:     c.handler,
			InstanceID:  id,
			Root:        r,
			Envelope:    env,
		})
	} else {
		obs.Notify(fact.AggregateInstanceNotFound{
			HandlerName: c.name,
			Handler:     c.handler,
			InstanceID:  id,
			Envelope:    env,
		})

		r = c.handler.New()

		if r == nil {
			panic(handler.NilRootError{
				HandlerName: c.name,
				HandlerType: c.Type(),
			})
		}
	}

	s := &scope{
		id:         id,
		name:       c.name,
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

	if (s.created || s.destroyed) && len(s.events) == 0 {
		panic(handler.EventNotRecordedError{
			HandlerName:  c.name,
			InstanceID:   id,
			WasDestroyed: s.destroyed,
		})
	}

	if s.exists {
		if c.instances == nil {
			c.instances = map[string]dogma.AggregateRoot{}
		}
		c.instances[id] = s.root
	} else {
		delete(c.instances, id)
	}

	return s.events, nil
}

// Reset clears the state of the controller.
func (c *Controller) Reset() {
	c.instances = nil
}
