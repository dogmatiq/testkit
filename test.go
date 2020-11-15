package testkit

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/iago/must"
	"github.com/dogmatiq/testkit/assert"
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

// PrepareX prepares the application for the test by executing the given set of
// messages without any assertions.
func (t *Test) PrepareX(messages ...interface{}) *Test {
	if h, ok := t.t.(tHelper); ok {
		h.Helper()
	}

	t.logHeading("PREPARING APPLICATION FOR TEST")

	for _, m := range messages {
		t.dispatch(m, nil, assert.Nothing)
	}

	return t
}

// Prepare performs a group of actions without making any assertions in order
// to place the application into a particular state.
func (t *Test) Prepare(actions ...Action) *Test {
	if h, ok := t.t.(tHelper); ok {
		h.Helper()
	}

	for _, act := range actions {
		t.logHeading("PREPARE: " + act.Heading())

		if err := act.Apply(
			t.ctx,
			ActionScope{
				App:              t.app,
				Test:             t,
				Engine:           t.engine,
				OperationOptions: t.options(nil, assert.Nothing),
			},
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

	t.logHeading("EXPECT: " + act.Heading())

	o := ExpectOptionSet{
		MessageComparator: compare.DefaultComparator{},
	}

	for _, opt := range act.ExpectOptions() {
		opt(&o)
	}

	for _, opt := range options {
		opt(&o)
	}

	t.begin(o, e)

	if err := act.Apply(
		t.ctx,
		ActionScope{
			App:              t.app,
			Test:             t,
			Engine:           t.engine,
			OperationOptions: t.options(nil, e),
		},
	); err != nil {
		t.t.Fatal(err)
	}

	t.end(e)
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

// dispatch disaptches m to the engine.
//
// It fails the test if the engine returns an error.
func (t *Test) dispatch(
	m dogma.Message,
	options []engine.OperationOption,
	e Expectation,
) {
	if h, ok := t.t.(tHelper); ok {
		h.Helper()
	}

	opts := t.options(options, e)

	if err := t.engine.Dispatch(t.ctx, m, opts...); err != nil {
		t.t.Fatal(err)
	}
}

// options returns the full set of operation options to use for given call to
// dispatch() or tick().
func (t *Test) options(
	options []engine.OperationOption,
	e Expectation,
) (opts []engine.OperationOption) {
	if h, ok := t.t.(tHelper); ok {
		h.Helper()
	}

	opts = append(opts, engine.WithObserver(e))
	opts = append(opts, engine.WithCurrentTime(t.now)) // current test-wide time
	opts = append(opts, t.operationOptions...)         // test-wide options
	opts = append(opts, options...)                    // per-message options

	return
}

func (t *Test) begin(s assert.ExpectOptionSet, a assert.Assertion) {
	if h, ok := t.t.(tHelper); ok {
		h.Helper()
	}

	if s.MessageComparator == nil {
		s.MessageComparator = t.comparator

		if s.MessageComparator == nil {
			s.MessageComparator = compare.DefaultComparator{}
		}
	}

	a.Begin(s)
}

func (t *Test) end(a assert.Assertion) {
	if h, ok := t.t.(tHelper); ok {
		h.Helper()
	}

	a.End()

	if t.t.Failed() {
		// Don't build an assertion report if the test has already failed.
		return
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

	rep := a.BuildReport(a.Ok(), r)
	must.WriteTo(buf, rep)

	t.t.Log(buf.String())

	if !rep.TreeOk {
		t.t.FailNow()
	}
}

func (t *Test) logHeading(f string, v ...interface{}) {
	if h, ok := t.t.(tHelper); ok {
		h.Helper()
	}

	t.t.Logf(
		"--- %s ---",
		fmt.Sprintf(f, v...),
	)
}
