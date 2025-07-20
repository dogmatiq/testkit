package testkit_test

import (
	"context"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/enginekit/enginetest/stubs"
	. "github.com/dogmatiq/testkit"
	"github.com/dogmatiq/testkit/engine"
	"github.com/dogmatiq/testkit/internal/testingmock"
	g "github.com/onsi/ginkgo/v2"
	gm "github.com/onsi/gomega"
)

var _ = g.Describe("func ToExecuteCommandType() (when used with the Call() action)", func() {
	var (
		testingT *testingmock.T
		app      dogma.Application
		test     *Test
	)

	type (
		CommandThatIsIgnored           = CommandStub[TypeX]
		CommandThatIsExecutedByProcess = CommandStub[TypeP]

		EventThatExecutesCommand = EventStub[TypeP]
	)

	g.BeforeEach(func() {
		testingT = &testingmock.T{
			FailSilently: true,
		}

		app = &ApplicationStub{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "f38c3003-bbd0-4b4a-b1f8-6922e9545acd")

				c.Routes(
					dogma.ViaIntegration(&IntegrationMessageHandlerStub{
						ConfigureFunc: func(c dogma.IntegrationConfigurer) {
							c.Identity("<integration>", "efa4e6c1-1131-4ff6-9417-5eda4356c5aa")
							c.Routes(
								dogma.HandlesCommand[CommandThatIsIgnored](),
							)
						},
					}),

					dogma.ViaProcess(&ProcessMessageHandlerStub{
						ConfigureFunc: func(c dogma.ProcessConfigurer) {
							c.Identity("<process>", "8b4c4701-be92-4b28-83b6-0d69b97fb451")
							c.Routes(
								dogma.HandlesEvent[EventThatExecutesCommand](),
								dogma.ExecutesCommand[CommandThatIsExecutedByProcess](),
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
							case EventThatExecutesCommand:
								s.ExecuteCommand(
									CommandThatIsExecutedByProcess{
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
	})

	executeCommandViaExecutor := func(m dogma.Command) Action {
		return Call(func() {
			err := test.CommandExecutor().ExecuteCommand(context.Background(), m)
			gm.Expect(err).ShouldNot(gm.HaveOccurred())
		})
	}

	g.DescribeTable(
		"expectation behavior",
		func(
			a Action,
			e Expectation,
			ok bool,
			rm reportMatcher,
			options ...TestOption,
		) {
			test = Begin(testingT, app, options...)
			test.Expect(a, e)
			rm(testingT)
			gm.Expect(testingT.Failed()).To(gm.Equal(!ok))
		},
		g.Entry(
			"command type executed as expected",
			executeCommandViaExecutor(CommandThatIsIgnored{}),
			ToExecuteCommandType[CommandThatIsIgnored](),
			expectPass,
			expectReport(
				`✓ execute any 'stubs.CommandStub[TypeX]' command`,
			),
		),
		g.Entry(
			"no messages produced at all",
			Call(func() {}),
			ToExecuteCommandType[CommandThatIsIgnored](),
			expectFail,
			expectReport(
				`✗ execute any 'stubs.CommandStub[TypeX]' command`,
				``,
				`  | EXPLANATION`,
				`  |     no messages were produced at all`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • verify the logic within the code that uses the dogma.CommandExecutor`,
			),
		),
		g.Entry(
			"no matching command type executed and all relevant handler types disabled",
			executeCommandViaExecutor(CommandThatIsIgnored{}),
			ToExecuteCommandType[CommandThatIsExecutedByProcess](),
			expectFail,
			expectReport(
				`✗ execute any 'stubs.CommandStub[TypeP]' command`,
				``,
				`  | EXPLANATION`,
				`  |     nothing executed a matching command`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • enable process handlers using the EnableHandlerType() option`,
				`  |     • verify the logic within the code that uses the dogma.CommandExecutor`,
			),
			WithUnsafeOperationOptions(
				engine.EnableProcesses(false),
			),
		),
	)
})
