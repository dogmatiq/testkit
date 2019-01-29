package dogmatest

import (
	"context"
	"strings"
	"time"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/dogmatest/assert"
	"github.com/dogmatiq/dogmatest/compare"
	"github.com/dogmatiq/dogmatest/engine"
	"github.com/dogmatiq/dogmatest/render"
)

// Test contains the state of a single test.
type Test interface {
	// Setup prepares the application for the test by executing the given set of
	// messages without any assertions.
	Setup(...dogma.Message) Test

	ExecuteCommand(dogma.Message, assert.Assertion) Test
	RecordEvent(dogma.Message, assert.Assertion) Test
	AdvanceTimeBy(time.Duration, assert.Assertion) Test
	AdvanceTimeTo(time.Time, assert.Assertion) Test
}

type test struct {
	ctx        context.Context
	t          T
	engine     *engine.Engine
	now        time.Time
	defaults   []engine.DispatchOption
	comparator compare.Comparator
	renderer   render.Renderer
}

func (t *test) Setup(messages ...dogma.Message) Test {
	for _, m := range messages {
		t.dispatch(m, nil)
	}

	return t
}

func (t *test) ExecuteCommand(m dogma.Message, a assert.Assertion) Test {
	if a == nil {
		panic("assertion must not be nil")
	}

	// TODO: fail if TypeOf(m)'s role is not correct
	t.dispatch(m, a)

	return t
}

func (t *test) RecordEvent(m dogma.Message, a assert.Assertion) Test {
	if a == nil {
		panic("assertion must not be nil")
	}

	// TODO: fail if TypeOf(m)'s role is not correct
	t.dispatch(m, a)

	return t
}

func (t *test) AdvanceTimeBy(delta time.Duration, a assert.Assertion) Test {
	if a == nil {
		panic("assertion must not be nil")
	}

	if delta < 0 {
		panic("delta must be positive")
	}

	t.now.Add(delta)
	t.tick(a)

	return t
}

func (t *test) AdvanceTimeTo(now time.Time, a assert.Assertion) Test {
	if a == nil {
		panic("assertion must not be nil")
	}

	if now.Before(t.now) {
		panic("time must be greater than the current time")
	}

	t.now = now
	t.tick(a)

	return t
}

func (t *test) dispatch(m dogma.Message, a assert.Assertion) {
	t.begin(a)

	opts := t.options(a)

	if err := t.engine.Dispatch(t.ctx, m, opts...); err != nil {
		t.t.Fatal(err)
	}

	t.end(a)
}

func (t *test) tick(a assert.Assertion) {
	t.begin(a)

	opts := t.options(a)

	if err := t.engine.Tick(t.ctx, opts...); err != nil {
		t.t.Fatal(err)
	}

	t.end(a)
}

func (t *test) options(a assert.Assertion) []engine.DispatchOption {
	return append(
		t.defaults,
		engine.WithCurrentTime(t.now),
		engine.WithObserver(a),
	)
}

func (t *test) begin(a assert.Assertion) {
	c := t.comparator
	if c == nil {
		c = compare.DefaultComparator{}
	}

	a.Begin(c)
}

func (t *test) end(a assert.Assertion) {
	r := t.renderer
	if r == nil {
		r = render.DefaultRenderer{}
	}

	buf := &strings.Builder{}
	buf.WriteString("assertion report:\n\n")

	if a.End(buf, r) {
		t.t.Log(buf.String())
	} else {
		t.t.Fatal(buf.String())
	}
}
