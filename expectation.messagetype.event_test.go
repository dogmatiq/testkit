package testkit_test

import (
	"context"
	"testing"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/enginekit/enginetest/stubs"
	. "github.com/dogmatiq/testkit"
	"github.com/dogmatiq/testkit/engine"
	"github.com/dogmatiq/testkit/internal/testingmock"
)

func TestToRecordEventType(t *testing.T) {
	type (
		CommandThatIsIgnored    = CommandStub[TypeX]
		CommandThatRecordsEvent = CommandStub[TypeE]

		EventThatIsRecorded      = EventStub[TypeE]
		EventThatIsNeverRecorded = EventStub[TypeX]
		EventThatExecutesCommand = EventStub[TypeC]
		EventThatIsOnlyConsumed  = EventStub[TypeO]
	)

	app := &ApplicationStub{
		ConfigureFunc: func(c dogma.ApplicationConfigurer) {
			c.Identity("<app>", "ef25ca55-2ace-40b5-9c2d-c53f5a80908a")

			c.Routes(
				dogma.ViaAggregate(&AggregateMessageHandlerStub{
					ConfigureFunc: func(c dogma.AggregateConfigurer) {
						c.Identity("<aggregate>", "2cc50fd0-3d22-4f96-81c6-5e28d6abe735")
						c.Routes(
							dogma.HandlesCommand[*CommandThatIsIgnored](),

							dogma.HandlesCommand[*CommandThatRecordsEvent](),
							dogma.RecordsEvent[*EventThatIsRecorded](),
							dogma.RecordsEvent[*EventThatIsNeverRecorded](),
						)
					},
					RouteCommandToInstanceFunc: func(dogma.Command) string {
						return "<instance>"
					},
					HandleCommandFunc: func(
						_ dogma.AggregateRoot,
						s dogma.AggregateCommandScope,
						m dogma.Command,
					) {
						switch m := m.(type) {
						case *CommandThatRecordsEvent:
							s.RecordEvent(&EventThatIsRecorded{
								Content: m.Content,
							})
						}
					},
				}),

				dogma.ViaProcess(&ProcessMessageHandlerStub{
					ConfigureFunc: func(c dogma.ProcessConfigurer) {
						c.Identity("<process>", "94de2bb7-c115-494d-ad15-bdfedbe4aec3")
						c.Routes(
							dogma.HandlesEvent[*EventThatExecutesCommand](),
							dogma.HandlesEvent[*EventThatIsOnlyConsumed](),
							dogma.ExecutesCommand[*CommandThatIsIgnored](),
						)
					},
					RouteEventToInstanceFunc: func(
						context.Context,
						dogma.Event,
					) (string, bool, error) {
						return "<instance>", true, nil
					},
					HandleEventFunc: func(
						_ context.Context,
						_ dogma.ProcessRoot,
						s dogma.ProcessEventScope,
						m dogma.Event,
					) error {
						switch m.(type) {
						case *EventThatExecutesCommand:
							s.ExecuteCommand(&CommandThatIsIgnored{})
						}
						return nil
					},
				}),
			)
		},
	}

	cases := []struct {
		Name        string
		Action      func(*testing.T, *Test) Action
		Expectation Expectation
		Passes      bool
		Report      reportMatcher
		Options     []TestOption
	}{
		{
			"event type recorded as expected",
			func(*testing.T, *Test) Action {
				return ExecuteCommand(&CommandThatRecordsEvent{})
			},
			ToRecordEventType[*EventThatIsRecorded](),
			expectPass,
			expectReport(
				`✓ record any '*stubs.EventStub[TypeE]' event`,
			),
			nil,
		},
		{
			"no matching event type recorded",
			func(*testing.T, *Test) Action {
				return ExecuteCommand(&CommandThatRecordsEvent{})
			},
			ToRecordEventType[*EventThatIsNeverRecorded](),
			expectFail,
			expectReport(
				`✗ record any '*stubs.EventStub[TypeX]' event`,
				``,
				`  | EXPLANATION`,
				`  |     none of the engaged handlers recorded a matching event`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • enable integration handlers using the EnableHandlerType() option`,
				`  |     • verify the logic within the '<aggregate>' aggregate message handler`,
			),
			nil,
		},
		{
			"no matching event type recorded and all relevant handler types disabled",
			func(*testing.T, *Test) Action {
				return ExecuteCommand(&CommandThatRecordsEvent{})
			},
			ToRecordEventType[*EventThatIsNeverRecorded](),
			expectFail,
			expectReport(
				`✗ record any '*stubs.EventStub[TypeX]' event`,
				``,
				`  | EXPLANATION`,
				`  |     no relevant handler types were enabled`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • enable aggregate handlers using the EnableHandlerType() option`,
				`  |     • enable integration handlers using the EnableHandlerType() option`,
			),
			[]TestOption{
				WithUnsafeOperationOptions(
					engine.EnableAggregates(false),
					engine.EnableIntegrations(false),
				),
			},
		},
		{
			"no matching event type recorded and no relevant handler types engaged",
			func(*testing.T, *Test) Action {
				return ExecuteCommand(&CommandThatRecordsEvent{})
			},
			ToRecordEventType[*EventThatIsNeverRecorded](),
			expectFail,
			expectReport(
				`✗ record any '*stubs.EventStub[TypeX]' event`,
				``,
				`  | EXPLANATION`,
				`  |     no relevant handlers (aggregate or integration) were engaged`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • enable aggregate handlers using the EnableHandlerType() option`,
				`  |     • check the application's routing configuration`,
			),
			[]TestOption{
				WithUnsafeOperationOptions(
					engine.EnableAggregates(false),
					engine.EnableIntegrations(true),
				),
			},
		},
		{
			"no messages produced at all",
			func(*testing.T, *Test) Action {
				return ExecuteCommand(&CommandThatIsIgnored{})
			},
			ToRecordEventType[*EventThatIsNeverRecorded](),
			expectFail,
			expectReport(
				`✗ record any '*stubs.EventStub[TypeX]' event`,
				``,
				`  | EXPLANATION`,
				`  |     no messages were produced at all`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • enable integration handlers using the EnableHandlerType() option`,
				`  |     • verify the logic within the '<aggregate>' aggregate message handler`,
			),
			nil,
		},
		{
			"no events recorded at all",
			func(*testing.T, *Test) Action {
				return RecordEvent(&EventThatExecutesCommand{})
			},
			ToRecordEventType[*EventThatIsRecorded](),
			expectFail,
			expectReport(
				`✗ record any '*stubs.EventStub[TypeE]' event`,
				``,
				`  | EXPLANATION`,
				`  |     no events were recorded at all`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • enable integration handlers using the EnableHandlerType() option`,
				`  |     • verify the logic within the '<aggregate>' aggregate message handler`,
			),
			nil,
		},
		{
			"does not include an explanation when negated and a sibling expectation passes",
			func(*testing.T, *Test) Action {
				return ExecuteCommand(&CommandThatRecordsEvent{})
			},
			NoneOf(
				ToRecordEventType[*EventThatIsRecorded](),
				ToRecordEventType[*EventThatIsNeverRecorded](),
			),
			expectFail,
			expectReport(
				`✗ none of (1 of the expectations passed unexpectedly)`,
				`    ✓ record any '*stubs.EventStub[TypeE]' event`,
				`    ✗ record any '*stubs.EventStub[TypeX]' event`,
			),
			nil,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			mt := &testingmock.T{FailSilently: true}
			tc := Begin(mt, app, c.Options...)
			tc.Expect(c.Action(t, tc), c.Expectation)
			c.Report(mt)

			if mt.Failed() != !c.Passes {
				t.Fatalf("testingT.Failed() = %v, want %v", mt.Failed(), !c.Passes)
			}
		})
	}

	t.Run("fails the test if the message type is unrecognized", func(t *testing.T) {
		mt := &testingmock.T{FailSilently: true}
		tc := Begin(mt, app)
		tc.Expect(
			noop,
			ToRecordEventType[*EventStub[TypeU]](),
		)

		if !mt.Failed() {
			t.Fatal("expected test to fail")
		}
		if !containsString(mt.Logs, "an event of type *stubs.EventStub[TypeU] can never be recorded, the application does not use this message type") {
			t.Fatalf("expected log message not found, got: %v", mt.Logs)
		}
	})

	t.Run("fails the test if the message type is not produced by any handlers", func(t *testing.T) {
		mt := &testingmock.T{FailSilently: true}
		tc := Begin(mt, app)
		tc.Expect(
			noop,
			ToRecordEventType[*EventThatIsOnlyConsumed](),
		)

		if !mt.Failed() {
			t.Fatal("expected test to fail")
		}
		if !containsString(mt.Logs, "no handlers record events of type *stubs.EventStub[TypeO], it is only ever consumed") {
			t.Fatalf("expected log message not found, got: %v", mt.Logs)
		}
	})
}
