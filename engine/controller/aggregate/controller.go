package aggregate

import (
	"context"
	"sync"
	"time"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/enginekit/handler"
	"github.com/dogmatiq/enginekit/message"
	"github.com/dogmatiq/testkit/engine/envelope"
	"github.com/dogmatiq/testkit/engine/fact"
)

// Controller is an implementation of engine.Controller for
// dogma.AggregateMessageHandler implementations.
type Controller struct {
	name       string
	handler    dogma.AggregateMessageHandler
	messageIDs *envelope.MessageIDGenerator

	m       sync.Mutex
	records map[string]*record
}

// record is a container for an aggregate root
type record struct {
	m      sync.Mutex
	root   dogma.AggregateRoot
	active bool // set to false when this record instance is removed from c.records
}

// NewController returns a new controller for the given handler.
func NewController(
	n string,
	h dogma.AggregateMessageHandler,
	g *envelope.MessageIDGenerator,
) *Controller {
	return &Controller{
		name:       n,
		handler:    h,
		messageIDs: g,
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
func (c *Controller) Tick(
	context.Context,
	fact.Observer,
	time.Time,
) ([]*envelope.Envelope, error) {
	return nil, nil
}

// Handle handles a message.
func (c *Controller) Handle(
	ctx context.Context,
	obs fact.Observer,
	_ time.Time,
	env *envelope.Envelope,
) ([]*envelope.Envelope, error) {
	env.Role.MustBe(message.CommandRole)

	id := c.handler.RouteCommandToInstance(env.Message)
	if id == "" {
		panic(handler.EmptyInstanceIDError{
			HandlerName: c.name,
			HandlerType: c.Type(),
		})
	}

	r, exists := c.lock(id)
	defer r.m.Unlock()

	if exists {
		obs.Notify(fact.AggregateInstanceLoaded{
			HandlerName: c.name,
			Handler:     c.handler,
			InstanceID:  id,
			Root:        r.root,
			Envelope:    env,
		})
	} else {
		obs.Notify(fact.AggregateInstanceNotFound{
			HandlerName: c.name,
			Handler:     c.handler,
			InstanceID:  id,
			Envelope:    env,
		})
	}

	s := &scope{
		id:         id,
		name:       c.name,
		handler:    c.handler,
		messageIDs: c.messageIDs,
		observer:   obs,
		root:       r.root,
		exists:     exists,
		command:    env,
	}

	c.handler.HandleCommand(s, env.Message)

	if (s.created || s.destroyed) && len(s.events) == 0 {
		panic(handler.EventNotRecordedError{
			HandlerName:  c.name,
			InstanceID:   id,
			WasDestroyed: s.destroyed,
		})
	}

	if !s.exists {
		c.m.Lock()
		defer c.m.Unlock()
		r.active = false
		delete(c.records, id)
	}

	return s.events, nil
}

// Reset clears the state of the controller.
func (c *Controller) Reset() {
	c.m.Lock()
	defer c.m.Unlock()

	for id, r := range c.records {
		r.m.Lock()
		r.active = false
		delete(c.records, id)
		r.m.Unlock()
	}
}

// lock the record for a specific instance, creating it if it does not exist.
func (c *Controller) lock(id string) (*record, bool) {
	for {
		// lock the controller to manipulate c.records
		c.m.Lock()
		r, exists := c.records[id]

		if exists {
			// release the controller's lock before waiting for the record lock
			// otherwise no other instances can process messages while we wait
			c.m.Unlock()

			// !!! IMPORTANT: other messages could be affecting THIS record between
			// !!! this unlock and lock call.

			r.m.Lock()

			// if this record is still "active", that means that c.records[id] == rec,
			// and we can use it. By using the active flag, we don't have to re-lock c.m
			// to check this.
			if r.active {
				return r, true
			}

			// otherwise, the instance has probably been destroyed (and perhaps
			// re-created), so we just retry from the start.
			r.m.Unlock()
			continue
		}

		// create a new root and record
		r = &record{
			root:   c.handler.New(),
			active: true,
		}

		if r.root == nil {
			c.m.Unlock()

			panic(handler.NilRootError{
				HandlerName: c.name,
				HandlerType: c.Type(),
			})
		}

		// acquire the record lock before we make it visible to other goroutines
		r.m.Lock()

		if c.records == nil {
			c.records = map[string]*record{}
		}

		c.records[id] = r

		// finally, release the controller lock
		c.m.Unlock()

		return r, false
	}
}
