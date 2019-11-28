package testkit

import (
	"context"
	"fmt"
	"path"
	"runtime"
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
	t                T
	verbose          bool
	engine           *engine.Engine
	now              time.Time
	operationOptions []engine.OperationOption
	comparator       compare.Comparator
	renderer         render.Renderer
}

// Prepare prepares the application for the test by executing the given set of
// messages without any assertions.
func (t *Test) Prepare(messages ...dogma.Message) *Test {
	t.t.Helper()

	if t.verbose {
		t.logHeading("PREPARING APPLICATION FOR TEST")
	}

	for _, m := range messages {
		t.dispatch(m, nil, nil)
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
	t.t.Helper()

	if t.verbose {
		t.logHeading("EXECUTING TEST COMMAND")
	}

	t.begin(a)
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
	t.t.Helper()

	if t.verbose {
		t.logHeading("RECORDING TEST EVENT")
	}

	t.begin(a)
	t.dispatch(m, options, a) // TODO: fail if TypeOf(m)'s role is not correct
	t.end(a)

	return t
}

// AdvanceTimeBy artificially advances the engine's notion of the current time
// by a fixed duration. The duration must be positive.
func (t *Test) AdvanceTimeBy(
	delta time.Duration,
	a assert.Assertion,
	options ...engine.OperationOption,
) *Test {
	t.t.Helper()

	if delta < 0 {
		panic("delta must be positive")
	}

	if t.verbose {
		t.logHeading("ADVANCING TIME BY %s", delta)
	}

	return t.advanceTime(t.now.Add(delta), a, options)
}

// AdvanceTimeTo artificially advances the engine's notion of the current time
// to a specific time. The time must be greater than the current engine time.
func (t *Test) AdvanceTimeTo(
	now time.Time,
	a assert.Assertion,
	options ...engine.OperationOption,
) *Test {
	t.t.Helper()

	if now.Before(t.now) {
		panic("time must be greater than the current time")
	}

	if t.verbose {
		t.logHeading("ADVANCING TIME TO %s", now.Format(time.RFC3339))
	}

	return t.advanceTime(now, a, options)
}

// advanceTime artificially advances the engine's notion of the current time
// to a specific time.
func (t *Test) advanceTime(
	now time.Time,
	a assert.Assertion,
	options []engine.OperationOption,
) *Test {
	t.t.Helper()

	t.now = now

	t.begin(a)
	t.tick(options, a)
	t.end(a)

	return t
}

// dispatch disaptches m to the engine.
// It fails the test if the engine returns an error.
func (t *Test) dispatch(
	m dogma.Message,
	options []engine.OperationOption,
	a assert.Assertion,
) {
	t.t.Helper()

	opts := t.options(options, a)

	if err := t.engine.Dispatch(t.ctx, m, opts...); err != nil {
		t.t.Log(err)
		t.t.FailNow()
	}
}

// tick ticks the engine.
// It fails the test if the engine returns an error.
func (t *Test) tick(
	options []engine.OperationOption,
	a assert.Assertion,
) {
	t.t.Helper()

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
	t.t.Helper()

	opts = append(opts, t.operationOptions...)         // test-wide options
	opts = append(opts, options...)                    // per-message options
	opts = append(opts, engine.WithCurrentTime(t.now)) // current test-wide time

	if a != nil {
		// add the assertion as an observer.
		opts = append(opts, engine.WithObserver(a))
	}

	return
}

func (t *Test) begin(a assert.Assertion) {
	t.t.Helper()

	if a == nil {
		panic("assertion must not be nil")
	}

	c := t.comparator
	if c == nil {
		c = compare.DefaultComparator{}
	}

	a.Prepare(c)
}

func (t *Test) end(a assert.Assertion) {
	t.t.Helper()

	r := t.renderer
	if r == nil {
		r = render.DefaultRenderer{}
	}

	caller := t.findCaller()
	buf := &strings.Builder{}
	fmt.Fprintf(
		buf,
		"--- ASSERTION REPORT (%s:%d) ---\n\n",
		path.Base(caller.File),
		caller.Line,
	)

	rep := a.BuildReport(a.Ok(), r)
	must.WriteTo(buf, rep)

	t.t.Log(buf.String())

	if !rep.TreeOk {
		t.t.FailNow()
	}
}

func (t *Test) logHeading(f string, v ...interface{}) {
	t.t.Helper()

	caller := t.findCaller()

	t.t.Logf(
		"--- %s (%s:%d) ---",
		fmt.Sprintf(f, v...),
		path.Base(caller.File),
		caller.Line,
	)
}

// findCaller returns the frame of the deepest caller in the stack that is
// not a method of the testkit.Test type.
func (t *Test) findCaller() (f runtime.Frame) {
	t.t.Helper()

	const window = 5
	offset := 2 // start by excluding this function and runtime.Callers()

	for {
		pc := make([]uintptr, window)
		n := runtime.Callers(offset, pc)
		pc = pc[:n]
		offset += window

		frames := runtime.CallersFrames(pc)
		more := true

		for more {
			f, more = frames.Next()
			if !strings.HasPrefix(f.Function, "github.com/dogmatiq/testkit.(*Test)") {
				return
			}
		}
	}
}
