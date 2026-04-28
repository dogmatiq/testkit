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

func TestToExecuteCommand_WhenUsedWithCallAction(t *testing.T) {
	type (
		CommandThatIsIgnored           = CommandStub[TypeX]
		CommandThatRecordsEvent        = CommandStub[TypeE]
		CommandThatIsExecutedByProcess = CommandStub[TypeP]

		EventThatExecutesCommand = EventStub[TypeP]
	)

	app := &ApplicationStub{
		ConfigureFunc: func(c dogma.ApplicationConfigurer) {
			c.Identity("<app>", "556be8fb-fa7b-4240-882e-86e735b7705d")

			c.Routes(
				dogma.ViaIntegration(&IntegrationMessageHandlerStub{
					ConfigureFunc: func(c dogma.IntegrationConfigurer) {
						c.Identity("<integration>", "b4f6e091-6171-4c61-bf3b-9952aea28547")
						c.Routes(
							dogma.HandlesCommand[*CommandThatIsIgnored](),
							dogma.HandlesCommand[*CommandThatRecordsEvent](),
						)
					},
					HandleCommandFunc: func(
						_ context.Context,
						s dogma.IntegrationCommandScope,
						m dogma.Command,
					) error {
						switch m.(type) {
						case *CommandThatRecordsEvent:
							s.RecordEvent(&EventThatExecutesCommand{})
						}
						return nil
					},
				}),

				dogma.ViaProcess(&ProcessMessageHandlerStub[*ProcessRootStub]{
					ConfigureFunc: func(c dogma.ProcessConfigurer) {
						c.Identity("<process>", "8b4c4701-be92-4b28-83b6-0d69b97fb451")
						c.Routes(
							dogma.HandlesEvent[*EventThatExecutesCommand](),
							dogma.ExecutesCommand[*CommandThatIsExecutedByProcess](),
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
						_ *ProcessRootStub,
						s dogma.ProcessEventScope[*ProcessRootStub],
						m dogma.Event,
					) error {
						switch m := m.(type) {
						case *EventThatExecutesCommand:
							s.ExecuteCommand(
								&CommandThatIsExecutedByProcess{
									Content: m.Content,
								},
							)
						}

						return nil
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
			"command executed as expected",
			func(t *testing.T, tc *Test) Action {
				return executeCommandViaExecutor(t, tc, &CommandThatIsIgnored{})
			},
			ToExecuteCommand(&CommandThatIsIgnored{}),
			expectPass,
			expectReport(
				`✓ execute a specific '*stubs.CommandStub[TypeX]' command`,
			),
			nil,
		},
		{
			"no messages produced at all",
			func(*testing.T, *Test) Action {
				return Call(func() {})
			},
			ToExecuteCommand(&CommandThatIsIgnored{}),
			expectFail,
			expectReport(
				`✗ execute a specific '*stubs.CommandStub[TypeX]' command`,
				``,
				`  | EXPLANATION`,
				`  |     no messages were produced at all`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • verify the logic within the code that uses the dogma.CommandExecutor`,
			),
			nil,
		},
		{
			"no matching command executed and all relevant handler types disabled",
			func(t *testing.T, tc *Test) Action {
				return executeCommandViaExecutor(t, tc, &CommandThatRecordsEvent{})
			},
			ToExecuteCommand(&CommandThatIsExecutedByProcess{}),
			expectFail,
			expectReport(
				`✗ execute a specific '*stubs.CommandStub[TypeP]' command`,
				``,
				`  | EXPLANATION`,
				`  |     nothing executed a matching command`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • enable process handlers using the EnableHandlerType() option`,
				`  |     • verify the logic within the code that uses the dogma.CommandExecutor`,
			),
			[]TestOption{
				WithUnsafeOperationOptions(
					engine.EnableProcesses(false),
				),
			},
		},
		{
			"similar command executed with a different value",
			func(t *testing.T, tc *Test) Action {
				return executeCommandViaExecutor(t, tc, &CommandThatIsIgnored{Content: "<content>"})
			},
			ToExecuteCommand(&CommandThatIsIgnored{Content: "<different>"}),
			expectFail,
			expectReport(
				`✗ execute a specific '*stubs.CommandStub[TypeX]' command`,
				``,
				`  | EXPLANATION`,
				`  |     a similar command was executed via a dogma.CommandExecutor`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • check the content of the message`,
				`  | `,
				`  | MESSAGE DIFF`,
				`  |     *stubs.CommandStub[github.com/dogmatiq/enginekit/enginetest/stubs.TypeX]{`,
				`  |         Content:         "<[-differ-]{+cont+}ent>"`,
				`  |         ValidationError: ""`,
				`  |     }`,
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
				t.Fatalf("testingT.Failed() = %v, want %v", mt.Failed(), !c.Passes)
			}
		})
	}
}
