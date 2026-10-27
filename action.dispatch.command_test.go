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
					dogma.ViaProcess(&ProcessMessageHandlerStub[*ProcessRootStub]{
						ConfigureFunc: func(c dogma.ProcessConfigurer) {
							c.Identity("<process>", "5b7f8a1e-3c2d-4e9f-8a6b-1d2c3e4f5a6b")
							c.Routes(
								dogma.HandlesEvent[*EventStub[TypeA]](),
								dogma.ExecutesCommand[*CommandStub[TypeB]](),
							)
						},
						RouteEventToInstanceFunc: func(context.Context, dogma.Event) (string, bool, error) {
							return "<instance>", true, nil
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

	t.Run("it fails the test if the message type is not handled", func(t *testing.T) {
		cases := []struct {
			Name    string
			Message dogma.Command
			Error   string
		}{
			{
				Name:    "unrouted type",
				Message: CommandX1,
				Error:   "no handlers within the application have routes for the *stubs.CommandStub[TypeX] command",
			},
			{
				Name:    "type that is only produced",
				Message: CommandB1,
				Error:   "no handlers within the application have inbound routes for the *stubs.CommandStub[TypeB] command, it is only ever executed",
			},
		}

		for _, c := range cases {
			t.Run(c.Name, func(t *testing.T) {
				tm, _, _, tc := newFixture()
				tm.FailSilently = true

				tc.Prepare(ExecuteCommand(c.Message))

				if !tm.Failed() {
					t.Fatal("expected test to fail")
				}
				xtesting.ExpectContains(t, "expected error log", tm.Logs, c.Error)
			})
		}
	})

	t.Run("it fails the test if the message type is not handled, regardless of the expectation", func(t *testing.T) {
		cases := []struct {
			Name        string
			Expectation Expectation
			Report      []string
		}{
			{
				Name:        "expectation that would fail",
				Expectation: ToRecordEvent(EventA1),
				Report: []string{
					`✗ record a specific '*stubs.EventStub[TypeA]' event`,
					``,
					`  | EXPLANATION`,
					`  |     no handlers within the application have routes for the *stubs.CommandStub[TypeX] command`,
					`  | `,
					`  | SUGGESTIONS`,
					`  |     • add a HandlesCommand[*stubs.CommandStub[TypeX]]() route to the aggregate or integration that should handle this command`,
				},
			},
			{
				Name:        "negated expectation",
				Expectation: Not(ToRecordEvent(EventA1)),
				Report: []string{
					`✓ do not record a specific '*stubs.EventStub[TypeA]' event`,
				},
			},
			{
				Name:        "expectation that always passes",
				Expectation: pass,
				Report: []string{
					`✓ <always pass>`,
				},
			},
		}

		for _, c := range cases {
			t.Run(c.Name, func(t *testing.T) {
				tm, _, _, tc := newFixture()
				tm.FailSilently = true

				tc.Expect(ExecuteCommand(CommandX1), c.Expectation)

				if !tm.Failed() {
					t.Fatal("expected test to fail")
				}

				// The report is logged first, followed by the error that failed
				// the test.
				n := len(tm.Logs)
				xtesting.Expect(
					t,
					"unexpected error log",
					tm.Logs[n-1],
					"no handlers within the application have routes for the *stubs.CommandStub[TypeX] command",
				)
				expectReport(c.Report...)(t, &testingmock.T{Logs: tm.Logs[:n-1]})
			})
		}
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

	t.Run("it panics if the message is invalid", func(t *testing.T) {
		xtesting.ExpectPanic(
			t,
			"ExecuteCommand(*stubs.CommandStub[TypeA]): <invalid>",
			func() {
				ExecuteCommand(&CommandStub[TypeA]{ValidationError: "<invalid>"})
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
