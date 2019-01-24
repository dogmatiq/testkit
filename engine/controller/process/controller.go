package process

import (
	"context"
	"sort"
	"time"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/dogmatest/engine/envelope"
	"github.com/dogmatiq/dogmatest/engine/fact"
	"github.com/dogmatiq/dogmatest/internal/enginekit/handler"
	"github.com/dogmatiq/dogmatest/internal/enginekit/message"
)

// Controller is an implementation of engine.Controller for
// dogma.ProcessMessageHandler implementations.
type Controller struct {
	name      string
	handler   dogma.ProcessMessageHandler
	instances map[string]dogma.ProcessRoot
	timeouts  []*envelope.Envelope
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

// Tick returns the timeout messages that are ready to be handled.
func (c *Controller) Tick(
	ctx context.Context,
	obs fact.Observer,
	now time.Time,
) ([]*envelope.Envelope, error) {
	var i int

	// find the index of the first timeout that is AFTER now
	for _, env := range c.timeouts {
		if env.TimeoutTime.After(now) {
			break
		}

		i++
	}

	// anything up to that index is ready to be executed
	ready := c.timeouts[:i]

	// anything else is still pending
	c.timeouts = c.timeouts[i:]

	return ready, nil
}

// Handle handles a message.
func (c *Controller) Handle(
	ctx context.Context,
	obs fact.Observer,
	now time.Time,
	env *envelope.Envelope,
) ([]*envelope.Envelope, error) {
	env.Role.MustBe(message.EventRole, message.TimeoutRole)

	id, ok, err := c.route(ctx, obs, env)
	if !ok || err != nil {
		return nil, err
	}

	r, exists := c.instances[id]

	if exists {
		obs.Notify(fact.ProcessInstanceLoaded{
			HandlerName: c.name,
			InstanceID:  id,
			Root:        r,
			Envelope:    env,
		})
	} else {
		obs.Notify(fact.ProcessInstanceNotFound{
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
		now:      now,
		root:     r,
		exists:   exists,
		env:      env,
	}

	if err := c.handle(ctx, s); err != nil {
		return nil, err
	}

	if s.exists {
		c.update(s)
	} else if exists {
		c.delete(id)
	}

	return append(s.commands, s.ready...), nil
}

// Reset clears the state of the controller.
func (c *Controller) Reset() {
	c.instances = nil
	c.timeouts = nil
}

// route returns the ID of the instance that a message should be routed to.
func (c *Controller) route(
	ctx context.Context,
	obs fact.Observer,
	env *envelope.Envelope,
) (string, bool, error) {
	if env.Role == message.TimeoutRole {
		// ignore any timeout for instances that no longer exist
		_, ok := c.instances[env.Origin.InstanceID]
		return env.Origin.InstanceID, ok, nil
	}

	id, ok, err := c.handler.RouteEventToInstance(ctx, env.Message)
	if err != nil {
		return "", false, err
	}

	if !ok {
		obs.Notify(fact.ProcessEventIgnored{
			HandlerName: c.name,
			Envelope:    env,
		})

		return "", false, nil
	}

	if id == "" {
		panic(handler.EmptyInstanceIDError{
			HandlerName: c.name,
			HandlerType: c.Type(),
		})
	}

	return id, true, nil
}

// handle calls the appropriate method on the handler based on the  message role.
func (c *Controller) handle(ctx context.Context, s *scope) error {
	if s.env.Role == message.TimeoutRole {
		return c.handler.HandleTimeout(ctx, s, s.env.Message)
	}

	return c.handler.HandleEvent(ctx, s, s.env.Message)
}

// update stores the process root and its pending timeouts.
func (c *Controller) update(s *scope) {
	if c.instances == nil {
		c.instances = map[string]dogma.ProcessRoot{}
	}

	c.instances[s.id] = s.root
	c.timeouts = append(c.timeouts, s.pending...)

	sort.Slice(
		c.timeouts,
		func(i, j int) bool {
			ti := *c.timeouts[i].TimeoutTime
			tj := *c.timeouts[j].TimeoutTime
			return ti.Before(tj)
		},
	)
}

// delete removes an instance and its pending timeouts from the store.
func (c *Controller) delete(id string) {
	delete(c.instances, id)

	timeouts := make([]*envelope.Envelope, 0, len(c.timeouts))

	// filter out any existing timeouts that belong to the deleted instance
	for _, env := range c.timeouts {
		if env.Origin.InstanceID != id {
			timeouts = append(timeouts, env)
		}
	}

	c.timeouts = timeouts
}
