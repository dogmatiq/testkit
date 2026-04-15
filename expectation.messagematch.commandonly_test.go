package testkit_test

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/enginekit/enginetest/stubs"
	. "github.com/dogmatiq/testkit"
	"github.com/dogmatiq/testkit/internal/testingmock"
	"github.com/dogmatiq/testkit/internal/x/xtesting"
)

func TestToOnlyExecuteCommandsMatching(t *testing.T) {
	type (
		EventThatExecutesCommands = EventStub[TypeC]

		CommandThatIsExecuted      = CommandStub[TypeC]
		CommandThatIsNeverExecuted = CommandStub[TypeX]
		CommandThatIsOnlyConsumed  = CommandStub[TypeO]
	)

	app := &ApplicationStub{
		ConfigureFunc: func(c dogma.ApplicationConfigurer) {
			c.Identity("<app>", "386480e5-4b83-4d3b-9b87-51e6d56e41e7")

			c.Routes(
				dogma.ViaProcess(&ProcessMessageHandlerStub{
					ConfigureFunc: func(c dogma.ProcessConfigurer) {
						c.Identity("<process>", "39869c73-5ff0-4ae6-8317-eb494c87167b")
						c.Routes(
							dogma.HandlesEvent[*EventThatExecutesCommands](),
							dogma.ExecutesCommand[*CommandThatIsExecuted](),
							dogma.ExecutesCommand[*CommandThatIsNeverExecuted](),
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
						s.ExecuteCommand(CommandC1)
						s.ExecuteCommand(CommandC2)
						s.ExecuteCommand(CommandC3)
						return nil
					},
				}),

				dogma.ViaIntegration(&IntegrationMessageHandlerStub{
					ConfigureFunc: func(c dogma.IntegrationConfigurer) {
						c.Identity("<integration>", "20bf2831-1887-4148-9539-eb7c294e80b6")
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
			"all executed commands match",
			func(*testing.T, *Test) Action {
				return RecordEvent(&EventThatExecutesCommands{})
			},
			ToOnlyExecuteCommandsMatching(
				func(m dogma.Command) error {
					return nil
				},
			),
			expectPass,
			expectReport(
				`✓ only execute commands that match the predicate near expectation.messagematch.commandonly_test.go:85`,
			),
			nil,
		},
		{
			"all executed commands match, using predicate with a more specific type",
			func(*testing.T, *Test) Action {
				return RecordEvent(&EventThatExecutesCommands{})
			},
			ToOnlyExecuteCommandsMatching(
				func(*CommandThatIsExecuted) error {
					return nil
				},
			),
			expectPass,
			expectReport(
				`✓ only execute commands that match the predicate near expectation.messagematch.commandonly_test.go:101`,
			),
			nil,
		},
		{
			"no commands executed at all",
			func(*testing.T, *Test) Action {
				return noop
			},
			ToOnlyExecuteCommandsMatching(
				func(m dogma.Command) error {
					panic("unexpected call")
				},
			),
			expectPass,
			expectReport(
				`✓ only execute commands that match the predicate near expectation.messagematch.commandonly_test.go:116`,
			),
			nil,
		},
		{
			"none of the executed commands match",
			func(*testing.T, *Test) Action {
				return RecordEvent(&EventThatExecutesCommands{})
			},
			ToOnlyExecuteCommandsMatching(
				func(m dogma.Command) error {
					return errors.New("<error>")
				},
			),
			expectFail,
			expectReport(
				`✗ only execute commands that match the predicate near expectation.messagematch.commandonly_test.go:132`,
				``,
				`  | EXPLANATION`,
				`  |     none of the 3 relevant commands matched the predicate`,
				`  | `,
				`  | FAILED MATCHES`,
				`  |     • *stubs.CommandStub[TypeC]: <error> (repeated 3 times)`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • verify the logic within the predicate function`,
				`  |     • verify the logic within the '<process>' process message handler`,
			),
			nil,
		},
		{
			"some matching commands executed",
			func(*testing.T, *Test) Action {
				return RecordEvent(&EventThatExecutesCommands{})
			},
			ToOnlyExecuteCommandsMatching(
				func(m dogma.Command) error {
					switch m {
					case CommandC1:
						return errors.New("<error>")
					case CommandC2:
						return IgnoreMessage
					default:
						return nil
					}
				},
			),
			expectFail,
			expectReport(
				`✗ only execute commands that match the predicate near expectation.messagematch.commandonly_test.go:158`,
				``,
				`  | EXPLANATION`,
				`  |     only 1 of 2 relevant commands matched the predicate`,
				`  | `,
				`  | FAILED MATCHES`,
				`  |     • *stubs.CommandStub[TypeC]: <error>`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • verify the logic within the predicate function, it ignored 1 command`,
				`  |     • verify the logic within the '<process>' process message handler`,
			),
			nil,
		},
		{
			"no executed commands match, using predicate with a more specific type",
			func(*testing.T, *Test) Action {
				return RecordEvent(&EventThatExecutesCommands{})
			},
			ToOnlyExecuteCommandsMatching(
				func(*CommandThatIsNeverExecuted) error {
					panic("unexpected call")
				},
			),
			expectFail,
			expectReport(
				`✗ only execute commands that match the predicate near expectation.messagematch.commandonly_test.go:191`,
				``,
				`  | EXPLANATION`,
				`  |     none of the 3 relevant commands matched the predicate`,
				`  | `,
				`  | FAILED MATCHES`,
				`  |     • *stubs.CommandStub[TypeC]: predicate function expected *stubs.CommandStub[TypeX] (repeated 3 times)`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • verify the logic within the predicate function`,
				`  |     • verify the logic within the '<process>' process message handler`,
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
			ToOnlyExecuteCommandsMatching(
				func(*CommandStub[TypeU]) error {
					return nil
				},
			),
		)

		if !mt.Failed() {
			t.Fatal("expected test to fail")
		}
		if !slices.Contains(mt.Logs, "a command of type *stubs.CommandStub[TypeU] can never be executed, the application does not use this message type") {
			t.Fatalf("expected log message not found, got: %v", mt.Logs)
		}
	})

	t.Run("fails the test if the message type is not produced by any handlers", func(t *testing.T) {
		mt := &testingmock.T{FailSilently: true}
		tc := Begin(mt, app)
		tc.Expect(
			noop,
			ToOnlyExecuteCommandsMatching(
				func(*CommandThatIsOnlyConsumed) error {
					return nil
				},
			),
		)

		if !mt.Failed() {
			t.Fatal("expected test to fail")
		}
		if !slices.Contains(mt.Logs, "no handlers execute commands of type *stubs.CommandStub[TypeO], it is only ever consumed") {
			t.Fatalf("expected log message not found, got: %v", mt.Logs)
		}
	})

	t.Run("panics if the function is nil", func(t *testing.T) {
		xtesting.ExpectPanic(
			t,
			"ToOnlyExecuteCommandsMatching(<nil>): function must not be nil",
			func() {
				var fn func(dogma.Command) error
				ToOnlyExecuteCommandsMatching(fn)
			},
		)
	})
}
