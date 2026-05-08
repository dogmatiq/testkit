package aggregate

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/enginekit/config"
	"github.com/dogmatiq/enginekit/message"
	"github.com/dogmatiq/enginekit/protobuf/uuidpb"
	"github.com/dogmatiq/testkit/engine/internal/panicx"
	"github.com/dogmatiq/testkit/envelope"
	"github.com/dogmatiq/testkit/fact"
	"github.com/dogmatiq/testkit/internal/x/xreflect"
	"github.com/dogmatiq/testkit/location"
)

type instance struct {
	// history is the set of event envelopes recorded by this instance.
	history []*envelope.Envelope

	// snapshotted is true if a snapshot of the aggregate root has been taken.
	snapshotted bool

	// snapshot is the serialized aggregate root state, populated by
	// MarshalBinary() after a successful call to the handler that produces
	// events. nil/empty is valid.
	snapshot []byte

	// snapshotOffset is the index into history at which the snapshot was
	// taken. Events before this offset are covered by the snapshot and do not
	// need to be replayed.
	snapshotOffset int
}

// Controller is an implementation of engine.Controller for
// dogma.AggregateMessageHandler implementations.
type Controller struct {
	Config     *config.Aggregate
	MessageIDs *envelope.MessageIDGenerator

	instances map[string]*instance
}

// HandlerConfig returns the config of the handler that is managed by this
// controller.
func (c *Controller) HandlerConfig() config.Handler {
	return c.Config
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
	_ context.Context,
	obs fact.Observer,
	now time.Time,
	env *envelope.Envelope,
) ([]*envelope.Envelope, error) {
	mt := message.TypeOf(env.Message)

	if !c.Config.RouteSet().DirectionOf(mt).Has(config.InboundDirection) {
		panic(fmt.Sprintf("%s does not handle %s messages", c.Config.Identity(), mt))
	}

	id := c.route(env, mt)
	inst, root := c.instanceByID(obs, env, id)

	s := &scope{
		instanceID: id,
		config:     c.Config,
		messageIDs: c.MessageIDs,
		observer:   obs,
		now:        now,
		root:       root,
		command:    env,
		streamID:   uuidpb.Derive(c.Config.Identity().GetKey(), id).AsString(),
		offset:     uint64(len(inst.history)),
	}

	panicx.EnrichUnexpectedMessage(
		c.Config,
		"AggregateMessageHandler",
		"HandleCommand",
		c.Config.Implementation(),
		env.Message,
		func() {
			c.Config.Source.Get().HandleCommand(
				root,
				s,
				env.Message.(dogma.Command),
			)
		},
	)

	if len(s.events) != 0 {
		if c.instances == nil {
			c.instances = map[string]*instance{}
		}
		inst.history = append(inst.history, s.events...)
		c.instances[id] = inst
		c.takeSnapshot(root, inst, env)
	}

	return s.events, nil
}

// Reset clears the state of the controller.
func (c *Controller) Reset() {
	c.instances = nil
}

// route returns the instance ID that the command should be routed to.
func (c *Controller) route(env *envelope.Envelope, mt message.Type) string {
	var id string
	panicx.EnrichUnexpectedMessage(
		c.Config,
		"AggregateMessageHandler",
		"RouteCommandToInstance",
		c.Config.Implementation(),
		env.Message,
		func() {
			id = c.Config.Source.Get().RouteCommandToInstance(
				env.Message.(dogma.Command),
			)
		},
	)

	if id == "" {
		panic(panicx.UnexpectedBehavior{
			Handler:        c.Config,
			Interface:      "AggregateMessageHandler",
			Method:         "RouteCommandToInstance",
			Implementation: c.Config.Implementation(),
			Message:        env.Message,
			Description:    fmt.Sprintf("routed a command of type %s to an empty ID", mt),
			Location:       location.OfMethod(c.Config.Implementation(), "RouteCommandToInstance"),
		})
	}

	return id
}

// instanceByID returns the instance and root for the given instance ID.
func (c *Controller) instanceByID(
	obs fact.Observer,
	env *envelope.Envelope,
	id string,
) (*instance, dogma.AggregateRoot) {
	root := c.Config.Source.Get().New()

	if xreflect.IsNil(root) {
		panic(panicx.UnexpectedBehavior{
			Handler:        c.Config,
			Interface:      "AggregateMessageHandler",
			Method:         "New",
			Implementation: c.Config.Implementation(),
			Message:        env.Message,
			Description:    "returned a nil aggregate root",
			Location:       location.OfMethod(c.Config.Implementation(), "New"),
		})
	}

	if inst, ok := c.instances[id]; ok {
		if inst.snapshotted {
			if err := root.UnmarshalBinary(inst.snapshot); err != nil {
				panic(panicx.UnexpectedBehavior{
					Handler:        c.Config,
					Interface:      "AggregateRoot",
					Method:         "UnmarshalBinary",
					Implementation: root,
					Message:        env.Message,
					Description:    fmt.Sprintf("unable to unmarshal the aggregate root: %s", err),
					Location:       location.OfMethod(root, "UnmarshalBinary"),
				})
			}
		}

		for _, ev := range inst.history[inst.snapshotOffset:] {
			panicx.EnrichUnexpectedMessage(
				c.Config,
				"AggregateRoot",
				"ApplyEvent",
				root,
				ev.Message,
				func() {
					root.ApplyEvent(
						ev.Message.(dogma.Event),
					)
				},
			)
		}

		obs.Notify(fact.AggregateInstanceLoaded{
			Handler:    c.Config,
			InstanceID: id,
			Root:       root,
			Envelope:   env,
		})

		return inst, root
	}

	obs.Notify(fact.AggregateInstanceNotFound{
		Handler:    c.Config,
		InstanceID: id,
		Envelope:   env,
	})

	return &instance{}, root
}

// takeSnapshot attempts to store a snapshot of the aggregate root.
func (c *Controller) takeSnapshot(
	r dogma.AggregateRoot,
	inst *instance,
	env *envelope.Envelope,
) {
	data, err := r.MarshalBinary()
	if err != nil {
		if errors.Is(err, dogma.ErrNotSupported) {
			return
		}

		panic(panicx.UnexpectedBehavior{
			Handler:        c.Config,
			Interface:      "AggregateRoot",
			Method:         "MarshalBinary",
			Implementation: r,
			Message:        env.Message,
			Description:    fmt.Sprintf("unable to marshal the aggregate root: %s", err),
			Location:       location.OfMethod(r, "MarshalBinary"),
		})
	}

	inst.snapshotted = true
	inst.snapshot = data
	inst.snapshotOffset = len(inst.history)
}
