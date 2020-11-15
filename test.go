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
	operationOptions []engine.OperationOption
	comparator       compare.Comparator
	renderer         render.Renderer
}

// Prepare performs a group of actions without making any assertions in order
// to place the application into a particular state.
func (t *Test) Prepare(actions ...Action) *Test {
	if h, ok := t.t.(tHelper); ok {
		h.Helper()
	}

	for _, act := range actions {
		s := ActionScope{
			App:      t.app,
			TestingT: t.t,
			Test:     t,
			Engine:   t.engine,
		}

		s.OperationOptions = append(s.OperationOptions, t.operationOptions...)
		s.OperationOptions = append(s.OperationOptions, engine.WithCurrentTime(t.now))

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
	if h, ok := t.t.(tHelper); ok {
		h.Helper()
	}

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
		App:      t.app,
		TestingT: t.t,
		Test:     t,
		Engine:   t.engine,
	}

	s.OperationOptions = append(s.OperationOptions, t.operationOptions...)
	s.OperationOptions = append(
		s.OperationOptions,
		engine.WithCurrentTime(t.now),
		engine.WithObserver(e),
	)

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
	if h, ok := t.t.(tHelper); ok {
		h.Helper()
	}

	r := t.renderer
	if r == nil {
		r = render.DefaultRenderer{}
	}

	buf := &strings.Builder{}
	fmt.Fprint(
		buf,
		"--- ASSERTION REPORT ---\n\n",
	)

	rep := e.BuildReport(e.Ok(), r)
	must.WriteTo(buf, rep)

	t.t.Log(buf.String())

	if !rep.TreeOk {
		t.t.FailNow()
	}
}
