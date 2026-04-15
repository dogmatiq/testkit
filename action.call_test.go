package testkit_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/enginekit/config"
	. "github.com/dogmatiq/enginekit/enginetest/stubs"
	. "github.com/dogmatiq/testkit"
	"github.com/dogmatiq/testkit/engine"
	"github.com/dogmatiq/testkit/envelope"
	"github.com/dogmatiq/testkit/fact"
	"github.com/dogmatiq/testkit/internal/testingmock"
	"github.com/dogmatiq/testkit/internal/x/xtesting"
)

func TestCall(t *testing.T) {
	newFixture := func() (*testingmock.T, time.Time, *fact.Buffer, *Test) {
		app := &ApplicationStub{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "b51cde16-aaec-4d75-ae14-06282e3a72c8")
				c.Routes(
					dogma.ViaIntegration(&IntegrationMessageHandlerStub{
						ConfigureFunc: func(c dogma.IntegrationConfigurer) {
							c.Identity("<integration>", "832d78d7-a006-414f-b6d7-3153aa7c9ab8")
							c.Routes(
								dogma.HandlesCommand[*CommandStub[TypeA]](),
							)
						},
					}),
				)
			},
		}

		tm := &testingmock.T{}
		startTime := time.Now()
		buf := &fact.Buffer{}

		test := Begin(
			tm,
			app,
			StartTimeAt(startTime),
			WithUnsafeOperationOptions(
				engine.WithObserver(buf),
			),
		)

		return tm, startTime, buf, test
	}

	t.Run("it allows use of the test's executor", func(t *testing.T) {
		_, startTime, buf, tc := newFixture()
		e := tc.CommandExecutor()

		tc.Prepare(
			Call(func() {
				e.ExecuteCommand(
					context.Background(),
					CommandA1,
				)
			}),
		)

		xtesting.ExpectContains[fact.Fact](
			t,
			"expected dispatch cycle begin fact",
			buf.Facts(),
			fact.DispatchCycleBegun{
				Envelope: &envelope.Envelope{
					MessageID:     "1",
					CausationID:   "1",
					CorrelationID: "1",
					Message:       CommandA1,
					CreatedAt:     startTime,
				},
				EngineTime: startTime,
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

	t.Run("it allows expectations to match commands executed via the test's executor", func(t *testing.T) {
		_, _, _, tc := newFixture()
		e := tc.CommandExecutor()

		tc.Expect(
			Call(func() {
				e.ExecuteCommand(
					context.Background(),
					CommandA1,
				)
			}),
			ToExecuteCommand(CommandA1),
		)
	})

	t.Run("it produces the expected caption", func(t *testing.T) {
		tm, _, _, tc := newFixture()

		tc.Prepare(Call(func() {}))

		xtesting.ExpectContains(
			t,
			"expected caption",
			tm.Logs,
			"--- calling user-defined function ---",
		)
	})

	t.Run("it panics if the function is nil", func(t *testing.T) {
		xtesting.ExpectPanic(
			t,
			"Call(<nil>): function must not be nil",
			func() {
				Call(nil)
			},
		)
	})

	t.Run("it captures the location that the action was created", func(t *testing.T) {
		act := call(func() {})
		loc := act.Location()

		xtesting.Expect(t, "unexpected function name", loc.Func, "github.com/dogmatiq/testkit_test.call")
		if !strings.HasSuffix(loc.File, "/action.linenumber_test.go") {
			t.Fatalf("unexpected file: %s", loc.File)
		}
		xtesting.Expect(t, "unexpected line", loc.Line, 51)
	})
}
