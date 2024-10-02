package process

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/configkit/message"
	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/testkit/engine/internal/panicx"
	"github.com/dogmatiq/testkit/envelope"
	"github.com/dogmatiq/testkit/fact"
	"github.com/dogmatiq/testkit/location"
)

// Controller is an implementation of engine.Controller for
// dogma.ProcessMessageHandler implementations.
type Controller struct {
	Config     configkit.RichProcess
	MessageIDs *envelope.MessageIDGenerator

	instances map[string]dogma.ProcessRoot
	timeouts  []*envelope.Envelope
}

// HandlerConfig returns the config of the handler that is managed by this
// controller.
func (c *Controller) HandlerConfig() configkit.RichHandler {
	return c.Config
}

// Tick returns the timeout messages that are ready to be handled.
func (c *Controller) Tick(
	_ context.Context,
	_ fact.Observer,
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
	mt := message.TypeOf(env.Message)

	if !c.Config.MessageTypes()[mt].IsConsumed {
		panic(fmt.Sprintf("%s does not handle %s messages", c.Config.Identity(), mt))
	}

	id, ok, err := c.route(ctx, obs, env)
	if !ok || err != nil {
		return nil, err
	}

	r, exists := c.instances[id]

	if exists {
		obs.Notify(fact.ProcessInstanceLoaded{
			Handler:    c.Config,
			InstanceID: id,
			Root:       r,
			Envelope:   env,
		})
	} else {
		obs.Notify(fact.ProcessInstanceNotFound{
			Handler:    c.Config,
			InstanceID: id,
			Envelope:   env,
		})

		r = c.Config.Handler().New()

		obs.Notify(fact.ProcessInstanceBegun{
			Handler:    c.Config,
			InstanceID: id,
			Root:       r,
			Envelope:   env,
		})

		if r == nil {
			panic(panicx.UnexpectedBehavior{
				Handler:        c.Config,
				Interface:      "ProcessMessageHandler",
				Method:         "New",
				Implementation: c.Config.Handler(),
				Message:        env.Message,
				Description:    "returned a nil ProcessRoot",
				Location:       location.OfMethod(c.Config.Handler(), "New"),
			})
		}
	}

	s := &scope{
		instanceID: id,
		config:     c.Config,
		handleMethod: message.Map(
			env.Message,
			nil,
			func(dogma.Event) string { return "HandleEvent" },
			func(dogma.Timeout) string { return "HandleTimeout" },
		),
		messageIDs: c.MessageIDs,
		observer:   obs,
		now:        now,
		root:       r,
		env:        env,
	}

	if err := c.handle(ctx, s); err != nil {
		return nil, err
	}

	if s.ended {
		if exists {
			c.delete(id)
		}

		return s.commands, nil
	}

	c.update(s)

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
) (id string, ok bool, err error) {
	message.Switch(
		env.Message,
		nil,
		func(m dogma.Event) { id, ok, err = c.routeEvent(ctx, obs, env, m) },
		func(m dogma.Timeout) { id, ok, err = c.routeTimeout(ctx, obs, env) },
	)
	return id, ok, err
}

func (c *Controller) routeEvent(
	ctx context.Context,
	obs fact.Observer,
	env *envelope.Envelope,
	m dogma.Event,
) (string, bool, error) {
	handler := c.Config.Handler()

	var (
		id  string
		ok  bool
		err error
	)
	panicx.EnrichUnexpectedMessage(
		c.Config,
		"ProcessMessageHandler",
		"RouteEventToInstance",
		handler,
		m,
		func() {
			id, ok, err = handler.RouteEventToInstance(ctx, m)
		},
	)

	if err != nil {
		return "", false, err
	}

	if ok {
		if id == "" {
			panic(panicx.UnexpectedBehavior{
				Handler:        c.Config,
				Interface:      "ProcessMessageHandler",
				Method:         "RouteEventToInstance",
				Implementation: handler,
				Message:        m,
				Description:    fmt.Sprintf("routed an event of type %s to an empty ID", message.TypeOf(m)),
				Location:       location.OfMethod(c.Config.Handler(), "RouteEventToInstance"),
			})
		}

		return id, true, nil
	}

	obs.Notify(fact.ProcessEventIgnored{
		Handler:  c.Config,
		Envelope: env,
	})

	return "", false, nil
}

func (c *Controller) routeTimeout(
	_ context.Context,
	obs fact.Observer,
	env *envelope.Envelope,
) (string, bool, error) {
	if _, ok := c.instances[env.Origin.InstanceID]; ok {
		return env.Origin.InstanceID, true, nil
	}

	obs.Notify(fact.ProcessTimeoutIgnored{
		Handler:    c.Config,
		InstanceID: env.Origin.InstanceID,
		Envelope:   env,
	})

	return "", false, nil
}

// handle calls the appropriate method on the handler based on the message role.
func (c *Controller) handle(ctx context.Context, s *scope) error {
	var err error
	panicx.EnrichUnexpectedMessage(
		c.Config,
		"ProcessMessageHandler",
		s.handleMethod,
		c.Config.Handler(),
		s.env.Message,
		func() {
			switch m := s.env.Message.(type) {
			case dogma.Event:
				err = c.Config.Handler().HandleEvent(ctx, s.root, s, m)
			case dogma.Timeout:
				err = c.Config.Handler().HandleTimeout(ctx, s.root, s, m)
			}
		},
	)

	return err
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
