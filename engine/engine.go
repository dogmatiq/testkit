package engine

import (
	"context"
	"fmt"

	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/configkit/message"
	"github.com/dogmatiq/cosyne"
	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/testkit/envelope"
	"github.com/dogmatiq/testkit/fact"
	"github.com/dogmatiq/testkit/internal/validation"
	"go.uber.org/multierr"
)

// Engine is an in-memory Dogma engine that is used to execute tests.
type Engine struct {
	messageIDs envelope.MessageIDGenerator

	// m protects the controllers, routes and resetters collections. The
	// collections themselves are static and hence may be read without acquiring
	// the mutex, but m must be held in order to call any method on a
	// controller, or to call a resetter.
	m           cosyne.Mutex
	controllers map[string]controller
	routes      map[message.Type][]controller
	resetters   []func()
}

// New returns a new engine that uses the given app configuration.
func New(app configkit.RichApplication, options ...Option) (_ *Engine, err error) {
	eo := newEngineOptions(options)

	e := &Engine{
		controllers: map[string]controller{},
		routes:      map[message.Type][]controller{},
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
	_ = e.m.Lock(context.Background())
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
	oo := newOperationOptions(e, options)

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

// Dispatch processes a [dogma.Command] or [dogma.Event].
//
// It panics if the message is a [dogma.Timeout], or is otherwise invalid.
func (e *Engine) Dispatch(
	ctx context.Context,
	m dogma.Message,
	options ...OperationOption,
) error {
	mt := message.TypeOf(m)
	id := e.messageIDs.Next()
	oo := newOperationOptions(e, options)

	env, err := message.TryMap(
		m,
		func(m dogma.Command) (*envelope.Envelope, error) {
			if err := m.Validate(validation.CommandValidationScope()); err != nil {
				return nil, err
			}
			return envelope.NewCommand(id, m, oo.now), nil
		},
		func(m dogma.Event) (*envelope.Envelope, error) {
			if err := m.Validate(validation.EventValidationScope()); err != nil {
				return nil, err
			}
			return envelope.NewEvent(id, m, oo.now), nil
		},
		func(t dogma.Timeout) (*envelope.Envelope, error) {
			panic("cannot dispatch timeout messages")
		},
	)
	if err != nil {
		panic(fmt.Sprintf("cannot dispatch invalid %s message: %s", mt, err))
	}

	if _, ok := e.routes[mt]; !ok {
		panic(fmt.Sprintf("the %s message type is not consumed by any handlers", mt))
	}

	oo.observers.Notify(
		fact.DispatchCycleBegun{
			Envelope:            env,
			EngineTime:          oo.now,
			EnabledHandlerTypes: oo.enabledHandlerTypes,
			EnabledHandlers:     oo.enabledHandlers,
		},
	)

	err = e.m.Lock(ctx)
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

		var controllers []controller

		mt := message.TypeOf(env.Message)

		if mt.Kind() == message.TimeoutKind {
			// always dispatch timeouts back to their origin handler
			controllers = []controller{
				e.controllers[env.Origin.Handler.Identity().Name],
			}
		} else {
			controllers = e.routes[mt]
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
	c controller,
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

	if h.IsDisabled() {
		return true, fact.IndividualHandlerDisabledByConfiguration
	}

	en := oo.enabledHandlerTypes[h.HandlerType()]
	return !en, fact.HandlerTypeDisabled
}
