package aggregate

import (
	"context"
	"time"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/dogmatest/engine/envelope"
	"github.com/dogmatiq/dogmatest/engine/fact"
	"github.com/dogmatiq/dogmatest/internal/enginekit/handler"
	"github.com/dogmatiq/dogmatest/internal/enginekit/message"
)

// Controller is an implementation of engine.Controller for
// dogma.AggregateMessageHandler implementations.
type Controller struct {
	name      string
	handler   dogma.AggregateMessageHandler
	instances map[string]dogma.AggregateRoot
}

// NewController returns a new controller for the given handler.
func NewController(
	n string,
	h dogma.AggregateMessageHandler,
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

// Type returns handler.AggregateType.
func (c *Controller) Type() handler.Type {
	return handler.AggregateType
}

// Tick does nothing.
func (c *Controller) Tick(ctx context.Context, now time.Time) (*time.Time, error) {
	return nil, nil
}

// Handle handles a message.
func (c *Controller) Handle(
	ctx context.Context,
	obs fact.Observer,
	env *envelope.Envelope,
) (
	*time.Time,
	[]*envelope.Envelope,
	error,
) {
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
			InstanceID:  id,
			Root:        r,
			Envelope:    env,
		})
	} else {
		obs.Notify(fact.AggregateInstanceNotFound{
			HandlerName: c.name,
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
		id:       id,
		name:     c.name,
		observer: obs,
		root:     r,
		exists:   exists,
		command:  env,
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

	return nil, s.events, nil
}

// Reset clears the state of the controller.
func (c *Controller) Reset() {
	c.instances = nil
}
