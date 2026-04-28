package testkit_test

import (
	"errors"
	"testing"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/enginekit/enginetest/stubs"
	. "github.com/dogmatiq/testkit"
	"github.com/dogmatiq/testkit/internal/testingmock"
	"github.com/dogmatiq/testkit/internal/x/xtesting"
)

func TestToOnlyRecordEventsMatching(t *testing.T) {
	type (
		CommandThatRecordsEvent = CommandStub[TypeE]

		EventThatIsRecorded      = EventStub[TypeE]
		EventThatIsNeverRecorded = EventStub[TypeX]
		EventThatIsOnlyConsumed  = EventStub[TypeO]
	)

	app := &ApplicationStub{
		ConfigureFunc: func(c dogma.ApplicationConfigurer) {
			c.Identity("<app>", "94f425c5-339a-4213-8309-16234225480e")

			c.Routes(
				dogma.ViaAggregate(&AggregateMessageHandlerStub[*AggregateRootStub]{
					ConfigureFunc: func(c dogma.AggregateConfigurer) {
						c.Identity("<aggregate>", "bc64cfe4-3339-4eee-a9d2-364d33dff47d")
						c.Routes(
							dogma.HandlesCommand[*CommandThatRecordsEvent](),
							dogma.RecordsEvent[*EventThatIsRecorded](),
							dogma.RecordsEvent[*EventThatIsNeverRecorded](),
						)
					},
					RouteCommandToInstanceFunc: func(dogma.Command) string {
						return "<instance>"
					},
					HandleCommandFunc: func(
						_ *AggregateRootStub,
						s dogma.AggregateCommandScope[*AggregateRootStub],
						m dogma.Command,
					) {
						s.RecordEvent(EventE1)
						s.RecordEvent(EventE2)
						s.RecordEvent(EventE3)
					},
				}),

				dogma.ViaProjection(&ProjectionMessageHandlerStub{
					ConfigureFunc: func(c dogma.ProjectionConfigurer) {
						c.Identity("<projection>", "de708f1d-3651-437e-91ae-275a423ecd15")
						c.Routes(
							dogma.HandlesEvent[*EventThatIsOnlyConsumed](),
						)
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
			"all recorded events match",
			func(*testing.T, *Test) Action {
				return ExecuteCommand(&CommandThatRecordsEvent{})
			},
			ToOnlyRecordEventsMatching(
				func(m dogma.Event) error {
					return nil
				},
			),
			expectPass,
			expectReport(
				`✓ only record events that match the predicate near expectation.messagematch.eventonly_test.go:78`,
			),
			nil,
		},
		{
			"all recorded events match, using predicate with a more specific type",
			func(*testing.T, *Test) Action {
				return ExecuteCommand(&CommandThatRecordsEvent{})
			},
			ToOnlyRecordEventsMatching(
				func(*EventThatIsRecorded) error {
					return nil
				},
			),
			expectPass,
			expectReport(
				`✓ only record events that match the predicate near expectation.messagematch.eventonly_test.go:94`,
			),
			nil,
		},
		{
			"no events recorded at all",
			func(*testing.T, *Test) Action {
				return noop
			},
			ToOnlyRecordEventsMatching(
				func(m dogma.Event) error {
					panic("unexpected call")
				},
			),
			expectPass,
			expectReport(
				`✓ only record events that match the predicate near expectation.messagematch.eventonly_test.go:109`,
			),
			nil,
		},
		{
			"none of the recorded events match",
			func(*testing.T, *Test) Action {
				return ExecuteCommand(&CommandThatRecordsEvent{})
			},
			ToOnlyRecordEventsMatching(
				func(m dogma.Event) error {
					return errors.New("<error>")
				},
			),
			expectFail,
			expectReport(
				`✗ only record events that match the predicate near expectation.messagematch.eventonly_test.go:125`,
				``,
				`  | EXPLANATION`,
				`  |     none of the 3 relevant events matched the predicate`,
				`  | `,
				`  | FAILED MATCHES`,
				`  |     • *stubs.EventStub[TypeE]: <error> (repeated 3 times)`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • verify the logic within the predicate function`,
				`  |     • enable integration handlers using the EnableHandlerType() option`,
				`  |     • verify the logic within the '<aggregate>' aggregate message handler`,
			),
			nil,
		},
		{
			"some matching events recorded",
			func(*testing.T, *Test) Action {
				return ExecuteCommand(&CommandThatRecordsEvent{})
			},
			ToOnlyRecordEventsMatching(
				func(m dogma.Event) error {
					switch m {
					case EventE1:
						return errors.New("<error>")
					case EventE2:
						return IgnoreMessage
					default:
						return nil
					}
				},
			),
			expectFail,
			expectReport(
				`✗ only record events that match the predicate near expectation.messagematch.eventonly_test.go:152`,
				``,
				`  | EXPLANATION`,
				`  |     only 1 of 2 relevant events matched the predicate`,
				`  | `,
				`  | FAILED MATCHES`,
				`  |     • *stubs.EventStub[TypeE]: <error>`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • verify the logic within the predicate function, it ignored 1 event`,
				`  |     • enable integration handlers using the EnableHandlerType() option`,
				`  |     • verify the logic within the '<aggregate>' aggregate message handler`,
			),
			nil,
		},
		{
			"no matching events recorded, using predicate with a more specific type",
			func(*testing.T, *Test) Action {
				return ExecuteCommand(&CommandThatRecordsEvent{})
			},
			ToOnlyRecordEventsMatching(
				func(*EventThatIsNeverRecorded) error {
					panic("unexpected call")
				},
			),
			expectFail,
			expectReport(
				`✗ only record events that match the predicate near expectation.messagematch.eventonly_test.go:186`,
				``,
				`  | EXPLANATION`,
				`  |     none of the 3 relevant events matched the predicate`,
				`  | `,
				`  | FAILED MATCHES`,
				`  |     • *stubs.EventStub[TypeE]: predicate function expected *stubs.EventStub[TypeX] (repeated 3 times)`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • verify the logic within the predicate function`,
				`  |     • enable integration handlers using the EnableHandlerType() option`,
				`  |     • verify the logic within the '<aggregate>' aggregate message handler`,
			),
			nil,
		},
		{
			"passes the test if the message type is unrecognized",
			func(*testing.T, *Test) Action { return noop },
			ToOnlyRecordEventsMatching(
				func(*EventStub[TypeU]) error {
					return nil
				},
			),
			expectPass,
			expectReport(
				`✓ only record events that match the predicate near expectation.messagematch.eventonly_test.go:212`,
			),
			nil,
		},
		{
			"passes the test if the message type is not produced by any handlers",
			func(*testing.T, *Test) Action { return noop },
			ToOnlyRecordEventsMatching(
				func(*EventThatIsOnlyConsumed) error {
					return nil
				},
			),
			expectPass,
			expectReport(
				`✓ only record events that match the predicate near expectation.messagematch.eventonly_test.go:226`,
			),
			nil,
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			mt := &testingmock.T{FailSilently: true}
			tc := Begin(mt, app, c.Options...)
			tc.Expect(c.Action(t, tc), c.Expectation)

			if mt.Failed() != !c.Passes {
				t.Fatalf("testingT.Failed() = %v, want %v", mt.Failed(), !c.Passes)
			}

			preReportCount := len(mt.Logs)
			c.Report(mt)
			if len(mt.Logs) > preReportCount {
				t.Fatalf("report content mismatch:\n%v", mt.Logs[preReportCount:])
			}
		})
	}

	t.Run("panics if the function is nil", func(t *testing.T) {
		xtesting.ExpectPanic(
			t,
			"ToOnlyRecordEventsMatching(<nil>): function must not be nil",
			func() {
				var fn func(dogma.Event) error
				ToOnlyRecordEventsMatching(fn)
			},
		)
	})
}
