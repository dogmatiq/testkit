package engine

import (
	"context"
	"fmt"

	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/configkit/message"
	"github.com/dogmatiq/cosyne"
	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/testkit/engine/controller"
	"github.com/dogmatiq/testkit/engine/envelope"
	"github.com/dogmatiq/testkit/engine/fact"
	"go.uber.org/multierr"
)

// Engine is an in-memory Dogma engine that is used to execute tests.
type Engine struct {
	messageIDs envelope.MessageIDGenerator
	roles      map[message.Type]message.Role

	// m protects the controllers, routes and resetters collections. The
	// collections themselves are static and hence may be read without acquiring
	// the mutex, but m must be held in order to call any method on a
	// controller, or to call a resetter.
	m           cosyne.Mutex
	controllers map[string]controller.Controller
	routes      map[message.Type][]controller.Controller
	resetters   []func()
}

// New returns a new engine that uses the given app configuration.
func New(app configkit.RichApplication, options ...Option) (_ *Engine, err error) {
	eo := newEngineOptions(options)

	e := &Engine{
		roles:       map[message.Type]message.Role{},
		controllers: map[string]controller.Controller{},
		routes:      map[message.Type][]controller.Controller{},
		resetters:   eo.resetters,
	}

	cfgr := &configurer{
		options: eo,
		engine:  e,
	}

	ctx := context.Background()

	if err := app.AcceptRichVisitor(ctx, cfgr); err != nil {
		return nil, err
	}

	return e, nil
}

// MustNew returns a new engine that uses the given app configuration, or panics
// if unable to do so.
func MustNew(app configkit.RichApplication, options ...Option) *Engine {
	e, err := New(app, options...)
	if err != nil {
		panic(err)
	}

	return e
}

// Reset clears the engine's state, such as aggregate and process roots.
func (e *Engine) Reset() {
	e.m.Lock(context.Background())
	defer e.m.Unlock()

	e.messageIDs.Reset()

	for _, c := range e.controllers {
		c.Reset()
	}

	for _, fn := range e.resetters {
		fn()
	}
}

// Tick performs one "tick" of the engine.
//
// This allows external control of time-based features of the engine. now is the
// time that the engine should treat as the current time.
func (e *Engine) Tick(
	ctx context.Context,
	options ...OperationOption,
) error {
	oo := newOperationOptions(options)

	oo.observers.Notify(
		fact.TickCycleBegun{
			EngineTime:          oo.now,
			EnabledHandlerTypes: oo.enabledHandlerTypes,
			EnabledHandlers:     oo.enabledHandlers,
		},
	)

	err := e.m.Lock(ctx)
	if err == nil {
		defer e.m.Unlock()
		err = e.tick(ctx, oo)
	}

	oo.observers.Notify(
		fact.TickCycleCompleted{
			Error:               err,
			EnabledHandlerTypes: oo.enabledHandlerTypes,
			EnabledHandlers:     oo.enabledHandlers,
		},
	)

	return err
}

func (e *Engine) tick(
	ctx context.Context,
	oo *operationOptions,
) error {
	var (
		err   error
		queue []*envelope.Envelope
	)

	for _, c := range e.controllers {
		if skip, reason := e.skipHandler(c.HandlerConfig(), oo); skip {
			oo.observers.Notify(
				fact.TickSkipped{
					Handler: c.HandlerConfig(),
					Reason:  reason,
				},
			)

			continue
		}

		oo.observers.Notify(
			fact.TickBegun{
				Handler: c.HandlerConfig(),
			},
		)

		envs, cerr := c.Tick(ctx, oo.observers, oo.now)
		queue = append(queue, envs...)

		if cerr != nil {
			err = multierr.Append(
				err,
				fmt.Errorf(
					"%s %s: %w",
					c.HandlerConfig().Identity().Name,
					c.HandlerConfig().HandlerType(),
					cerr,
				),
			)
		}

		oo.observers.Notify(
			fact.TickCompleted{
				Handler: c.HandlerConfig(),
				Error:   cerr,
			},
		)

		if e := ctx.Err(); e != nil {
			return e
		}
	}

	return multierr.Append(
		err,
		e.dispatch(
			ctx,
			oo,
			queue...,
		),
	)
}

