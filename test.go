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

	ExecuteCommand(dogma.Message, ...assert.Assertion) Test
	RecordEvent(dogma.Message, ...assert.Assertion) Test
	AdvanceTimeBy(time.Duration, ...assert.Assertion) Test
	AdvanceTimeTo(time.Time, ...assert.Assertion) Test
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

func (t *test) ExecuteCommand(m dogma.Message, assertions ...assert.Assertion) Test {
	// TODO: add assertion that m is executed a command
	t.dispatch(m, assertions)

	return t
}

func (t *test) RecordEvent(m dogma.Message, assertions ...assert.Assertion) Test {
	// TODO: add assertion that m is recorded as an event
	t.dispatch(m, assertions)

	return t
}

func (t *test) AdvanceTimeBy(delta time.Duration, assertions ...assert.Assertion) Test {
	if delta < 0 {
		panic("delta must be positive")
	}

	t.now.Add(delta)
	t.tick(assertions)

	return t
}

func (t *test) AdvanceTimeTo(now time.Time, assertions ...assert.Assertion) Test {
	if now.Before(t.now) {
		panic("time must be greater than the current time")
	}

	t.now = now
	t.tick(assertions)

	return t
}

func (t *test) dispatch(m dogma.Message, assertions []assert.Assertion) {
	t.begin(assertions)

	opts := t.options(assertions)

	if err := t.engine.Dispatch(t.ctx, m, opts...); err != nil {
		t.t.Fatal(err)
	}

	t.end(assertions)
}

func (t *test) tick(assertions []assert.Assertion) {
	t.begin(assertions)

	opts := t.options(assertions)

	if err := t.engine.Tick(t.ctx, opts...); err != nil {
		t.t.Fatal(err)
	}

	t.end(assertions)
}

func (t *test) options(assertions []assert.Assertion) []engine.DispatchOption {
	opts := append(
		t.defaults,
		engine.WithCurrentTime(t.now),
	)

	for _, a := range assertions {
		opts = append(
			opts,
			engine.WithObserver(a),
		)
	}

	return opts
}

func (t *test) begin(assertions []assert.Assertion) {
	c := t.comparator
	if c == nil {
		c = compare.DefaultComparator{}
	}

	for _, a := range assertions {
		a.Begin(c)
	}
}

func (t *test) end(assertions []assert.Assertion) {
	r := t.renderer
	if r == nil {
		r = render.DefaultRenderer{}
	}

	buf := &strings.Builder{}
	pass := true

	for _, a := range assertions {
		if !a.End(buf, r) {
			pass = false
		}
	}

	if pass {
		t.t.Log(buf.String())
	} else {
		t.t.Fatal(buf.String())
	}
}
