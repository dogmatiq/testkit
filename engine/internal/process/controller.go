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
	"github.com/dogmatiq/testkit/internal/x/xreflect"
	"github.com/dogmatiq/testkit/location"
)

type instance struct {
	data  []byte
	ended bool
}

// Controller is an implementation of engine.Controller for
// dogma.ProcessMessageHandler implementations.
type Controller struct {
	Config     *config.Process
	MessageIDs *envelope.MessageIDGenerator

	instances map[string]*instance
	deadlines []*envelope.Envelope
}

// HandlerConfig returns the config of the handler that is managed by this
// controller.
func (c *Controller) HandlerConfig() config.Handler {
	return c.Config
}

// Tick returns the deadline messages that are ready to be handled.
func (c *Controller) Tick(
	_ context.Context,
	_ fact.Observer,
	now time.Time,
) ([]*envelope.Envelope, error) {
	var i int

	// find the index of the first deadline that is AFTER now
	for _, env := range c.deadlines {
		if env.ScheduledFor.After(now) {
			break
		}

		i++
	}

	// anything up to that index is ready to be executed
	ready := c.deadlines[:i]

	// anything else is still pending
	c.deadlines = c.deadlines[i:]

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

	inst, root := c.instanceByID(obs, env, id)

	s := &scope{
		instanceID: id,
		instance:   inst,
		root:       root,
		config:     c.Config,
		handleMethod: message.MapByKindOf(
			env.Message,
			nil,
			func(dogma.Event) string { return "HandleEvent" },
			func(dogma.Deadline) string { return "HandleDeadline" },
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
		c.cancelDeadlines(id)
		return s.commands, nil
	}

	if s.mutated {
		data, err := s.root.MarshalBinary()
		if err != nil {
			panic(panicx.UnexpectedBehavior{
				Handler:        c.Config,
				Interface:      "ProcessRoot",
				Method:         "MarshalBinary",
				Implementation: s.root,
				Message:        env.Message,
				Description:    fmt.Sprintf("unable to marshal the process root: %s", err),
				Location:       location.OfMethod(s.root, "MarshalBinary"),
			})
		}
		inst.data = data
	}

	c.scheduleDeadlines(s)
	return append(s.commands, s.ready...), nil
}

func (c *Controller) instanceByID(
	obs fact.Observer,
	env *envelope.Envelope,
	id string,
) (*instance, dogma.ProcessRoot) {
	root := c.Config.Source.Get().New()

	if xreflect.IsNil(root) {
		panic(panicx.UnexpectedBehavior{
			Handler:        c.Config,
			Interface:      "ProcessMessageHandler",
			Method:         "New",
			Implementation: c.Config.Implementation(),
			Message:        env.Message,
			Description:    "returned a nil process root",
			Location:       location.OfMethod(c.Config.Implementation(), "New"),
		})
	}

	if inst, ok := c.instances[id]; ok {
		if inst.data != nil {
			if err := root.UnmarshalBinary(inst.data); err != nil {
				panic(panicx.UnexpectedBehavior{
					Handler:        c.Config,
					Interface:      "ProcessRoot",
					Method:         "UnmarshalBinary",
					Implementation: root,
					Message:        env.Message,
					Description:    fmt.Sprintf("unable to unmarshal the process root: %s", err),
					Location:       location.OfMethod(root, "UnmarshalBinary"),
				})
			}
		}

		obs.Notify(fact.ProcessInstanceLoaded{
			Handler:    c.Config,
			InstanceID: id,
			Root:       root,
			Envelope:   env,
		})
		return inst, root
	}

	obs.Notify(fact.ProcessInstanceNotFound{
		Handler:    c.Config,
		InstanceID: id,
		Envelope:   env,
	})

	if c.instances == nil {
		c.instances = map[string]*instance{}
	}

	inst := &instance{}
	c.instances[id] = inst

	obs.Notify(fact.ProcessInstanceBegun{
		Handler:    c.Config,
		InstanceID: id,
		Root:       root,
		Envelope:   env,
	})

	return inst, root
}

// Reset clears the state of the controller.
func (c *Controller) Reset() {
	c.instances = nil
	c.deadlines = nil
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
		func(_ dogma.Deadline) { id, ok, err = c.routeDeadline(ctx, obs, env) },
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
		c.Config.Implementation(),
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
			Implementation: c.Config.Implementation(),
			Message:        m,
			Description:    fmt.Sprintf("routed an event of type %s to an empty ID", message.TypeOf(m)),
			Location:       location.OfMethod(c.Config.Implementation(), "RouteEventToInstance"),
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

func (c *Controller) routeDeadline(
	_ context.Context,
	obs fact.Observer,
	env *envelope.Envelope,
) (string, bool, error) {
	if inst, ok := c.instances[env.Origin.InstanceID]; ok {
		if !inst.ended {
			return env.Origin.InstanceID, true, nil
		}
	}

	obs.Notify(fact.ProcessDeadlineRoutedToEndedInstance{
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
		c.Config.Implementation(),
		s.env.Message,
		func() {
			switch m := s.env.Message.(type) {
			case dogma.Event:
				err = c.Config.Source.Get().HandleEvent(
					ctx,
					s.root,
					s,
					m,
				)
			case dogma.Deadline:
				err = c.Config.Source.Get().HandleDeadline(
					ctx,
					s.root,
					s,
					m,
				)
			}
		},
	)

	return err
}

// scheduleDeadlines enqueues pending deadlines from the given scope.
func (c *Controller) scheduleDeadlines(s *scope) {
	c.deadlines = append(c.deadlines, s.pending...)

	sort.Slice(
		c.deadlines,
		func(i, j int) bool {
			ti := c.deadlines[i].ScheduledFor
			tj := c.deadlines[j].ScheduledFor
			return ti.Before(tj)
		},
	)
}

// cancelDeadlines removes an instance's deadlines.
func (c *Controller) cancelDeadlines(id string) {
	c.deadlines = slices.DeleteFunc(
		c.deadlines,
		func(env *envelope.Envelope) bool {
			return env.Origin.InstanceID == id
		},
	)
}
