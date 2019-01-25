package engine

import (
	"context"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/dogmatest/engine/controller"
	"github.com/dogmatiq/dogmatest/engine/envelope"
	"github.com/dogmatiq/dogmatest/engine/fact"
	"github.com/dogmatiq/enginekit/config"
	"github.com/dogmatiq/enginekit/message"
	"go.uber.org/multierr"
)

// Engine is an in-memory Dogma engine that is used to execute tests.
type Engine struct {
	controllers map[string]controller.Controller
	roles       map[message.Type]message.Role
	routes      map[message.Type][]controller.Controller
}

// New returns a new engine that uses the given app configuration.
func New(
	cfg *config.ApplicationConfig,
	options ...Option,
) (*Engine, error) {
	e := &Engine{
		controllers: map[string]controller.Controller{},
		roles:       map[message.Type]message.Role{},
		routes:      map[message.Type][]controller.Controller{},
	}

	cfgr := &configurer{
		engine: e,
	}

	ctx := context.Background()

	for _, opt := range options {
		if err := opt(cfgr); err != nil {
			return nil, err
		}
	}

	if err := cfg.Accept(ctx, cfgr); err != nil {
		return nil, err
	}

	return e, nil
}

// Reset clears the engine's state, such as aggregate and process roots.
func (e *Engine) Reset() {
	for _, c := range e.controllers {
		c.Reset()
	}
}

// Tick performs one "tick" of the engine.
//
// This allows external control of time-based features of the engine. now is the
// time that the engine should treat as the current time.
func (e *Engine) Tick(
	ctx context.Context,
	options ...DispatchOption,
) error {
	do, err := newDispatchOptions(options)
	if err != nil {
		return err
	}

	do.observers.Notify(
		fact.EngineTickBegun{
			Now:             do.now,
			EnabledHandlers: do.enabledHandlers,
		},
	)

	err = e.tick(ctx, do)

	do.observers.Notify(
		fact.EngineTickCompleted{
			Now:             do.now,
			Error:           err,
			EnabledHandlers: do.enabledHandlers,
		},
	)

	return err
}

func (e *Engine) tick(
	ctx context.Context,
	do *dispatchOptions,
) error {
	var (
		err   error
		queue []*envelope.Envelope
	)

	for n, c := range e.controllers {
		t := c.Type()

		do.observers.Notify(
			fact.ControllerTickBegun{
				HandlerName: n,
				HandlerType: t,
				Now:         do.now,
			},
		)

		envs, cerr := c.Tick(ctx, do.observers, do.now)
		err = multierr.Append(err, cerr)
		queue = append(queue, envs...)

		do.observers.Notify(
			fact.ControllerTickCompleted{
				HandlerName: n,
				HandlerType: t,
				Now:         do.now,
				Error:       cerr,
			},
		)
	}

	return multierr.Append(
		err,
		e.dispatch(
			ctx,
			do,
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
	options ...DispatchOption,
) error {
	do, err := newDispatchOptions(options)
	if err != nil {
		return err
	}

	t := message.TypeOf(m)
	r, ok := e.roles[t]

	if !ok {
		do.observers.Notify(
			fact.UnroutableMessageDispatched{
				Message:         m,
				EnabledHandlers: do.enabledHandlers,
			},
		)

		return nil
	}

	env := envelope.New(m, r)

	do.observers.Notify(
		fact.MessageDispatchBegun{
			Envelope:        env,
			EnabledHandlers: do.enabledHandlers,
		},
	)

	err = e.dispatch(ctx, do, env)

	do.observers.Notify(
		fact.MessageDispatchCompleted{
			Envelope:        env,
			Error:           err,
			EnabledHandlers: do.enabledHandlers,
		},
	)

	return err
}

func (e *Engine) dispatch(
	ctx context.Context,
	do *dispatchOptions,
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

		for _, c := range controllers {
			envs, cerr := e.handle(ctx, do, env, c)
			queue = append(queue, envs...)
			err = multierr.Append(err, cerr)
		}
	}

	return err
}

func (e *Engine) handle(
	ctx context.Context,
	do *dispatchOptions,
	env *envelope.Envelope,
	c controller.Controller,
) ([]*envelope.Envelope, error) {
	n := c.Name()
	t := c.Type()

	if !do.enabledHandlers[t] {
		do.observers.Notify(
			fact.MessageHandlingSkipped{
				HandlerName: n,
				HandlerType: c.Type(),
				Envelope:    env,
			},
		)

		return nil, nil
	}

	do.observers.Notify(
		fact.MessageHandlingBegun{
			HandlerName: n,
			HandlerType: t,
			Envelope:    env,
		},
	)

	envs, err := c.Handle(ctx, do.observers, do.now, env)

	do.observers.Notify(
		fact.MessageHandlingCompleted{
			HandlerName: n,
			HandlerType: t,
			Envelope:    env,
			Error:       err,
		},
	)

	return envs, err
}
