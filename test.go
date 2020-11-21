package testkit

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/iago/must"
	"github.com/dogmatiq/testkit/compare"
	"github.com/dogmatiq/testkit/engine"
	"github.com/dogmatiq/testkit/engine/fact"
	"github.com/dogmatiq/testkit/render"
)

// Test contains the state of a single test.
type Test struct {
	ctx              context.Context
	t                TestingT
	app              configkit.RichApplication
	engine           *engine.Engine
	executor         engine.CommandExecutor
	recorder         engine.EventRecorder
	now              time.Time
	comparator       compare.Comparator
	renderer         render.Renderer
	enabledHandlers  map[configkit.HandlerType]bool
	operationOptions []engine.OperationOption
}

// Begin starts a new test.
func Begin(
	t TestingT,
	app dogma.Application,
	options ...TestOption,
) *Test {
	return BeginContext(
		context.Background(),
		t,
		app,
		options...,
	)
}

// BeginContext starts a new test within a context.
func BeginContext(
	ctx context.Context,
	t TestingT,
	app dogma.Application,
	options ...TestOption,
) *Test {
	to := newTestOptions(options)

	cfg := configkit.FromApplication(app)
	e := engine.MustNew(cfg, to.engineOptions...)

	return &Test{
		ctx:             ctx,
		t:               t,
		app:             cfg,
		engine:          e,
		now:             to.time,
		enabledHandlers: map[configkit.HandlerType]bool{},
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

// Prepare performs a group of actions without making any assertions in order
// to place the application into a particular state.
func (t *Test) Prepare(actions ...Action) *Test {
	t.t.Helper()

	for _, act := range actions {
		s := ActionScope{
			App:              t.app,
			TestingT:         t.t,
			VirtualClock:     &t.now,
			Executor:         &t.executor,
			Recorder:         &t.recorder,
			Engine:           t.engine,
			OperationOptions: t.buildOperationOptions(),
		}

		if err := act.Apply(
			t.ctx,
			s,
		); err != nil {
			t.t.Fatal(err)
		}
	}

	return t
}

// Expect ensures that a single action results in some expected behavior.
func (t *Test) Expect(act Action, e Expectation, options ...ExpectOption) {
	t.t.Helper()

	o := ExpectOptionSet{
		MessageComparator: compare.DefaultComparator{},
	}

	for _, opt := range act.ExpectOptions() {
		opt(&o)
	}

	for _, opt := range options {
		opt(&o)
	}

	e.Begin(o)

	s := ActionScope{
		App:          t.app,
		TestingT:     t.t,
		VirtualClock: &t.now,
		Executor:     &t.executor,
		Recorder:     &t.recorder,
		Engine:       t.engine,
		OperationOptions: append(
			t.buildOperationOptions(),
			engine.WithObserver(e),
		),
	}

	func() {
		defer e.End()

		if err := act.Apply(t.ctx, s); err != nil {
			t.t.Fatal(err)
		}
	}()

	if !t.t.Failed() {
		t.buildReport(e)
	}
}

// CommandExecutor returns a dogma.CommandExecutor which can be used to execute
// commands within the context of this test.
func (t *Test) CommandExecutor() dogma.CommandExecutor {
	return &t.executor
}

// EventRecorder returns a dogma.EventRecorder which can be used to record
// events within the context of this test.
func (t *Test) EventRecorder() dogma.EventRecorder {
	return &t.recorder
}

func (t *Test) buildReport(e Expectation) {
	t.t.Helper()

	r := t.renderer
	if r == nil {
		r = render.DefaultRenderer{}
	}

	buf := &strings.Builder{}
	fmt.Fprint(
		buf,
		"--- TEST REPORT ---\n\n",
	)

	rep := e.BuildReport(e.Ok(), r)
	must.WriteTo(buf, rep)

	t.t.Log(buf.String())

	if !rep.TreeOk {
		t.t.FailNow()
	}
}

// buildOperationOptions builds the operation options to provide to each action.
func (t *Test) buildOperationOptions() []engine.OperationOption {
	options := []engine.OperationOption{
		engine.EnableProjections(false),
		engine.EnableIntegrations(false),
	}

	options = append(options, t.operationOptions...)
	options = append(options, engine.WithCurrentTime(t.now))
	return options
}
