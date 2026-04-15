package testkit_test

import (
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

func TestExecuteCommand(t *testing.T) {
	newFixture := func() (*testingmock.T, time.Time, *fact.Buffer, *Test) {
		app := &ApplicationStub{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "a84b2620-4675-4024-b55b-cd5dbeb6e293")
				c.Routes(
					dogma.ViaIntegration(&IntegrationMessageHandlerStub{
						ConfigureFunc: func(c dogma.IntegrationConfigurer) {
							c.Identity("<integration>", "d1cf3af1-6c20-4125-8e68-192a6075d0b4")
							c.Routes(
								dogma.HandlesCommand[*CommandStub[TypeA]](),
								dogma.RecordsEvent[*EventStub[TypeA]](),
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

	t.Run("it dispatches the message", func(t *testing.T) {
		_, startTime, buf, tc := newFixture()

		tc.Prepare(ExecuteCommand(CommandA1))

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

	t.Run("it fails the test if the message type is unrecognized", func(t *testing.T) {
		tm, _, _, tc := newFixture()
		tm.FailSilently = true

		tc.Prepare(ExecuteCommand(CommandX1))

		if !tm.Failed() {
			t.Fatal("expected test to fail")
		}
		xtesting.ExpectContains(
			t,
			"expected error log",
			tm.Logs,
			"cannot execute command, *stubs.CommandStub[TypeX] is a not a recognized message type",
		)
	})

	t.Run("it does not satisfy its own expectations", func(t *testing.T) {
		tm, _, _, tc := newFixture()
		tm.FailSilently = true

		tc.Expect(
			ExecuteCommand(CommandA1),
			ToExecuteCommand(CommandA1),
		)

		if !tm.Failed() {
			t.Fatal("expected test to fail")
		}
	})

	t.Run("it produces the expected caption", func(t *testing.T) {
		tm, _, _, tc := newFixture()

		tc.Prepare(ExecuteCommand(CommandA1))

		xtesting.ExpectContains(
			t,
			"expected caption",
			tm.Logs,
			"--- executing *stubs.CommandStub[TypeA] command ---",
		)
	})

	t.Run("it panics if the message is nil", func(t *testing.T) {
		xtesting.ExpectPanic(
			t,
			"ExecuteCommand(<nil>): message must not be nil",
			func() {
				ExecuteCommand(nil)
			},
		)
	})

	t.Run("it captures the location that the action was created", func(t *testing.T) {
		act := executeCommand(CommandA1)
		loc := act.Location()

		xtesting.Expect(t, "unexpected function name", loc.Func, "github.com/dogmatiq/testkit_test.executeCommand")
		if !strings.HasSuffix(loc.File, "/action.linenumber_test.go") {
			t.Fatalf("unexpected file: %s", loc.File)
		}
		xtesting.Expect(t, "unexpected line", loc.Line, 52)
	})
}
