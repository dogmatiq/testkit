package engine_test

import (
	"context"
	"testing"
	"time"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/enginekit/config/runtimeconfig"
	. "github.com/dogmatiq/enginekit/enginetest/stubs"
	. "github.com/dogmatiq/testkit/engine"
	"github.com/dogmatiq/testkit/fact"
	"github.com/dogmatiq/testkit/x/xtesting"
)

func TestRun(t *testing.T) {
	newEngine := func() *Engine {
		app := &ApplicationStub{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "9e55f1ed-1f9a-46d9-a01f-e57638f74eb7")
			},
		}

		return MustNew(runtimeconfig.FromApplication(app))
	}

	t.Run("it calls tick repeatedly", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func() {
			time.Sleep(16 * time.Millisecond)
			cancel()
		}()

		buf := &fact.Buffer{}
		Run(ctx, newEngine(), 5*time.Millisecond, WithObserver(buf))

		facts := buf.Facts()
		if len(facts) < 6 {
			t.Fatalf("unexpected fact count: got %d, want >= 6", len(facts))
		}

		want := []any{
			fact.TickCycleBegun{},
			fact.TickCycleCompleted{},
			fact.TickCycleBegun{},
			fact.TickCycleCompleted{},
			fact.TickCycleBegun{},
			fact.TickCycleCompleted{},
		}

		for i, expected := range want {
			switch expected.(type) {
			case fact.TickCycleBegun:
				if _, ok := facts[i].(fact.TickCycleBegun); !ok {
					t.Fatalf("unexpected fact at %d: got %T, want fact.TickCycleBegun", i, facts[i])
				}
			case fact.TickCycleCompleted:
				if _, ok := facts[i].(fact.TickCycleCompleted); !ok {
					t.Fatalf("unexpected fact at %d: got %T, want fact.TickCycleCompleted", i, facts[i])
				}
			}
		}
	})

	t.Run("it returns an error if the context is canceled while ticking", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := Run(ctx, newEngine(), 0)
		xtesting.Expect(t, "unexpected error", err, context.Canceled)
	})

	t.Run("it returns an error if the context is canceled between ticks", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		go func() {
			time.Sleep(10 * time.Millisecond)
			cancel()
		}()

		err := Run(ctx, newEngine(), 0)
		xtesting.Expect(t, "unexpected error", err, context.Canceled)
	})
}

func TestRunTimeScaled(t *testing.T) {
	newEngine := func() *Engine {
		app := &ApplicationStub{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "4f06c58d-b854-41e9-92ee-d4e4ba137670")
			},
		}

		return MustNew(runtimeconfig.FromApplication(app))
	}

	t.Run("it scales time by the given factor", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func() {
			time.Sleep(30 * time.Millisecond)
			cancel()
		}()

		epoch := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

		buf := &fact.Buffer{}

		RunTimeScaled(
			ctx,
			newEngine(),
			10*time.Millisecond,
			0.5,
			epoch,
			WithObserver(buf),
		)

		var tickTimes []time.Time

		for _, f := range buf.Facts() {
			if begun, ok := f.(fact.TickCycleBegun); ok {
				tickTimes = append(tickTimes, begun.EngineTime)

				if len(tickTimes) == 3 {
					break
				}
			}
		}

		if len(tickTimes) != 3 {
			t.Fatalf("unexpected number of tick cycles: got %d, want at least 3", len(tickTimes))
		}

		if firstOffset := tickTimes[0].Sub(epoch); firstOffset < 0 || firstOffset >= 5*time.Millisecond {
			t.Fatalf("unexpected first tick offset: got %s, want in [0s, 5ms)", firstOffset)
		}

		const (
			minGap = 3 * time.Millisecond
			maxGap = 8 * time.Millisecond
		)

		for i, curr := range tickTimes[1:] {
			prev := tickTimes[i]

			if gap := curr.Sub(prev); gap < minGap || gap >= maxGap {
				t.Fatalf(
					"unexpected gap between tick %d and %d: got %s, want in [%s, %s)",
					i+1,
					i+2,
					gap,
					minGap,
					maxGap,
				)
			}
		}
	})
}
