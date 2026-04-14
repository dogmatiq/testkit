package testkit_test

import (
	"context"
	"testing"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/enginekit/enginetest/stubs"
	. "github.com/dogmatiq/testkit"
	"github.com/dogmatiq/testkit/internal/testingmock"
)

func TestToRecordEvent_WhenUsedWithCallAction(t *testing.T) {
	type (
		CommandThatIsIgnored    = CommandStub[TypeX]
		CommandThatRecordsEvent = CommandStub[TypeE]

		EventThatIsRecorded      = EventStub[TypeE]
		EventThatIsNeverRecorded = EventStub[TypeX]
	)

	app := &ApplicationStub{
		ConfigureFunc: func(c dogma.ApplicationConfigurer) {
			c.Identity("<app>", "58067ffe-c1d6-4097-8acb-c55a7936cb4b")

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
			)
		},
	}

	executeCommandViaExecutor := func(tb *testing.T, tc *Test, m dogma.Command) Action {
		tb.Helper()

		return Call(func() {
			err := tc.CommandExecutor().ExecuteCommand(context.Background(), m)
			if err != nil {
				tb.Fatalf("unexpected execute error: %v", err)
			}
		})
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
				return executeCommandViaExecutor(t, tc, &CommandThatRecordsEvent{})
			},
			ToRecordEvent(&EventThatIsRecorded{}),
			expectPass,
			expectReport(
				`✓ record a specific '*stubs.EventStub[TypeE]' event`,
			),
			nil,
		},
		{
			"no matching event recorded",
			func(t *testing.T, tc *Test) Action {
				return executeCommandViaExecutor(t, tc, &CommandThatRecordsEvent{})
			},
			ToRecordEvent(&EventThatIsNeverRecorded{}),
			expectFail,
			expectReport(
				`✗ record a specific '*stubs.EventStub[TypeX]' event`,
				``,
				`  | EXPLANATION`,
				`  |     nothing recorded a matching event`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • enable integration handlers using the EnableHandlerType() option`,
				`  |     • verify the logic within the '<aggregate>' aggregate message handler`,
				`  |     • verify the logic within the code that uses the dogma.EventRecorder`,
			),
			nil,
		},
		{
			"no messages produced at all",
			func(*testing.T, *Test) Action {
				return Call(func() {})
			},
			ToRecordEvent(&EventThatIsRecorded{}),
			expectFail,
			expectReport(
				`✗ record a specific '*stubs.EventStub[TypeE]' event`,
				``,
				`  | EXPLANATION`,
				`  |     no messages were produced at all`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • verify the logic within the code that uses the dogma.EventRecorder`,
			),
			nil,
		},
		{
			"no events recorded at all",
			func(t *testing.T, tc *Test) Action {
				return executeCommandViaExecutor(t, tc, &CommandThatIsIgnored{})
			},
			ToRecordEvent(&EventThatIsRecorded{}),
			expectFail,
			expectReport(
				`✗ record a specific '*stubs.EventStub[TypeE]' event`,
				``,
				`  | EXPLANATION`,
				`  |     no events were recorded at all`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • enable integration handlers using the EnableHandlerType() option`,
				`  |     • verify the logic within the '<aggregate>' aggregate message handler`,
				`  |     • verify the logic within the code that uses the dogma.EventRecorder`,
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
}
