package testkit

import (
	"context"
	"fmt"
	"strings"
	"time"

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
	engine           *engine.Engine
	executor         engine.CommandExecutor
	recorder         engine.EventRecorder
	now              time.Time
	operationOptions []engine.OperationOption
	comparator       compare.Comparator
	renderer         render.Renderer
}

// Prepare prepares the application for the test by executing the given set of
// messages without any assertions.
func (t *Test) Prepare(messages ...dogma.Message) *Test {
	if h, ok := t.t.(tHelper); ok {
		h.Helper()
	}

	t.logHeading("PREPARING APPLICATION FOR TEST")

	for _, m := range messages {
		t.dispatch(m, nil, assert.Nothing)
	}

	return t
}

// ExecuteCommand makes an assertion about the application's behavior when a
// specific command is executed.
func (t *Test) ExecuteCommand(
	m dogma.Message,
	a assert.Assertion,
	options ...engine.OperationOption,
) *Test {
	if h, ok := t.t.(tHelper); ok {
		h.Helper()
	}

	t.logHeading("EXECUTING TEST COMMAND")

	t.begin(assert.ExecuteCommandOperation, a)
	t.dispatch(m, options, a) // TODO: fail if TypeOf(m)'s role is not correct
	t.end(a)

	return t
}

// RecordEvent makes an assertion about the application's behavior when a
// specific event is recorded.
func (t *Test) RecordEvent(
	m dogma.Message,
	a assert.Assertion,
	options ...engine.OperationOption,
) *Test {
	if h, ok := t.t.(tHelper); ok {
		h.Helper()
	}

	t.logHeading("RECORDING TEST EVENT")

	t.begin(assert.RecordEventOperation, a)
	t.dispatch(m, options, a) // TODO: fail if TypeOf(m)'s role is not correct
	t.end(a)

	return t
}

// AdvanceTime artificially advances the engine's notion of the current time
// according to the given "advancer".
//
// It panics if the advancer returns a time that is before the current engine
// time.
func (t *Test) AdvanceTime(
	ta TimeAdvancer,
	a assert.Assertion,
	options ...engine.OperationOption,
) *Test {
	if h, ok := t.t.(tHelper); ok {
		h.Helper()
	}

	now, heading := ta(t.now)

	if now.Before(t.now) {
		panic("new time must be after the current time")
	}

	t.logHeading("%s", heading)

	t.now = now

	t.begin(assert.AdvanceTimeOperation, a)
	t.tick(options, a)
	t.end(a)

	return t
}

// Call makes an assertion about the application's behavior within a
// user-defined function.
//
// Code executed within fn() can make use of the command executor and event
// recorder returned by t.CommandExecutor() and t.EventRecorder(), respectively.
func (t *Test) Call(
	fn func() error,
	a assert.Assertion,
	options ...engine.OperationOption,
) *Test {
	if h, ok := t.t.(tHelper); ok {
		h.Helper()
	}

	t.executor.Engine = t.engine
	t.recorder.Engine = t.engine

	opts := t.options(options, a)
	t.executor.Options = opts
	t.recorder.Options = opts

	defer func() {
		// Ensure that the executor and recorder only use the options from this
		// test while used within Call().
		t.executor.Options = nil
		t.recorder.Options = nil
	}()

	t.logHeading("CALLING USER-DEFINED FUNCTION")

	t.begin(assert.CallOperation, a)

	if err := fn(); err != nil {
		t.t.Log(err)
		t.t.FailNow()
	}

	t.end(a)

	return t
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
	a assert.Assertion,
) {
	if h, ok := t.t.(tHelper); ok {
		h.Helper()
	}

	opts := t.options(options, a)

	if err := t.engine.Dispatch(t.ctx, m, opts...); err != nil {
		t.t.Log(err)
		t.t.FailNow()
	}
}

// tick ticks the engine.
//
// It fails the test if the engine returns an error.
func (t *Test) tick(
	options []engine.OperationOption,
	a assert.Assertion,
) {
	if h, ok := t.t.(tHelper); ok {
		h.Helper()
	}

	opts := t.options(options, a)

	if err := t.engine.Tick(t.ctx, opts...); err != nil {
		t.t.Log(err)
		t.t.FailNow()
	}
}

// options returns the full set of operation options to use for given call to
// dispatch() or tick().
func (t *Test) options(
	options []engine.OperationOption,
	a assert.Assertion,
) (opts []engine.OperationOption) {
	if h, ok := t.t.(tHelper); ok {
		h.Helper()
	}

	opts = append(opts, t.operationOptions...)         // test-wide options
	opts = append(opts, options...)                    // per-message options
	opts = append(opts, engine.WithCurrentTime(t.now)) // current test-wide time
	opts = append(opts, engine.WithObserver(a))

	return
}

func (t *Test) begin(op assert.Operation, a assert.Assertion) {
	if h, ok := t.t.(tHelper); ok {
		h.Helper()
	}

	c := t.comparator
	if c == nil {
		c = compare.DefaultComparator{}
	}

	a.Begin(op, c)
}

func (t *Test) end(a assert.Assertion) {
	if h, ok := t.t.(tHelper); ok {
		h.Helper()
	}

	a.End()

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
