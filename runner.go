package dogmatest

import (
	"context"
	"time"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/dogmatest/engine"
	"github.com/dogmatiq/enginekit/config"
)

// A Runner executes tests.
type Runner struct {
	engine *engine.Engine
}

// New returns a test runner.
func New(
	app dogma.Application,
	options ...RunnerOption,
) *Runner {
	ro := newRunnerOptions(options)

	cfg, err := config.NewApplicationConfig(app)
	if err != nil {
		panic(err)
	}

	e, err := engine.New(cfg, ro.engineOptions...)
	if err != nil {
		panic(err)
	}

	return &Runner{e}
}

// Begin starts a new test.
func (r *Runner) Begin(
	t T,
	options ...TestOption,
) *Test {
	return r.BeginContext(
		context.Background(),
		t,
		options...,
	)
}

// BeginContext starts a new test within a context.
func (r *Runner) BeginContext(
	ctx context.Context,
	t T,
	options ...TestOption,
) *Test {
	to := newTestOptions(options)

	r.engine.Reset()

	return &Test{
		ctx:              ctx,
		t:                t,
		engine:           r.engine,
		now:              time.Now(),
		operationOptions: to.operationOptions,
	}
}
