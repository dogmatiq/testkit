package process

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"time"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/enginekit/config"
	"github.com/dogmatiq/enginekit/message"
	"github.com/dogmatiq/testkit/engine/internal/panicx"
	"github.com/dogmatiq/testkit/envelope"
	"github.com/dogmatiq/testkit/fact"
	"github.com/dogmatiq/testkit/location"
)

type instance struct {
	root  dogma.ProcessRoot
	ended bool
}

// Controller is an implementation of engine.Controller for
// dogma.ProcessMessageHandler implementations.
type Controller struct {
	Config     *config.Process
	MessageIDs *envelope.MessageIDGenerator

	instances map[string]*instance
	timeouts  []*envelope.Envelope
}

// HandlerConfig returns the config of the handler that is managed by this
// controller.
func (c *Controller) HandlerConfig() config.Handler {
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

	if !c.Config.RouteSet().DirectionOf(mt).Has(config.InboundDirection) {
		panic(fmt.Sprintf("%s does not handle %s messages", c.Config.Identity(), mt))
	}

	id, ok, err := c.route(ctx, obs, env)
	if !ok || err != nil {
		return nil, err
	}

	inst := c.instanceByID(obs, env, id)

	s := &scope{
		instanceID: id,
		instance:   inst,
		config:     c.Config,
		handleMethod: message.MapByKindOf(
			env.Message,
			nil,
			func(dogma.Event) string { return "HandleEvent" },
			func(dogma.Timeout) string { return "HandleTimeout" },
		),
		messageIDs: c.MessageIDs,
		observer:   obs,
		now:        now,
		env:        env,
	}

	if err := c.handle(ctx, s); err != nil {
		return nil, err
	}

	if inst.ended {
		c.cancelTimeouts(id)
		return s.commands, nil
	}

	c.scheduleTimeouts(s)
	return append(s.commands, s.ready...), nil
}

func (c *Controller) instanceByID(
	obs fact.Observer,
	env *envelope.Envelope,
	id string,
) *instance {
	if inst, ok := c.instances[id]; ok {
		obs.Notify(fact.ProcessInstanceLoaded{
			Handler:    c.Config,
			InstanceID: id,
			Root:       inst.root,
			Envelope:   env,
		})
		return inst
	}

	obs.Notify(fact.ProcessInstanceNotFound{
		Handler:    c.Config,
		InstanceID: id,
		Envelope:   env,
	})

	if c.instances == nil {
		c.instances = map[string]*instance{}
	}

	inst := &instance{
		root: c.Config.Source.Get().New(),
	}
	c.instances[id] = inst

	obs.Notify(fact.ProcessInstanceBegun{
		Handler:    c.Config,
		InstanceID: id,
		Root:       inst.root,
		Envelope:   env,
	})

	if inst.root == nil {
		panic(panicx.UnexpectedBehavior{
			Handler:        c.Config,
			Interface:      "ProcessMessageHandler",
			Method:         "New",
			Implementation: c.Config.Source.Get(),
			Message:        env.Message,
			Description:    "returned a nil ProcessRoot",
			Location:       location.OfMethod(c.Config.Source.Get(), "New"),
		})
	}

	return inst
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
	message.SwitchByKindOf(
		env.Message,
		nil,
		func(m dogma.Event) { id, ok, err = c.routeEvent(ctx, obs, env, m) },
		func(_ dogma.Timeout) { id, ok, err = c.routeTimeout(ctx, obs, env) },
	)
	return id, ok, err
}

func (c *Controller) routeEvent(
	ctx context.Context,
	obs fact.Observer,
	env *envelope.Envelope,
	m dogma.Event,
) (string, bool, error) {
	handler := c.Config.Source.Get()

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

	if !ok {
		obs.Notify(fact.ProcessEventIgnored{
			Handler:  c.Config,
			Envelope: env,
		})

		return "", false, nil
	}

	if id == "" {
		panic(panicx.UnexpectedBehavior{
			Handler:        c.Config,
			Interface:      "ProcessMessageHandler",
			Method:         "RouteEventToInstance",
			Implementation: handler,
			Message:        m,
			Description:    fmt.Sprintf("routed an event of type %s to an empty ID", message.TypeOf(m)),
			Location:       location.OfMethod(handler, "RouteEventToInstance"),
		})
	}

	if inst, ok := c.instances[id]; ok && inst.ended {
		obs.Notify(fact.ProcessEventRoutedToEndedInstance{
			Handler:    c.Config,
			InstanceID: id,
			Envelope:   env,
		})

		return "", false, nil
	}

	return id, true, nil
}

func (c *Controller) routeTimeout(
	_ context.Context,
	obs fact.Observer,
	env *envelope.Envelope,
) (string, bool, error) {
	if inst, ok := c.instances[env.Origin.InstanceID]; ok {
		if !inst.ended {
			return env.Origin.InstanceID, true, nil
		}
	}

	obs.Notify(fact.ProcessTimeoutRoutedToEndedInstance{
		Handler:    c.Config,
		InstanceID: env.Origin.InstanceID,
		Envelope:   env,
	})

	return "", false, nil
}

// handle calls the appropriate method on the handler based on the message kind.
func (c *Controller) handle(ctx context.Context, s *scope) error {
	var err error
	panicx.EnrichUnexpectedMessage(
		c.Config,
		"ProcessMessageHandler",
		s.handleMethod,
		c.Config.Source.Get(),
		s.env.Message,
		func() {
			switch m := s.env.Message.(type) {
			case dogma.Event:
				err = c.Config.Source.Get().HandleEvent(
					ctx,
					s.instance.root,
					s,
					m,
				)
			case dogma.Timeout:
				err = c.Config.Source.Get().HandleTimeout(
					ctx,
					s.instance.root,
					s,
					m,
				)
			}
		},
	)

	return err
}

// scheduleTimeouts enqueues pending timeouts from the given scope.
func (c *Controller) scheduleTimeouts(s *scope) {
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

// cancelTimeouts removes an instance's timeouts.
func (c *Controller) cancelTimeouts(id string) {
	c.timeouts = slices.DeleteFunc(
		c.timeouts,
		func(env *envelope.Envelope) bool {
			return env.Origin.InstanceID == id
		},
	)
}
