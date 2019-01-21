package process

import (
	"context"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/dogmatest/engine/controller"
	"github.com/dogmatiq/dogmatest/engine/envelope"
	"github.com/dogmatiq/dogmatest/engine/fact"
	"github.com/dogmatiq/dogmatest/internal/enginekit/handler"
)

// Controller is an implementation of engine.Controller for
// dogma.ProcessMessageHandler implementations.
type Controller struct {
	name      string
	handler   dogma.ProcessMessageHandler
	instances map[string]dogma.ProcessRoot
}

// NewController returns a new controller for the given handler.
func NewController(
	n string,
	h dogma.ProcessMessageHandler,
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

// Type returns handler.ProcessType.
func (c *Controller) Type() handler.Type {
	return handler.ProcessType
}

// Handle handles a message.
func (c *Controller) Handle(ctx context.Context, cs controller.Scope) ([]*envelope.Envelope, error) {
	env := cs.Envelope()

	id, ok, err := c.handler.RouteEventToInstance(ctx, env.Message)
	if err != nil {
		return nil, err
	}

	if !ok {
		cs.RecordFacts(fact.ProcessEventIgnored{
			HandlerName: c.name,
			Envelope:    env,
		})

		return nil, nil
	}

	if id == "" {
		panic(handler.EmptyInstanceIDError{
			HandlerName: c.name,
			HandlerType: c.Type(),
		})
	}

	r, exists := c.instances[id]

	if exists {
		cs.RecordFacts(fact.ProcessInstanceLoaded{
			HandlerName: c.name,
			InstanceID:  id,
			Root:        r,
			Envelope:    env,
		})
	} else {
		cs.RecordFacts(fact.ProcessInstanceNotFound{
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

	s := &eventScope{
		id:     id,
		name:   c.name,
		parent: cs,
		root:   r,
		exists: exists,
		event:  env,
	}

	if err := c.handler.HandleEvent(ctx, s, env.Message); err != nil {
		return nil, err
	}

	if s.exists {
		if c.instances == nil {
			c.instances = map[string]dogma.ProcessRoot{}
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
