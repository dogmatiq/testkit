package testkit

import (
	"context"

	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/testkit/engine"
	"github.com/dogmatiq/testkit/engine/fact"
)

// A Runner executes tests.
type Runner struct {
	app    configkit.RichApplication
	engine *engine.Engine
}

// New returns a test runner.
func New(
	app dogma.Application,
	options ...RunnerOption,
) *Runner {
	cfg := configkit.FromApplication(app)
	ro := newRunnerOptions(options)

	return &Runner{
		cfg,
		engine.MustNew(cfg, ro.engineOptions...),
	}
}

// Begin starts a new test.
func (r *Runner) Begin(t TestingT, options ...TestOption) *Test {
	return r.BeginContext(
		context.Background(),
		t,
		options...,
	)
}

// BeginContext starts a new test within a context.
func (r *Runner) BeginContext(ctx context.Context, t TestingT, options ...TestOption) *Test {
	to := newTestOptions(options)

	r.engine.Reset()

	return &Test{
		ctx:    ctx,
		t:      t,
		app:    r.app,
		engine: r.engine,
		now:    to.time,
		operationOptions: append(
			to.operationOptions,
			engine.WithObserver(
				fact.NewLogger(func(s string) {
					log(t, s)
				}),
			),
		),
	}
}
