package testkit_test

import (
	"context"
	"errors"
	"testing"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/enginekit/enginetest/stubs"
	. "github.com/dogmatiq/testkit"
	"github.com/dogmatiq/testkit/engine"
	"github.com/dogmatiq/testkit/internal/testingmock"
	"github.com/dogmatiq/testkit/x/xtesting"
)

func TestToRecordEventMatching(t *testing.T) {
	type (
		CommandThatIsIgnored    = CommandStub[TypeX]
		CommandThatRecordsEvent = CommandStub[TypeE]

		EventThatIsRecorded      = EventStub[TypeE]
		EventThatIsNeverRecorded = EventStub[TypeX]
		EventThatExecutesCommand = EventStub[TypeC]
	)

	app := &ApplicationStub{
		ConfigureFunc: func(c dogma.ApplicationConfigurer) {
			c.Identity("<app>", "43962b88-b25c-4a59-938e-64540c473a7c")

			c.Routes(
				dogma.ViaAggregate(&AggregateMessageHandlerStub{
					ConfigureFunc: func(c dogma.AggregateConfigurer) {
						c.Identity("<aggregate>", "6a182af9-4659-49cf-8c44-85dfd47dbefc")
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
							s.RecordEvent(
								&EventThatIsRecorded{
									Content: m.Content,
								},
							)

							if m.Content == "<multiple>" {
								s.RecordEvent(
									&EventThatIsRecorded{
										Content: m.Content,
									},
								)
							}
						}
					},
				}),

				dogma.ViaProcess(&ProcessMessageHandlerStub{
					ConfigureFunc: func(c dogma.ProcessConfigurer) {
						c.Identity("<process>", "7651ad19-1526-48c0-a53d-676286a34ca6")
						c.Routes(
							dogma.HandlesEvent[*EventThatExecutesCommand](),
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
			"matching event recorded as expected",
			func(*testing.T, *Test) Action {
				return ExecuteCommand(&CommandThatRecordsEvent{})
			},
			ToRecordEventMatching(
				func(m dogma.Event) error {
					if m, ok := m.(*EventThatIsRecorded); ok && *m == (EventThatIsRecorded{}) {
						return nil
					}

					return errors.New("<error>")
				},
			),
			expectPass,
			expectReport(
				`✓ record an event that matches the predicate near expectation.messagematch.event_test.go:114`,
			),
			nil,
		},
		{
			"matching event recorded as expected, using predicate with a more specific type",
			func(*testing.T, *Test) Action {
				return ExecuteCommand(&CommandThatRecordsEvent{})
			},
			ToRecordEventMatching(
				func(m *EventThatIsRecorded) error {
					if *m == (EventThatIsRecorded{}) {
						return nil
					}

					return errors.New("<error>")
				},
			),
			expectPass,
			expectReport(
				`✓ record an event that matches the predicate near expectation.messagematch.event_test.go:134`,
			),
			nil,
		},
		{
			"no matching event recorded",
			func(*testing.T, *Test) Action {
				return ExecuteCommand(&CommandThatRecordsEvent{})
			},
			ToRecordEventMatching(
				func(m dogma.Event) error {
					return errors.New("<error>")
				},
			),
			expectFail,
			expectReport(
				`✗ record an event that matches the predicate near expectation.messagematch.event_test.go:154`,
				``,
				`  | EXPLANATION`,
				`  |     none of the engaged handlers recorded a matching event`,
				`  | `,
				`  | FAILED MATCHES`,
				`  |     • *stubs.EventStub[TypeE]: <error>`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • verify the logic within the predicate function`,
				`  |     • enable integration handlers using the EnableHandlerType() option`,
				`  |     • verify the logic within the '<aggregate>' aggregate message handler`,
			),
			nil,
		},
		{
			"no matching event recorded, using predicate with a more specific type",
			func(*testing.T, *Test) Action {
				return ExecuteCommand(&CommandThatRecordsEvent{})
			},
			ToRecordEventMatching(
				func(*EventThatIsNeverRecorded) error {
					panic("unexpected call")
				},
			),
			expectFail,
			expectReport(
				`✗ record an event that matches the predicate near expectation.messagematch.event_test.go:181`,
				``,
				`  | EXPLANATION`,
				`  |     none of the engaged handlers recorded a matching event`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • verify the logic within the predicate function, it ignored 1 event`,
				`  |     • enable integration handlers using the EnableHandlerType() option`,
				`  |     • verify the logic within the '<aggregate>' aggregate message handler`,
			),
			nil,
		},
		{
			"no messages produced at all",
			func(*testing.T, *Test) Action {
				return ExecuteCommand(&CommandThatIsIgnored{})
			},
			ToRecordEventMatching(
				func(m dogma.Event) error {
					panic("unexpected call")
				},
			),
			expectFail,
			expectReport(
				`✗ record an event that matches the predicate near expectation.messagematch.event_test.go:205`,
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
			ToRecordEventMatching(
				func(dogma.Event) error {
					panic("unexpected call")
				},
			),
			expectFail,
			expectReport(
				`✗ record an event that matches the predicate near expectation.messagematch.event_test.go:228`,
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
			"no matching event recorded and all relevant handler types disabled",
			func(*testing.T, *Test) Action {
				return ExecuteCommand(&CommandThatRecordsEvent{})
			},
			ToRecordEventMatching(
				func(m dogma.Event) error {
					panic("unexpected call")
				},
			),
			expectFail,
			expectReport(
				`✗ record an event that matches the predicate near expectation.messagematch.event_test.go:251`,
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
			"no matching event recorded and events were ignored",
			func(*testing.T, *Test) Action {
				return ExecuteCommand(&CommandThatRecordsEvent{})
			},
			ToRecordEventMatching(
				func(m dogma.Event) error {
					return IgnoreMessage
				},
			),
			expectFail,
			expectReport(
				`✗ record an event that matches the predicate near expectation.messagematch.event_test.go:279`,
				``,
				`  | EXPLANATION`,
				`  |     none of the engaged handlers recorded a matching event`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • verify the logic within the predicate function, it ignored 1 event`,
				`  |     • enable integration handlers using the EnableHandlerType() option`,
				`  |     • verify the logic within the '<aggregate>' aggregate message handler`,
			),
			nil,
		},
		{
			"no matching event recorded and error messages were repeated",
			func(*testing.T, *Test) Action {
				return ExecuteCommand(&CommandThatRecordsEvent{
					Content: "<multiple>",
				})
			},
			ToRecordEventMatching(
				func(m dogma.Event) error {
					return errors.New("<error>")
				},
			),
			expectFail,
			expectReport(
				`✗ record an event that matches the predicate near expectation.messagematch.event_test.go:305`,
				``,
				`  | EXPLANATION`,
				`  |     none of the engaged handlers recorded a matching event`,
				`  | `,
				`  | FAILED MATCHES`,
				`  |     • *stubs.EventStub[TypeE]: <error> (repeated 2 times)`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • verify the logic within the predicate function`,
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
				ToRecordEventMatching(
					func(m dogma.Event) error {
						return nil
					},
				),
				ToRecordEventMatching(
					func(m dogma.Event) error {
						return errors.New("<error>")
					},
				),
			),
			expectFail,
			expectReport(
				`✗ none of (1 of the expectations passed unexpectedly)`,
				`    ✓ record an event that matches the predicate near expectation.messagematch.event_test.go:334`,
				`    ✗ record an event that matches the predicate near expectation.messagematch.event_test.go:338`,
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

	t.Run("fails the test if the message type is unrecognized", func(t *testing.T) {
		mt := &testingmock.T{FailSilently: true}
		tc := Begin(mt, app)
		tc.Expect(
			noop,
			ToRecordEventMatching(
				func(*EventStub[TypeU]) error {
					return nil
				},
			),
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
			ToRecordEventMatching(
				func(*EventThatExecutesCommand) error {
					return nil
				},
			),
		)

		if !mt.Failed() {
			t.Fatal("expected test to fail")
		}
		if !containsString(mt.Logs, "no handlers record events of type *stubs.EventStub[TypeC], it is only ever consumed") {
			t.Fatalf("expected log message not found, got: %v", mt.Logs)
		}
	})

	t.Run("panics if the function is nil", func(t *testing.T) {
		xtesting.ExpectPanic(
			t,
			"ToRecordEventMatching(<nil>): function must not be nil",
			func() {
				var fn func(dogma.Event) error
				ToRecordEventMatching(fn)
			},
		)
	})
}
