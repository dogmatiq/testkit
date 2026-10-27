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

func TestRecordEvent(t *testing.T) {
	newFixture := func() (*testingmock.T, time.Time, *fact.Buffer, *Test) {
		app := &ApplicationStub{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "38408e83-e8eb-4f82-abe1-7fa02cee0657")
				c.Routes(
					dogma.ViaProcess(&ProcessMessageHandlerStub[*ProcessRootStub]{
						ConfigureFunc: func(c dogma.ProcessConfigurer) {
							c.Identity("<process>", "1c0dd111-fe12-4dee-a8bc-64abea1dce8f")
							c.Routes(
								dogma.HandlesEvent[*EventStub[TypeA]](),
								dogma.ExecutesCommand[*CommandStub[TypeA]](),
							)
						},
						RouteEventToInstanceFunc: func(context.Context, dogma.Event) (string, bool, error) {
							return "<instance>", true, nil
						},
					}),
					dogma.ViaIntegration(&IntegrationMessageHandlerStub{
						ConfigureFunc: func(c dogma.IntegrationConfigurer) {
							c.Identity("<integration>", "9e4d2b7a-6f1c-4a3e-b8d5-2c7f0e1a9b3d")
							c.Routes(
								dogma.HandlesCommand[*CommandStub[TypeA]](),
								dogma.RecordsEvent[*EventStub[TypeB]](),
							)
						},
					}),
				)
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

	t.Run("it dispatches the message", func(t *testing.T) {
		_, startTime, buf, tc := newFixture()

		tc.Prepare(RecordEvent(EventA1))

		xtesting.ExpectContains[fact.Fact](
			t,
			"expected dispatch cycle begin fact",
			buf.Facts(),
			fact.DispatchCycleBegun{
				Envelope: &envelope.Envelope{
					MessageID:         "1",
					CausationID:       "1",
					CorrelationID:     "1",
					Message:           EventA1,
					CreatedAt:         startTime,
					EventStreamID:     "ef0750de-15cd-5d0c-932e-adee5e8ebf47",
					EventStreamOffset: 0,
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
			Message dogma.Event
			Error   string
		}{
			{
				Name:    "unrouted type",
				Message: EventX1,
				Error:   "no handlers within the application have routes for the *stubs.EventStub[TypeX] event",
			},
			{
				Name:    "type that is only produced",
				Message: EventB1,
				Error:   "no handlers within the application have inbound routes for the *stubs.EventStub[TypeB] event, it is only ever recorded",
			},
		}

		for _, c := range cases {
			t.Run(c.Name, func(t *testing.T) {
				tm, _, _, tc := newFixture()
				tm.FailSilently = true

				tc.Prepare(RecordEvent(c.Message))

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
				Expectation: ToExecuteCommand(CommandA1),
				Report: []string{
					`✗ execute a specific '*stubs.CommandStub[TypeA]' command`,
					``,
					`  | EXPLANATION`,
					`  |     no handlers within the application have routes for the *stubs.EventStub[TypeX] event`,
					`  | `,
					`  | SUGGESTIONS`,
					`  |     • add a HandlesEvent[*stubs.EventStub[TypeX]]() route to each process or projection that should handle this event`,
				},
			},
			{
				Name:        "negated expectation",
				Expectation: Not(ToExecuteCommand(CommandA1)),
				Report: []string{
					`✓ do not execute a specific '*stubs.CommandStub[TypeA]' command`,
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

				tc.Expect(RecordEvent(EventX1), c.Expectation)

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
					"no handlers within the application have routes for the *stubs.EventStub[TypeX] event",
				)
				expectReport(c.Report...)(t, &testingmock.T{Logs: tm.Logs[:n-1]})
			})
		}
	})

	t.Run("it does not satisfy its own expectations", func(t *testing.T) {
		tm, _, _, tc := newFixture()
		tm.FailSilently = true

		tc.Expect(
			RecordEvent(EventA1),
			ToRecordEvent(EventA1),
		)

		if !tm.Failed() {
			t.Fatal("expected test to fail")
		}
	})

	t.Run("it produces the expected caption", func(t *testing.T) {
		tm, _, _, tc := newFixture()

		tc.Prepare(RecordEvent(EventA1))

		xtesting.ExpectContains(
			t,
			"expected caption",
			tm.Logs,
			"--- recording *stubs.EventStub[TypeA] event ---",
		)
	})

	t.Run("it panics if the message is nil", func(t *testing.T) {
		xtesting.ExpectPanic(
			t,
			"RecordEvent(<nil>): message must not be nil",
			func() {
				RecordEvent(nil)
			},
		)
	})

	t.Run("it panics if the message is invalid", func(t *testing.T) {
		xtesting.ExpectPanic(
			t,
			"RecordEvent(*stubs.EventStub[TypeA]): <invalid>",
			func() {
				RecordEvent(&EventStub[TypeA]{ValidationError: "<invalid>"})
			},
		)
	})

	t.Run("it captures the location that the action was created", func(t *testing.T) {
		act := recordEvent(EventA1)
		loc := act.Location()

		xtesting.Expect(t, "unexpected function name", loc.Func, "github.com/dogmatiq/testkit_test.recordEvent")
		if !strings.HasSuffix(loc.File, "/action.linenumber_test.go") {
			t.Fatalf("unexpected file: %s", loc.File)
		}
		xtesting.Expect(t, "unexpected line", loc.Line, 53)
	})
}
