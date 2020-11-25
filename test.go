package testkit

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/iago/must"
	"github.com/dogmatiq/testkit/engine"
	"github.com/dogmatiq/testkit/fact"
)

// Test contains the state of a single test.
type Test struct {
	ctx              context.Context
	testingT         TestingT
	app              configkit.RichApplication
	virtualClock     time.Time
	engine           *engine.Engine
	executor         engine.CommandExecutor
	recorder         engine.EventRecorder
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
	cfg := configkit.FromApplication(app)

	test := &Test{
		ctx:          ctx,
		testingT:     t,
		app:          cfg,
		virtualClock: time.Now(),
		engine: engine.MustNew(
			cfg,
			engine.EnableProjectionCompactionDuringHandling(true),
		),
		operationOptions: []engine.OperationOption{
			engine.EnableProjections(false),
			engine.EnableIntegrations(false),
			engine.WithObserver(
				fact.NewLogger(func(s string) {
					log(t, s)
				}),
			),
		},
	}

	for _, opt := range options {
		opt(test)
	}

	return test
}

// Prepare performs a group of actions without making any expectations. It is
// used to place the application into a particular state.
func (t *Test) Prepare(actions ...Action) *Test {
	t.testingT.Helper()

	for _, act := range actions {
		logf(t.testingT, "--- %s ---", act.Banner())
		if err := t.applyAction(act); err != nil {
			t.testingT.Fatal(err)
		}
	}

	return t
}

// Expect ensures that a single action results in some expected behavior.
func (t *Test) Expect(act Action, e Expectation) {
	t.testingT.Helper()

	if pe, ok := e.(predicateBasedExpectation); ok {
		t.expectPred(act, pe)
	}

	o := PredicateOptions{}
	act.ConfigurePredicate(&o)

	e.Begin(o)

	func() {
		// Wrapping this defer in the closure guarantees not only that it is
		// always called, but that it is called before t.buildReport().
		defer e.End()

		logf(t.testingT, "--- EXPECT %s %s ---", act.Banner(), e.Banner())
		if err := t.applyAction(act, engine.WithObserver(e)); err != nil {
			t.testingT.Fatal(err)
		}
	}()

	if !t.testingT.Failed() {
		t.buildReport(e)
	}
}

// expectPred ensures that a single action results in some expected behavior.
func (t *Test) expectPred(act Action, e predicateBasedExpectation) {
	t.testingT.Helper()

	o := PredicateOptions{}
	act.ConfigurePredicate(&o)

	p := e.Predicate(o)

	logf(t.testingT, "--- EXPECT %s %s ---", act.Banner(), e.Banner())
	if err := t.applyAction(act, engine.WithObserver(p)); err != nil {
		t.testingT.Fatal(err)
	}

	func() {
		// Wrapping this defer in the closure guarantees not only that it is
		// always called, but that it is called before t.buildReport().
		defer e.End()

	}()

	rep := p.Done(p.Ok())

	if !t.testingT.Failed() {
		buf := &strings.Builder{}
		fmt.Fprint(
			buf,
			"--- TEST REPORT ---\n\n",
		)
		must.WriteTo(buf, rep)

		t.testingT.Log(buf.String())

		if !rep.TreeOk {
			t.testingT.FailNow()
		}
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

// EnableHandlers enables a set of handlers by name.
//
// By default all integration and projection handlers are disabled.
func (t *Test) EnableHandlers(names ...string) *Test {
	for _, n := range names {
		t.operationOptions = append(
			t.operationOptions,
			engine.EnableHandler(n, true),
		)
	}

	return t
}

// DisableHandlers disables a set of handlers by name.
//
// By default all integration and projection handlers are disabled.
func (t *Test) DisableHandlers(names ...string) *Test {
	for _, n := range names {
		t.operationOptions = append(
			t.operationOptions,
			engine.EnableHandler(n, false),
		)
	}

	return t
}

func (t *Test) buildReport(e Expectation) {
	t.testingT.Helper()

	buf := &strings.Builder{}
	fmt.Fprint(
		buf,
		"--- TEST REPORT ---\n\n",
	)

	rep := e.BuildReport(e.Ok())
	must.WriteTo(buf, rep)

	t.testingT.Log(buf.String())

	if !rep.TreeOk {
		t.testingT.FailNow()
	}
}

// applyAction calls act.Apply() with a scope appropriate for this test.
func (t *Test) applyAction(act Action, options ...engine.OperationOption) error {
	opts := []engine.OperationOption{
		engine.WithCurrentTime(t.virtualClock),
	}
	opts = append(opts, t.operationOptions...)
	opts = append(opts, options...)

	return act.Apply(
		t.ctx,
		ActionScope{
			TestingT:         t.testingT,
			App:              t.app,
			VirtualClock:     &t.virtualClock,
			Engine:           t.engine,
			Executor:         &t.executor,
			Recorder:         &t.recorder,
			OperationOptions: opts,
		},
	)
}
