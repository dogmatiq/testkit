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

func TestToExecuteCommandType_WhenUsedWithCallAction(t *testing.T) {
	type (
		CommandThatIsIgnored           = CommandStub[TypeX]
		CommandThatIsExecutedByProcess = CommandStub[TypeP]

		EventThatExecutesCommand = EventStub[TypeP]
	)

	app := &ApplicationStub{
		ConfigureFunc: func(c dogma.ApplicationConfigurer) {
			c.Identity("<app>", "f38c3003-bbd0-4b4a-b1f8-6922e9545acd")

			c.Routes(
				dogma.ViaIntegration(&IntegrationMessageHandlerStub{
					ConfigureFunc: func(c dogma.IntegrationConfigurer) {
						c.Identity("<integration>", "efa4e6c1-1131-4ff6-9417-5eda4356c5aa")
						c.Routes(
							dogma.HandlesCommand[*CommandThatIsIgnored](),
						)
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
			"command type executed as expected",
			func(t *testing.T, tc *Test) Action {
				return executeCommandViaExecutor(t, tc, &CommandThatIsIgnored{})
			},
			ToExecuteCommandType[*CommandThatIsIgnored](),
			expectPass,
			expectReport(
				`✓ execute any '*stubs.CommandStub[TypeX]' command`,
			),
			nil,
		},
		{
			"no messages produced at all",
			func(*testing.T, *Test) Action {
				return Call(func() {})
			},
			ToExecuteCommandType[*CommandThatIsIgnored](),
			expectFail,
			expectReport(
				`✗ execute any '*stubs.CommandStub[TypeX]' command`,
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
			"no matching command type executed and all relevant handler types disabled",
			func(t *testing.T, tc *Test) Action {
				return executeCommandViaExecutor(t, tc, &CommandThatIsIgnored{})
			},
			ToExecuteCommandType[*CommandThatIsExecutedByProcess](),
			expectFail,
			expectReport(
				`✗ execute any '*stubs.CommandStub[TypeP]' command`,
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
