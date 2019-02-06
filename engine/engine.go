package engine

import (
	"context"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/enginekit/config"
	"github.com/dogmatiq/enginekit/message"
	"github.com/dogmatiq/testkit/engine/controller"
	"github.com/dogmatiq/testkit/engine/envelope"
	"github.com/dogmatiq/testkit/engine/fact"
	"go.uber.org/multierr"
)

// Engine is an in-memory Dogma engine that is used to execute tests.
type Engine struct {
	messageIDs  envelope.MessageIDGenerator
	controllers map[string]controller.Controller
	roles       map[message.Type]message.Role
	routes      map[message.Type][]controller.Controller
	resetters   []func()
}

// New returns a new engine that uses the given app configuration.
func New(app dogma.Application, options ...Option) (*Engine, error) {
	cfg, err := config.NewApplicationConfig(app)
	if err != nil {
		return nil, err
	}

	eo := newEngineOptions(options)

	e := &Engine{
		controllers: map[string]controller.Controller{},
		roles:       map[message.Type]message.Role{},
		routes:      map[message.Type][]controller.Controller{},
		resetters:   eo.resetters,
	}

	cfgr := &configurer{
		engine: e,
	}

	ctx := context.Background()

	if err := cfg.Accept(ctx, cfgr); err != nil {
		return nil, err
	}

	return e, nil
}

// Reset clears the engine's state, such as aggregate and process roots.
func (e *Engine) Reset() {
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
			EngineTime:      oo.now,
			EnabledHandlers: oo.enabledHandlers,
		},
	)

	err := e.tick(ctx, oo)

	oo.observers.Notify(
		fact.TickCycleCompleted{
			Error:           err,
			EnabledHandlers: oo.enabledHandlers,
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

	for n, c := range e.controllers {
		t := c.Type()

		oo.observers.Notify(
			fact.TickBegun{
				HandlerName: n,
				HandlerType: t,
			},
		)

		envs, cerr := c.Tick(ctx, oo.observers, oo.now)
		err = multierr.Append(err, cerr)
		queue = append(queue, envs...)

		oo.observers.Notify(
			fact.TickCompleted{
				HandlerName: n,
				HandlerType: t,
				Error:       cerr,
			},
		)
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
func (e *Engine) Dispatch(
	ctx context.Context,
	m dogma.Message,
	options ...OperationOption,
) error {
	oo := newOperationOptions(options)

	t := message.TypeOf(m)
	r, ok := e.roles[t]

	if !ok {
		oo.observers.Notify(
			fact.DispatchCycleSkipped{
				Message:         m,
				EngineTime:      oo.now,
				EnabledHandlers: oo.enabledHandlers,
			},
		)

		return nil
	}

	env := envelope.New(e.messageIDs.Next(), m, r)

	oo.observers.Notify(
		fact.DispatchCycleBegun{
			Envelope:        env,
			EngineTime:      oo.now,
			EnabledHandlers: oo.enabledHandlers,
		},
	)

	err := e.dispatch(ctx, oo, env)

	oo.observers.Notify(
		fact.DispatchCycleCompleted{
			Envelope:        env,
			Error:           err,
			EnabledHandlers: oo.enabledHandlers,
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
				e.controllers[env.Origin.HandlerName],
			}
		} else {
			// for all other message types check to see the role matches the expected
			// role from the configuration, and if so dispatch it to all of the handlers
			// associated with that type
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
			derr = multierr.Append(derr, cerr)
		}

		oo.observers.Notify(
			fact.DispatchCompleted{
				Envelope: env,
				Error:    derr,
			},
		)

		err = multierr.Append(err, derr)
	}

	return err
}

func (e *Engine) handle(
	ctx context.Context,
	oo *operationOptions,
	env *envelope.Envelope,
	c controller.Controller,
) ([]*envelope.Envelope, error) {
	n := c.Name()
	t := c.Type()

	if !oo.enabledHandlers[t] {
		oo.observers.Notify(
			fact.HandlingSkipped{
				HandlerName: n,
				HandlerType: c.Type(),
				Envelope:    env,
			},
		)

		return nil, nil
	}

	oo.observers.Notify(
		fact.HandlingBegun{
			HandlerName: n,
			HandlerType: t,
			Envelope:    env,
		},
	)

	envs, err := c.Handle(ctx, oo.observers, oo.now, env)

	oo.observers.Notify(
		fact.HandlingCompleted{
			HandlerName: n,
			HandlerType: t,
			Envelope:    env,
			Error:       err,
		},
	)

	return envs, err
}

// ExecuteCommand enqueues a command for execution. It panics if the command is
// not routed to any handlers.
func (e *Engine) ExecuteCommand(ctx context.Context, m dogma.Message) error {
	t := message.TypeOf(m)
	e.roles[t].MustBe(message.CommandRole)

	return e.Dispatch(ctx, m)
}

// RecordEvent records the occurrence of an event. It is not an error to record
// an event that is not routed to any handlers.
func (e *Engine) RecordEvent(ctx context.Context, m dogma.Message) error {
	return e.Dispatch(ctx, m)
}
