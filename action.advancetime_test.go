package testkit_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/enginekit/config"
	. "github.com/dogmatiq/enginekit/enginetest/stubs"
	. "github.com/dogmatiq/testkit"
	"github.com/dogmatiq/testkit/engine"
	"github.com/dogmatiq/testkit/fact"
	"github.com/dogmatiq/testkit/internal/testingmock"
	"github.com/dogmatiq/testkit/internal/x/xtesting"
)

func TestAdvanceTime(t *testing.T) {
	newFixture := func() (*testingmock.T, time.Time, *fact.Buffer, *Test) {
		app := &ApplicationStub{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "140ca29b-7a05-4f26-968b-6285255e6d8a")
			},
		}

		tm := &testingmock.T{}
		startTime := time.Now()
		buf := &fact.Buffer{}

		tc := Begin(
			tm,
			app,
			StartTimeAt(startTime),
			WithUnsafeOperationOptions(
				engine.WithObserver(buf),
			),
		)

		return tm, startTime, buf, tc
	}

	t.Run("it retains the virtual time between calls", func(t *testing.T) {
		_, startTime, buf, tc := newFixture()

		tc.Prepare(
			AdvanceTime(ByDuration(1*time.Second)),
			AdvanceTime(ByDuration(1*time.Second)),
		)

		xtesting.ExpectContains[fact.Fact](
			t,
			"expected tick fact",
			buf.Facts(),
			fact.TickCycleBegun{
				EngineTime: startTime.Add(2 * time.Second),
				EnabledHandlerTypes: map[config.HandlerType]bool{
					config.AggregateHandlerType:   true,
					config.IntegrationHandlerType: false,
					config.ProcessHandlerType:     true,
					config.ProjectionHandlerType:  false,
				},
				EnabledHandlers: map[string]bool{},
			},
		)
	})

	t.Run("it fails the test if time is reversed", func(t *testing.T) {
		tm, startTime, _, tc := newFixture()
		tm.FailSilently = true

		target := startTime.Add(-1 * time.Second)

		tc.Prepare(AdvanceTime(ToTime(target)))

		if !tm.Failed() {
			t.Fatal("expected test to fail")
		}
		xtesting.ExpectContains(
			t,
			"expected failure log",
			tm.Logs,
			fmt.Sprintf("adjusting the clock to %s would reverse time", target.Format(time.RFC3339)),
		)
	})

	t.Run("it panics if the adjustment is nil", func(t *testing.T) {
		xtesting.ExpectPanic(
			t,
			"AdvanceTime(<nil>): adjustment must not be nil",
			func() {
				AdvanceTime(nil)
			},
		)
	})

	t.Run("it captures the location that the action was created", func(t *testing.T) {
		act := advanceTime(ByDuration(10 * time.Second))
		loc := act.Location()

		xtesting.Expect(t, "unexpected function name", loc.Func, "github.com/dogmatiq/testkit_test.advanceTime")
		if !strings.HasSuffix(loc.File, "/action.linenumber_test.go") {
			t.Fatalf("unexpected file: %s", loc.File)
		}
		xtesting.Expect(t, "unexpected line", loc.Line, 50)
	})

	t.Run("passed a ToTime() adjustment", func(t *testing.T) {
		targetTime := time.Date(2100, 1, 2, 3, 4, 5, 6, time.UTC)

		t.Run("it advances the clock to the provided time", func(t *testing.T) {
			_, _, buf, tc := newFixture()

			tc.Prepare(AdvanceTime(ToTime(targetTime)))

			xtesting.ExpectContains[fact.Fact](
				t,
				"expected tick fact",
				buf.Facts(),
				fact.TickCycleBegun{
					EngineTime: targetTime,
					EnabledHandlerTypes: map[config.HandlerType]bool{
						config.AggregateHandlerType:   true,
						config.IntegrationHandlerType: false,
						config.ProcessHandlerType:     true,
						config.ProjectionHandlerType:  false,
					},
					EnabledHandlers: map[string]bool{},
				},
			)
		})

		t.Run("it produces the expected caption", func(t *testing.T) {
			tm, _, _, tc := newFixture()

			tc.Prepare(AdvanceTime(ToTime(targetTime)))

			xtesting.ExpectContains(
				t,
				"expected caption",
				tm.Logs,
				"--- advancing time to 2100-01-02T03:04:05Z ---",
			)
		})
	})

	t.Run("passed a ByDuration() adjustment", func(t *testing.T) {
		t.Run("it advances the clock then performs a tick", func(t *testing.T) {
			_, startTime, buf, tc := newFixture()

			tc.Prepare(AdvanceTime(ByDuration(3 * time.Second)))

			xtesting.ExpectContains[fact.Fact](
				t,
				"expected tick fact",
				buf.Facts(),
				fact.TickCycleBegun{
					EngineTime: startTime.Add(3 * time.Second),
					EnabledHandlerTypes: map[config.HandlerType]bool{
						config.AggregateHandlerType:   true,
						config.IntegrationHandlerType: false,
						config.ProcessHandlerType:     true,
						config.ProjectionHandlerType:  false,
					},
					EnabledHandlers: map[string]bool{},
				},
			)
		})

		t.Run("it produces the expected caption", func(t *testing.T) {
			tm, _, _, tc := newFixture()

			tc.Prepare(AdvanceTime(ByDuration(3 * time.Second)))

			xtesting.ExpectContains(
				t,
				"expected caption",
				tm.Logs,
				"--- advancing time by 3s ---",
			)
		})

		t.Run("it panics if the duration is negative", func(t *testing.T) {
			xtesting.ExpectPanic(
				t,
				"ByDuration(-1s): duration must not be negative",
				func() {
					ByDuration(-1 * time.Second)
				},
			)
		})
	})
}
