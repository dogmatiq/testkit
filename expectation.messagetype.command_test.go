package testkit_test

import (
	"context"
	"testing"
	"time"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/enginekit/enginetest/stubs"
	. "github.com/dogmatiq/enginekit/enginetest/stubs"
	. "github.com/dogmatiq/testkit"
	"github.com/dogmatiq/testkit/engine"
	"github.com/dogmatiq/testkit/internal/testingmock"
)

func TestToExecuteCommandType(t *testing.T) {
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
			c.Identity("<app>", "936ab3fa-f379-42e7-9100-a4d28accc932")

			c.Routes(
				dogma.ViaProcess(&ProcessMessageHandlerStub{
					ConfigureFunc: func(c dogma.ProcessConfigurer) {
						c.Identity("<process>", "72df8a82-b6ab-4fed-bfdc-1fedf6636041")
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
						c.Identity("<integration>", "bc84090e-270c-4ca9-bb4e-4b152031d996")
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
			"command type executed as expected",
			func(*testing.T, *Test) Action {
				return RecordEvent(&EventThatExecutesCommand{})
			},
			ToExecuteCommandType[*CommandThatIsExecuted](),
			expectPass,
			expectReport(
				`✓ execute any '*stubs.CommandStub[TypeC]' command`,
			),
			nil,
		},
		{
			"no matching command types executed",
			func(*testing.T, *Test) Action {
				return RecordEvent(&EventThatExecutesCommand{})
			},
			ToExecuteCommandType[*CommandThatIsNeverExecuted](),
			expectFail,
			expectReport(
				`✗ execute any '*stubs.CommandStub[TypeX]' command`,
				``,
				`  | EXPLANATION`,
				`  |     none of the engaged handlers executed a matching command`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • verify the logic within the '<process>' process message handler`,
			),
			nil,
		},
		{
			"no messages produced at all",
			func(*testing.T, *Test) Action {
				return RecordEvent(&EventThatIsIgnored{})
			},
			ToExecuteCommandType[*CommandThatIsExecuted](),
			expectFail,
			expectReport(
				`✗ execute any '*stubs.CommandStub[TypeC]' command`,
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
			ToExecuteCommandType[*CommandThatIsExecuted](),
			expectFail,
			expectReport(
				`✗ execute any '*stubs.CommandStub[TypeC]' command`,
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
			"no matching command type executed and all relevant handler types disabled",
			func(*testing.T, *Test) Action {
				return RecordEvent(&EventThatExecutesCommand{})
			},
			ToExecuteCommandType[*CommandThatIsExecuted](),
			expectFail,
			expectReport(
				`✗ execute any '*stubs.CommandStub[TypeC]' command`,
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
			"does not include an explanation when negated and a sibling expectation passes",
			func(*testing.T, *Test) Action {
				return RecordEvent(&EventThatExecutesCommand{})
			},
			NoneOf(
				ToExecuteCommandType[*CommandThatIsExecuted](),
				ToExecuteCommandType[*CommandThatIsNeverExecuted](),
			),
			expectFail,
			expectReport(
				`✗ none of (1 of the expectations passed unexpectedly)`,
				`    ✓ execute any '*stubs.CommandStub[TypeC]' command`,
				`    ✗ execute any '*stubs.CommandStub[TypeX]' command`,
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
			ToExecuteCommandType[*stubs.CommandStub[TypeU]](),
		)

		if !mt.Failed() {
			t.Fatal("expected test to fail")
		}
		if !containsString(mt.Logs, "a command of type *stubs.CommandStub[TypeU] can never be executed, the application does not use this message type") {
			t.Fatalf("expected log message not found, got: %v", mt.Logs)
		}
	})

	t.Run("fails the test if the message type is not produced by any handlers", func(t *testing.T) {
		mt := &testingmock.T{FailSilently: true}
		tc := Begin(mt, app)
		tc.Expect(
			noop,
			ToExecuteCommandType[*CommandThatIsOnlyConsumed](),
		)

		if !mt.Failed() {
			t.Fatal("expected test to fail")
		}
		if !containsString(mt.Logs, "no handlers execute commands of type *stubs.CommandStub[TypeO], it is only ever consumed") {
			t.Fatalf("expected log message not found, got: %v", mt.Logs)
		}
	})
}