// Dispatch processes a message.
//
// It is not an error to process a message that is not routed to any handlers.
//
// It panics if the message is invalid.
func (e *Engine) Dispatch(
	ctx context.Context,
	m dogma.Message,
	options ...OperationOption,
) error {
	if err := dogma.ValidateMessage(m); err != nil {
		panic(fmt.Sprintf(
			"can not dispatch invalid %T message: %s",
			m,
			err,
		))
	}

	oo := newOperationOptions(options)
	t := message.TypeOf(m)

	if _, ok := e.routes[t]; !ok {
		panic(fmt.Sprintf(
			"the %s message type is not consumed by any handlers",
			t,
		))
	}

	r := e.roles[t]
	r.MustBe(message.CommandRole, message.EventRole)

	var env *envelope.Envelope
	id := e.messageIDs.Next()

	if r == message.CommandRole {
		env = envelope.NewCommand(id, m, oo.now)
	} else {
		env = envelope.NewEvent(id, m, oo.now)
	}

	oo.observers.Notify(
		fact.DispatchCycleBegun{
			Envelope:            env,
			EngineTime:          oo.now,
			EnabledHandlerTypes: oo.enabledHandlerTypes,
			EnabledHandlers:     oo.enabledHandlers,
		},
	)

	err := e.m.Lock(ctx)
	if err == nil {
		defer e.m.Unlock()
		err = e.dispatch(ctx, oo, env)
	}

	oo.observers.Notify(
		fact.DispatchCycleCompleted{
			Envelope:            env,
			Error:               err,
			EnabledHandlerTypes: oo.enabledHandlerTypes,
			EnabledHandlers:     oo.enabledHandlers,
		},
	)

	return err
}

func (e *Engine) dispatch(
	ctx context.Context,
	oo *operationOptions,
	queue ...*envelope.Envelope,
) error {
	var err error

	for len(queue) > 0 {
		env := queue[0]
		queue = queue[1:]

		var controllers []controller.Controller

		if env.Role == message.TimeoutRole {
			// always dispatch timeouts back to their origin handler
			controllers = []controller.Controller{
				e.controllers[env.Origin.Handler.Identity().Name],
			}
		} else {
			// for all other message types check to see the role matches the
			// expected role from the configuration, and if so dispatch it to
			// all of the handlers associated with that type
			r, ok := e.roles[env.Type]
			if !ok {
				continue
			}

			env.Role.MustBe(r)
			controllers = e.routes[env.Type]
		}

		oo.observers.Notify(
			fact.DispatchBegun{
				Envelope: env,
			},
		)

		var derr error
		for _, c := range controllers {
			envs, cerr := e.handle(ctx, oo, env, c)
			queue = append(queue, envs...)

			if cerr != nil {
				derr = multierr.Append(
					derr,
					fmt.Errorf(
						"%s %s: %w",
						c.HandlerConfig().Identity().Name,
						c.HandlerConfig().HandlerType(),
						cerr,
					),
				)
			}
		}

		oo.observers.Notify(
			fact.DispatchCompleted{
				Envelope: env,
				Error:    derr,
			},
		)

		err = multierr.Append(err, derr)

		if e := ctx.Err(); e != nil {
			return e
		}
	}

	return err
}

func (e *Engine) handle(
	ctx context.Context,
	oo *operationOptions,
	env *envelope.Envelope,
	c controller.Controller,
) ([]*envelope.Envelope, error) {
	if skip, reason := e.skipHandler(c.HandlerConfig(), oo); skip {
		oo.observers.Notify(
			fact.HandlingSkipped{
				Handler:  c.HandlerConfig(),
				Envelope: env,
				Reason:   reason,
			},
		)

		return nil, nil
	}

	oo.observers.Notify(
		fact.HandlingBegun{
			Handler:  c.HandlerConfig(),
			Envelope: env,
		},
	)

	envs, err := c.Handle(ctx, oo.observers, oo.now, env)

	oo.observers.Notify(
		fact.HandlingCompleted{
			Handler:  c.HandlerConfig(),
			Envelope: env,
			Error:    err,
		},
	)

	return envs, err
}

// skipHandler returns true if a specific handler should be skipped during a
// call to Dispatch() or Tick().
func (e *Engine) skipHandler(
	h configkit.Handler,
	oo *operationOptions,
) (bool, fact.HandlerSkipReason) {
	if en, ok := oo.enabledHandlers[h.Identity().Name]; ok {
		return !en, fact.IndividualHandlerDisabled
	}

	en := oo.enabledHandlerTypes[h.HandlerType()]
	return !en, fact.HandlerTypeDisabled
}
