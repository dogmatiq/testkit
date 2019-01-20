package aggregate

import (
	"context"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/dogmatest/engine/controller"
	"github.com/dogmatiq/dogmatest/engine/envelope"
	"github.com/dogmatiq/dogmatest/engine/fact"
	"github.com/dogmatiq/dogmatest/internal/enginekit/errors"
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

// Name returns the name of the handler that managed by this controller.
func (c *Controller) Name() string {
	return c.name
}

// Handle handles a message.
func (c *Controller) Handle(ctx context.Context, cs controller.Scope) ([]*envelope.Envelope, error) {
	env := cs.Envelope()

	id := c.handler.RouteCommandToInstance(env.Message)
	if id == "" {
		return nil, errors.EmptyInstanceID{
			Handler: c.name,
		}
	}

	r, exists := c.instances[id]

	if exists {
		cs.RecordFacts(fact.AggregateInstanceLoaded{
			HandlerName: c.name,
			InstanceID:  id,
			Root:        r,
			Envelope:    env,
		})
	} else {
		cs.RecordFacts(fact.AggregateInstanceNotFound{
			HandlerName: c.name,
			InstanceID:  id,
			Envelope:    env,
		})
	}

	r = c.handler.New()

	if r == nil {
		return nil, errors.NilRoot{
			Handler: c.name,
		}
	}

	s := &commandScope{
		id:      id,
		name:    c.name,
		parent:  cs,
		root:    r,
		exists:  exists,
		command: env,
	}

	c.handler.HandleCommand(s, env.Message)

	if s.exists {
		if c.instances == nil {
			c.instances = map[string]dogma.AggregateRoot{}
		}
		c.instances[id] = s.root
	} else {
		delete(c.instances, id)
	}

	return s.children, nil
}

// Reset clears the state of the controller.
func (c *Controller) Reset() {
	c.instances = nil
}
