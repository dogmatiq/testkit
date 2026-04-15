package testkit_test

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/enginekit/enginetest/stubs"
	. "github.com/dogmatiq/testkit"
	"github.com/dogmatiq/testkit/engine"
	"github.com/dogmatiq/testkit/internal/testingmock"
	"github.com/dogmatiq/testkit/internal/x/xtesting"
)

func TestToExecuteCommandMatching(t *testing.T) {
	type (
		EventThatIsIgnored        = EventStub[TypeX]
		EventThatExecutesCommand  = EventStub[TypeC]
		EventThatSchedulesTimeout = EventStub[TypeT]

		CommandThatIsExecuted      = CommandStub[TypeC]
		CommandThatIsNeverExecuted = CommandStub[TypeX]
		CommandThatIsOnlyConsumed  = CommandStub[TypeO]

		TimeoutThatIsScheduled = TimeoutStub[TypeT]
	)

	app := &ApplicationStub{
		ConfigureFunc: func(c dogma.ApplicationConfigurer) {
			c.Identity("<app>", "95d4b9b2-a0ec-4dfb-aa57-c7e5ef5b1f02")

			c.Routes(
				dogma.ViaProcess(&ProcessMessageHandlerStub{
					ConfigureFunc: func(c dogma.ProcessConfigurer) {
						c.Identity("<process>", "8b4c4701-be92-4b28-83b6-0d69b97fb451")
						c.Routes(
							dogma.HandlesEvent[*EventThatIsIgnored](),

							dogma.HandlesEvent[*EventThatExecutesCommand](),
							dogma.ExecutesCommand[*CommandThatIsExecuted](),
							dogma.ExecutesCommand[*CommandThatIsNeverExecuted](),

							dogma.HandlesEvent[*EventThatSchedulesTimeout](),
							dogma.SchedulesTimeout[*TimeoutThatIsScheduled](),
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
						switch m := m.(type) {
						case *EventThatExecutesCommand:
							s.ExecuteCommand(
								&CommandThatIsExecuted{
									Content: m.Content,
								},
							)

							if m.Content == "<multiple>" {
								s.ExecuteCommand(
									&CommandThatIsExecuted{
										Content: m.Content,
									},
								)
							}

						case *EventThatSchedulesTimeout:
							s.ScheduleTimeout(
								&TimeoutThatIsScheduled{
									Content: m.Content,
								},
								time.Now().Add(1*time.Hour),
							)
						}

						return nil
					},
				}),

				dogma.ViaIntegration(&IntegrationMessageHandlerStub{
					ConfigureFunc: func(c dogma.IntegrationConfigurer) {
						c.Identity("<integration>", "7cf5a7fe-9f69-46be-8c59-cc12c4825aaf")
						c.Routes(
							dogma.HandlesCommand[*CommandThatIsOnlyConsumed](),
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
			"matching command executed as expected",
			func(*testing.T, *Test) Action {
				return RecordEvent(&EventThatExecutesCommand{})
			},
			ToExecuteCommandMatching(
				func(m dogma.Command) error {
					if m, ok := m.(*CommandThatIsExecuted); ok && *m == (CommandThatIsExecuted{}) {
						return nil
					}

					return errors.New("<error>")
				},
			),
			expectPass,
			expectReport(
				`✓ execute a command that matches the predicate near expectation.messagematch.command_test.go:117`,
			),
			nil,
		},
		{
			"matching command executed as expected, using predicate with a more specific type",
			func(*testing.T, *Test) Action {
				return RecordEvent(&EventThatExecutesCommand{})
			},
			ToExecuteCommandMatching(
				func(m *CommandThatIsExecuted) error {
					if *m == (CommandThatIsExecuted{}) {
						return nil
					}

					return errors.New("<error>")
				},
			),
			expectPass,
			expectReport(
				`✓ execute a command that matches the predicate near expectation.messagematch.command_test.go:137`,
			),
			nil,
		},
		{
			"no matching command executed",
			func(*testing.T, *Test) Action {
				return RecordEvent(&EventThatExecutesCommand{})
			},
			ToExecuteCommandMatching(
				func(m dogma.Command) error {
					return errors.New("<error>")
				},
			),
			expectFail,
			expectReport(
				`✗ execute a command that matches the predicate near expectation.messagematch.command_test.go:157`,
				``,
				`  | EXPLANATION`,
				`  |     none of the engaged handlers executed a matching command`,
				`  | `,
				`  | FAILED MATCHES`,
				`  |     • *stubs.CommandStub[TypeC]: <error>`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • verify the logic within the predicate function`,
				`  |     • verify the logic within the '<process>' process message handler`,
			),
			nil,
		},
		{
			"no matching command executed, using predicate with a more specific type",
			func(*testing.T, *Test) Action {
				return RecordEvent(&EventThatExecutesCommand{})
			},
			ToExecuteCommandMatching(
				func(m *CommandThatIsNeverExecuted) error {
					panic("unexpected call")
				},
			),
			expectFail,
			expectReport(
				`✗ execute a command that matches the predicate near expectation.messagematch.command_test.go:183`,
				``,
				`  | EXPLANATION`,
				`  |     none of the engaged handlers executed a matching command`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • verify the logic within the predicate function, it ignored 1 command`,
				`  |     • verify the logic within the '<process>' process message handler`,
			),
			nil,
		},
		{
			"no messages produced at all",
			func(*testing.T, *Test) Action {
				return RecordEvent(&EventThatIsIgnored{})
			},
			ToExecuteCommandMatching(
				func(m dogma.Command) error {
					panic("unexpected call")
				},
			),
			expectFail,
			expectReport(
				`✗ execute a command that matches the predicate near expectation.messagematch.command_test.go:206`,
				``,
				`  | EXPLANATION`,
				`  |     no messages were produced at all`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • verify the logic within the '<process>' process message handler`,
			),
			nil,
		},
		{
			"no commands produced at all",
			func(*testing.T, *Test) Action {
				return RecordEvent(&EventThatSchedulesTimeout{})
			},
			ToExecuteCommandMatching(
				func(m dogma.Command) error {
					panic("unexpected call")
				},
			),
			expectFail,
			expectReport(
				`✗ execute a command that matches the predicate near expectation.messagematch.command_test.go:228`,
				``,
				`  | EXPLANATION`,
				`  |     no commands were executed at all`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • verify the logic within the '<process>' process message handler`,
			),
			nil,
		},
		{
			"no matching command executed and all relevant handler types disabled",
			func(*testing.T, *Test) Action {
				return RecordEvent(&EventThatExecutesCommand{})
			},
			ToExecuteCommandMatching(
				func(m dogma.Command) error {
					return errors.New("<error>")
				},
			),
			expectFail,
			expectReport(
				`✗ execute a command that matches the predicate near expectation.messagematch.command_test.go:250`,
				``,
				`  | EXPLANATION`,
				`  |     no relevant handler types were enabled`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • enable process handlers using the EnableHandlerType() option`,
			),
			[]TestOption{
				WithUnsafeOperationOptions(
					engine.EnableProcesses(false),
				),
			},
		},
		{
			"no matching command executed and commands were ignored",
			func(*testing.T, *Test) Action {
				return RecordEvent(&EventThatExecutesCommand{})
			},
			ToExecuteCommandMatching(
				func(m dogma.Command) error {
					return IgnoreMessage
				},
			),
			expectFail,
			expectReport(
				`✗ execute a command that matches the predicate near expectation.messagematch.command_test.go:276`,
				``,
				`  | EXPLANATION`,
				`  |     none of the engaged handlers executed a matching command`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • verify the logic within the predicate function, it ignored 1 command`,
				`  |     • verify the logic within the '<process>' process message handler`,
			),
			nil,
		},
		{
			"no matching command executed and error messages were repeated",
			func(*testing.T, *Test) Action {
				return RecordEvent(&EventThatExecutesCommand{
					Content: "<multiple>",
				})
			},
			ToExecuteCommandMatching(
				func(m dogma.Command) error {
					return errors.New("<error>")
				},
			),
			expectFail,
			expectReport(
				`✗ execute a command that matches the predicate near expectation.messagematch.command_test.go:301`,
				``,
				`  | EXPLANATION`,
				`  |     none of the engaged handlers executed a matching command`,
				`  | `,
				`  | FAILED MATCHES`,
				`  |     • *stubs.CommandStub[TypeC]: <error> (repeated 2 times)`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • verify the logic within the predicate function`,
				`  |     • verify the logic within the '<process>' process message handler`,
			),
			nil,
		},
		{
			"does not include an explanation when negated and a sibling expectation passes",
			func(*testing.T, *Test) Action {
				return RecordEvent(&EventThatExecutesCommand{})
			},
			NoneOf(
				ToExecuteCommandMatching(
					func(m dogma.Command) error {
						return nil
					},
				),
				ToExecuteCommandMatching(
					func(m dogma.Command) error {
						return errors.New("<error>")
					},
				),
			),
			expectFail,
			expectReport(
				`✗ none of (1 of the expectations passed unexpectedly)`,
				`    ✓ execute a command that matches the predicate near expectation.messagematch.command_test.go:329`,
				`    ✗ execute a command that matches the predicate near expectation.messagematch.command_test.go:333`,
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

	t.Run("fails the test if the message type is unrecognized", func(t *testing.T) {
		mt := &testingmock.T{FailSilently: true}
		tc := Begin(mt, app)
		tc.Expect(
			noop,
			ToExecuteCommandMatching(
				func(*CommandStub[TypeU]) error {
					return nil
				},
			),
		)

		if !mt.Failed() {
			t.Fatal("expected test to fail")
		}
		if !slices.Contains(mt.Logs, "  |     a command of type *stubs.CommandStub[TypeU] can never be executed, the application does not use this message type") {
			t.Fatalf("expected log message not found, got: %v", mt.Logs)
		}
	})

	t.Run("fails the test if the message type is not produced by any handlers", func(t *testing.T) {
		mt := &testingmock.T{FailSilently: true}
		tc := Begin(mt, app)
		tc.Expect(
			noop,
			ToExecuteCommandMatching(
				func(*CommandThatIsOnlyConsumed) error {
					return nil
				},
			),
		)

		if !mt.Failed() {
			t.Fatal("expected test to fail")
		}
		if !slices.Contains(mt.Logs, "  |     no handlers execute commands of type *stubs.CommandStub[TypeO], it is only ever consumed") {
			t.Fatalf("expected log message not found, got: %v", mt.Logs)
		}
	})

	t.Run("panics if the function is nil", func(t *testing.T) {
		xtesting.ExpectPanic(
			t,
			"ToExecuteCommandMatching(<nil>): function must not be nil",
			func() {
				var fn func(dogma.Command) error
				ToExecuteCommandMatching(fn)
			},
		)
	})
}
