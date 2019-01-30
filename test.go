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
	"github.com/dogmatiq/dogmatest/engine/fact"
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
		t.dispatch(m)
	}

	return t
}

func (t *test) ExecuteCommand(m dogma.Message, a assert.Assertion) Test {
	t.begin(a)
	t.dispatch(m, a) // TODO: fail if TypeOf(m)'s role is not correct
	t.end(a)

	return t
}

func (t *test) RecordEvent(m dogma.Message, a assert.Assertion) Test {
	t.begin(a)
	t.dispatch(m, a) // TODO: fail if TypeOf(m)'s role is not correct
	t.end(a)

	return t
}

func (t *test) AdvanceTimeBy(delta time.Duration, a assert.Assertion) Test {
	if delta < 0 {
		panic("delta must be positive")
	}

	return t.AdvanceTimeTo(
		t.now.Add(delta),
		a,
	)
}

func (t *test) AdvanceTimeTo(now time.Time, a assert.Assertion) Test {
	if now.Before(t.now) {
		panic("time must be greater than the current time")
	}

	t.now = now

	t.begin(a)
	t.tick(a)
	t.end(a)

	return t
}

func (t *test) dispatch(m dogma.Message, observers ...fact.Observer) {
	opts := t.options(observers)
	if err := t.engine.Dispatch(t.ctx, m, opts...); err != nil {
		t.t.Fatal(err)
	}
}

func (t *test) tick(observers ...fact.Observer) {
	opts := t.options(observers)
	if err := t.engine.Tick(t.ctx, opts...); err != nil {
		t.t.Fatal(err)
	}
}

func (t *test) options(observers []fact.Observer) []engine.DispatchOption {
	opts := append(
		t.defaults,
		engine.WithCurrentTime(t.now),
	)

	for _, obs := range observers {
		opts = append(
			opts,
			engine.WithObserver(obs),
		)
	}

	return opts
}

func (t *test) begin(a assert.Assertion) {
	if a == nil {
		panic("assertion must not be nil")
	}

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

	res := a.End(r)
	iago.MustWriteTo(buf, res)

	if res.Ok {
		t.t.Log(buf.String())
	} else {
		t.t.Fatal(buf.String())
	}
}
