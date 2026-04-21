package testkit_test

import (
	"context"
	"testing"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/enginekit/enginetest/stubs"
	. "github.com/dogmatiq/testkit"
	"github.com/dogmatiq/testkit/engine"
	"github.com/dogmatiq/testkit/internal/testingmock"
	"github.com/dogmatiq/testkit/internal/x/xtesting"
)

func TestToRecordEvent(t *testing.T) {
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
			c.Identity("<app>", "adb2ed1e-b1f4-4756-abfa-a5e3a3e08def")

			c.Routes(
				dogma.ViaAggregate(&AggregateMessageHandlerStub{
					ConfigureFunc: func(c dogma.AggregateConfigurer) {
						c.Identity("<aggregate>", "8746651e-df4d-421c-9eea-177585e5b8eb")
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
						c.Identity("<process>", "209c7f0f-49ad-4419-88a6-4e9ee1cf204a")
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
			"event recorded as expected",
			func(t *testing.T, tc *Test) Action {
				return ExecuteCommand(&CommandThatRecordsEvent{})
			},
			ToRecordEvent(&EventThatIsRecorded{}),
			true,
			expectReport(
				`✓ record a specific '*stubs.EventStub[TypeE]' event`,
			),
			nil,
		},
		{
			"no matching event recorded",
			func(t *testing.T, tc *Test) Action {
				return ExecuteCommand(&CommandThatRecordsEvent{})
			},
			ToRecordEvent(&EventThatIsNeverRecorded{}),
			false,
			expectReport(
				`✗ record a specific '*stubs.EventStub[TypeX]' event`,
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
			"no matching event recorded and all relevant handler types disabled",
			func(t *testing.T, tc *Test) Action {
				return ExecuteCommand(&CommandThatRecordsEvent{})
			},
			ToRecordEvent(&EventThatIsRecorded{}),
			false,
			expectReport(
				`✗ record a specific '*stubs.EventStub[TypeE]' event`,
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
			"no matching event recorded and no relevant handler types engaged",
			func(t *testing.T, tc *Test) Action {
				return ExecuteCommand(&CommandThatRecordsEvent{})
			},
			ToRecordEvent(&EventThatIsRecorded{}),
			false,
			expectReport(
				`✗ record a specific '*stubs.EventStub[TypeE]' event`,
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
			func(t *testing.T, tc *Test) Action {
				return ExecuteCommand(&CommandThatIsIgnored{})
			},
			ToRecordEvent(&EventThatIsRecorded{}),
			false,
			expectReport(
				`✗ record a specific '*stubs.EventStub[TypeE]' event`,
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
			func(t *testing.T, tc *Test) Action {
				return RecordEvent(&EventThatExecutesCommand{})
			},
			ToRecordEvent(&EventThatIsRecorded{}),
			false,
			expectReport(
				`✗ record a specific '*stubs.EventStub[TypeE]' event`,
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
			"similar event recorded with a different value",
			func(t *testing.T, tc *Test) Action {
				return ExecuteCommand(&CommandThatRecordsEvent{Content: "<content>"})
			},
			ToRecordEvent(&EventThatIsRecorded{Content: "<different>"}),
			false,
			expectReport(
				`✗ record a specific '*stubs.EventStub[TypeE]' event`,
				``,
				`  | EXPLANATION`,
				`  |     a similar event was recorded by the '<aggregate>' aggregate message handler`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • check the content of the message`,
				`  | `,
				`  | MESSAGE DIFF`,
				`  |     *stubs.EventStub[github.com/dogmatiq/enginekit/enginetest/stubs.TypeE]{`,
				`  |         Content:         "<[-differ-]{+cont+}ent>"`,
				`  |         ValidationError: ""`,
				`  |     }`,
			),
			nil,
		},
		{
			"does not include an explanation when negated and a sibling expectation passes",
			func(t *testing.T, tc *Test) Action {
				return ExecuteCommand(&CommandThatRecordsEvent{})
			},
			NoneOf(
				ToRecordEvent(&EventThatIsRecorded{}),
				ToRecordEvent(&EventThatIsNeverRecorded{}),
			),
			false,
			expectReport(
				`✗ none of (1 of the expectations passed unexpectedly)`,
				`    ✓ record a specific '*stubs.EventStub[TypeE]' event`,
				`    ✗ record a specific '*stubs.EventStub[TypeX]' event`,
			),
			nil,
		},
		{
			"fails the test if the message type is unrecognized",
			func(*testing.T, *Test) Action { return noop },
			ToRecordEvent(EventU1),
			false,
			expectReport(
				`✗ record a specific '*stubs.EventStub[TypeU]' event`,
				``,
				`  | EXPLANATION`,
				`  |     an event of type *stubs.EventStub[TypeU] can never be recorded, the application does not use this message type`,
			),
			nil,
		},
		{
			"fails the test if the message type is not produced by any handlers",
			func(*testing.T, *Test) Action { return noop },
			ToRecordEvent(&EventThatIsOnlyConsumed{}),
			false,
			expectReport(
				`✗ record a specific '*stubs.EventStub[TypeO]' event`,
				``,
				`  | EXPLANATION`,
				`  |     no handlers record events of type *stubs.EventStub[TypeO], it is only ever consumed`,
			),
			nil,
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			mt := &testingmock.T{FailSilently: true}
			tc := Begin(mt, app, c.Options...)
			tc.Expect(c.Action(t, tc), c.Expectation)
			c.Report(mt)
			if mt.Failed() != !c.Passes {
				t.Fatalf(
					"expectation should have %s but %s",
					map[bool]string{true: "passed", false: "failed"}[c.Passes],
					map[bool]string{true: "passed", false: "failed"}[mt.Failed()],
				)
			}
		})
	}
}

func TestToRecordEvent_NilMessage(t *testing.T) {
	xtesting.ExpectPanic(t, "ToRecordEvent(<nil>): message must not be nil", func() {
		ToRecordEvent(nil)
	})
}
