package aggregate

import (
	"context"
	"fmt"
	"time"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/enginekit/config"
	"github.com/dogmatiq/enginekit/message"
	"github.com/dogmatiq/testkit/engine/internal/panicx"
	"github.com/dogmatiq/testkit/envelope"
	"github.com/dogmatiq/testkit/fact"
	"github.com/dogmatiq/testkit/location"
)

// Controller is an implementation of engine.Controller for
// dogma.AggregateMessageHandler implementations.
type Controller struct {
	Config     *config.Aggregate
	MessageIDs *envelope.MessageIDGenerator

	history map[string][]*envelope.Envelope
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

	var id string
	panicx.EnrichUnexpectedMessage(
		c.Config,
		"AggregateMessageHandler",
		"RouteCommandToInstance",
		c.Config.Source.Get(),
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
			Implementation: c.Config.Source.Get(),
			Message:        env.Message,
			Description:    fmt.Sprintf("routed a command of type %s to an empty ID", mt),
			Location:       location.OfMethod(c.Config.Source.Get(), "RouteCommandToInstance"),
		})
	}

	history, exists := c.history[id]
	r := c.Config.Source.Get().New()
	if r == nil {
		panic(panicx.UnexpectedBehavior{
			Handler:        c.Config,
			Interface:      "AggregateMessageHandler",
			Method:         "New",
			Implementation: c.Config.Source.Get(),
			Message:        env.Message,
			Description:    "returned a nil AggregateRoot",
			Location:       location.OfMethod(c.Config.Source.Get(), "New"),
		})
	}

	if exists {
		for _, env := range history {
			panicx.EnrichUnexpectedMessage(
				c.Config,
				"AggregateRoot",
				"ApplyEvent",
				r,
				env.Message,
				func() {
					r.ApplyEvent(
						env.Message.(dogma.Event),
					)
				},
			)
		}

		obs.Notify(fact.AggregateInstanceLoaded{
			Handler:    c.Config,
			InstanceID: id,
			Root:       r,
			Envelope:   env,
		})
	} else {
		obs.Notify(fact.AggregateInstanceNotFound{
			Handler:    c.Config,
			InstanceID: id,
			Envelope:   env,
		})
	}

	s := &scope{
		instanceID: id,
		config:     c.Config,
		messageIDs: c.MessageIDs,
		observer:   obs,
		now:        now,
		root:       r,
		exists:     exists,
		command:    env,
	}

	panicx.EnrichUnexpectedMessage(
		c.Config,
		"AggregateMessageHandler",
		"HandleCommand",
		c.Config.Source.Get(),
		env.Message,
		func() {
			c.Config.Source.Get().HandleCommand(
				r,
				s,
				env.Message.(dogma.Command),
			)
		},
	)

	if s.exists {
		if c.history == nil {
			c.history = map[string][]*envelope.Envelope{}
		}
		c.history[id] = append(c.history[id], s.events...)
	} else {
		delete(c.history, id)
	}

	return s.events, nil
}

// Reset clears the state of the controller.
func (c *Controller) Reset() {
	c.history = nil
}
