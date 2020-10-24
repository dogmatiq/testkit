package process

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/configkit/message"
	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/testkit/engine/envelope"
	"github.com/dogmatiq/testkit/engine/fact"
)

// Controller is an implementation of engine.Controller for
// dogma.ProcessMessageHandler implementations.
type Controller struct {
	config     configkit.RichProcess
	messageIDs *envelope.MessageIDGenerator
	produced   message.TypeCollection
	instances  map[string]dogma.ProcessRoot
	timeouts   []*envelope.Envelope
}

// NewController returns a new controller for the given handler.
func NewController(
	c configkit.RichProcess,
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

// Type returns configkit.ProcessHandlerType.
func (c *Controller) Type() configkit.HandlerType {
	return configkit.ProcessHandlerType
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
		if env.ScheduledFor.After(now) {
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

	ident := c.config.Identity()
	handler := c.config.Handler()

	if t := handler.TimeoutHint(env.Message); t != 0 {
		var cancel func()
		ctx, cancel = context.WithTimeout(ctx, t)
		defer cancel()
	}

	id, ok, err := c.route(ctx, obs, env)
	if !ok || err != nil {
		return nil, err
	}

	r, exists := c.instances[id]

	if exists {
		obs.Notify(fact.ProcessInstanceLoaded{
			HandlerName: ident.Name,
			Handler:     handler,
			InstanceID:  id,
			Root:        r,
			Envelope:    env,
		})
	} else {
		obs.Notify(fact.ProcessInstanceNotFound{
			HandlerName: ident.Name,
			Handler:     handler,
			InstanceID:  id,
			Envelope:    env,
		})

		r = handler.New()

		if r == nil {
			panic(fmt.Sprintf(
				"the '%s' process message handler returned a nil root from New()",
				ident.Name,
			))
		}
	}

	s := &scope{
		instanceID: id,
		identity:   ident,
		handler:    handler,
		messageIDs: c.messageIDs,
		observer:   obs,
		now:        now,
		root:       r,
		exists:     exists,
		produced:   c.produced,
		env:        env,
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
	if env.Role == message.EventRole {
		return c.routeEvent(ctx, obs, env)
	}

	return c.routeTimeout(ctx, obs, env)
}

func (c *Controller) routeEvent(
	ctx context.Context,
	obs fact.Observer,
	env *envelope.Envelope,
) (string, bool, error) {
	ident := c.config.Identity()
	handler := c.config.Handler()

	id, ok, err := handler.RouteEventToInstance(ctx, env.Message)
	if err != nil {
		return "", false, err
	}

	if ok {
		if id == "" {
			panic(fmt.Sprintf(
				"the '%s' process message handler attempted to route a %s event to an empty instance ID",
				ident.Name,
				message.TypeOf(env.Message),
			))
		}

		return id, true, nil
	}

	obs.Notify(fact.ProcessEventIgnored{
		HandlerName: ident.Name,
		Handler:     handler,
		Envelope:    env,
	})

	return "", false, nil
}

func (c *Controller) routeTimeout(
	ctx context.Context,
	obs fact.Observer,
	env *envelope.Envelope,
) (string, bool, error) {
	// ignore any timeout for instances that no longer exist
	if _, ok := c.instances[env.Origin.InstanceID]; ok {
		return env.Origin.InstanceID, true, nil
	}

	obs.Notify(fact.ProcessTimeoutIgnored{
		HandlerName: c.config.Identity().Name,
		Handler:     c.config.Handler(),
		InstanceID:  env.Origin.InstanceID,
		Envelope:    env,
	})

	return "", false, nil
}

// handle calls the appropriate method on the handler based on the message role.
func (c *Controller) handle(ctx context.Context, s *scope) error {
	if s.env.Role == message.EventRole {
		return c.config.Handler().HandleEvent(ctx, s, s.env.Message)
	}

	return c.config.Handler().HandleTimeout(ctx, s, s.env.Message)
}

// update stores the process root and its pending timeouts.
func (c *Controller) update(s *scope) {
	if c.instances == nil {
		c.instances = map[string]dogma.ProcessRoot{}
	}

	c.instances[s.instanceID] = s.root
	c.timeouts = append(c.timeouts, s.pending...)

	sort.Slice(
		c.timeouts,
		func(i, j int) bool {
			ti := c.timeouts[i].ScheduledFor
			tj := c.timeouts[j].ScheduledFor
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
