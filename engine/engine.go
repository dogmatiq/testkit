package engine

import (
	"context"
	"time"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/dogmatest/compare"
	"github.com/dogmatiq/dogmatest/engine/controller"
	"github.com/dogmatiq/dogmatest/engine/envelope"
	"github.com/dogmatiq/dogmatest/engine/fact"
	"github.com/dogmatiq/dogmatest/enginekit/config"
	"github.com/dogmatiq/dogmatest/enginekit/message"
	"github.com/dogmatiq/dogmatest/render"
	"go.uber.org/multierr"
)

// Engine is an in-memory Dogma engine that is used to execute tests.
type Engine struct {
	roles  map[message.Type]message.Role
	routes map[message.Type][]controller.Controller
}

// New returns a new engine that uses the given app configuration.
func New(
	cfg *config.AppConfig,
	options ...Option,
) (*Engine, error) {
	e := &Engine{
		roles:  map[message.Type]message.Role{},
		routes: map[message.Type][]controller.Controller{},
	}

	cfgr := &configurer{
		engine:     e,
		comparator: compare.DefaultComparator{},
		renderer:   render.DefaultRenderer{},
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
	for _, controllers := range e.routes {
		for _, c := range controllers {
			c.Reset()
		}
	}
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

	env := envelope.New(m, r, time.Time{})

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
	env *envelope.Envelope,
) error {
	var err error
	queue := []*envelope.Envelope{env}

	for len(queue) > 0 {
		env := queue[0]
		queue = queue[1:]

		s := &scope{
			env:       env,
			observers: do.observers,
		}

		for _, c := range e.routes[env.Type] {
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

				continue
			}

			do.observers.Notify(
				fact.MessageHandlingBegun{
					HandlerName: n,
					HandlerType: t,
					Envelope:    env,
				},
			)

			envs, herr := c.Handle(ctx, s)

			if herr != nil {
				err = multierr.Append(err, herr)
			} else {
				queue = append(queue, envs...)
			}

			do.observers.Notify(
				fact.MessageHandlingCompleted{
					HandlerName: n,
					HandlerType: t,
					Envelope:    env,
					Error:       herr,
				},
			)
		}
	}

	return err
}
