package dogmatest

import (
	"context"
	"strings"
	"time"

	"github.com/dogmatiq/iago"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/dogmatest/assert"
	"github.com/dogmatiq/dogmatest/compare"
	"github.com/dogmatiq/dogmatest/engine"
	"github.com/dogmatiq/dogmatest/render"
)

// Test contains the state of a single test.
type Test struct {
	ctx              context.Context
	t                T
	engine           *engine.Engine
	now              time.Time
	operationOptions []engine.OperationOption
	comparator       compare.Comparator
	renderer         render.Renderer
}

// Setup prepares the application for the test by executing the given set of
// messages without any assertions.
func (t *Test) Setup(messages ...dogma.Message) *Test {
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
	if delta < 0 {
		panic("delta must be positive")
	}

	return t.AdvanceTimeTo(
		t.now.Add(delta),
		a,
		options...,
	)
}

// AdvanceTimeTo artificially advances the engine's notion of the current time
// to a specific time. The time must be greater than the current engine time.
func (t *Test) AdvanceTimeTo(
	now time.Time,
	a assert.Assertion,
	options ...engine.OperationOption,
) *Test {
	if now.Before(t.now) {
		panic("time must be greater than the current time")
	}

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
	opts := t.options(options, a)

	if err := t.engine.Dispatch(t.ctx, m, opts...); err != nil {
		t.t.Fatal(err)
	}
}

// tick ticks the engine.
// It fails the test if the engine returns an error.
func (t *Test) tick(
	options []engine.OperationOption,
	a assert.Assertion,
) {
	opts := t.options(options, a)

	if err := t.engine.Tick(t.ctx, opts...); err != nil {
		t.t.Fatal(err)
	}
}

// options returns the full set of operation options to use for given call to
// dispatch() or tick().
func (t *Test) options(
	options []engine.OperationOption,
	a assert.Assertion,
) (opts []engine.OperationOption) {
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
	if a == nil {
		panic("assertion must not be nil")
	}

	c := t.comparator
	if c == nil {
		c = compare.DefaultComparator{}
	}

	a.Begin(c)
}

func (t *Test) end(a assert.Assertion) {
	r := t.renderer
	if r == nil {
		r = render.DefaultRenderer{}
	}

	buf := &strings.Builder{}
	buf.WriteString("assertion report:\n\n")

	res := a.End(r)
	iago.MustWriteTo(buf, res)

	if res.Ok {
		t.t.Log(buf.String())
	} else {
		t.t.Fatal(buf.String())
	}
}
