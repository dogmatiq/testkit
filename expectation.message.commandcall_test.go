package testkit_test

import (
	"context"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/enginekit/enginetest/stubs"
	. "github.com/dogmatiq/testkit"
	"github.com/dogmatiq/testkit/engine"
	"github.com/dogmatiq/testkit/internal/testingmock"
	g "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = g.Describe("func ToExecuteCommand() (when used with the Call() action)", func() {
	var (
		testingT *testingmock.T
		app      dogma.Application
		test     *Test
	)

	type (
		CommandThatIsIgnored           = CommandStub[TypeX]
		CommandThatRecordsEvent        = CommandStub[dogma.Event]
		CommandThatIsExecutedByProcess = CommandStub[TypeP]

		EventThatExecutesCommand = EventStub[TypeP]
	)

	g.BeforeEach(func() {
		testingT = &testingmock.T{
			FailSilently: true,
		}

		app = &ApplicationStub{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "556be8fb-fa7b-4240-882e-86e735b7705d")

				c.RegisterIntegration(&IntegrationMessageHandlerStub{
					ConfigureFunc: func(c dogma.IntegrationConfigurer) {
						c.Identity("<integration>", "b4f6e091-6171-4c61-bf3b-9952aea28547")
						c.Routes(
							dogma.HandlesCommand[CommandThatIsIgnored](),
							dogma.HandlesCommand[*CommandThatIsIgnored](), // pointer, used to test type similarity
							dogma.HandlesCommand[CommandThatRecordsEvent](),
						)
					},
					HandleCommandFunc: func(
						_ context.Context,
						s dogma.IntegrationCommandScope,
						m dogma.Command,
					) error {
						switch m := m.(type) {
						case CommandThatRecordsEvent:
							s.RecordEvent(m.Content)
						}
						return nil
					},
				})

				c.RegisterProcess(&ProcessMessageHandlerStub{
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
				})
			},
		}
	})

	executeCommandViaExecutor := func(m dogma.Command) Action {
		return Call(func() {
			err := test.CommandExecutor().ExecuteCommand(context.Background(), m)
			Expect(err).ShouldNot(HaveOccurred())
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
			Expect(testingT.Failed()).To(Equal(!ok))
		},
		g.Entry(
			"command executed as expected",
			executeCommandViaExecutor(CommandThatIsIgnored{}),
			ToExecuteCommand(CommandThatIsIgnored{}),
			expectPass,
			expectReport(
				`✓ execute a specific 'stubs.CommandStub[TypeX]' command`,
			),
		),
		g.Entry(
			"no messages produced at all",
			Call(func() {}),
			ToExecuteCommand(CommandThatIsIgnored{}),
			expectFail,
			expectReport(
				`✗ execute a specific 'stubs.CommandStub[TypeX]' command`,
				``,
				`  | EXPLANATION`,
				`  |     no messages were produced at all`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • verify the logic within the code that uses the dogma.CommandExecutor`,
			),
		),
		g.Entry(
			"no matching command executed and all relevant handler types disabled",
			executeCommandViaExecutor(CommandThatRecordsEvent{
				Content: EventThatExecutesCommand{},
			}),
			ToExecuteCommand(CommandThatIsExecutedByProcess{}),
			expectFail,
			expectReport(
				`✗ execute a specific 'stubs.CommandStub[TypeP]' command`,
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
		g.Entry(
			"similar command executed with a different type",
			executeCommandViaExecutor(CommandThatIsIgnored{}),
			ToExecuteCommand(&CommandThatIsIgnored{}), // note: message type is pointer
			expectFail,
			expectReport(
				`✗ execute a specific '*stubs.CommandStub[TypeX]' command`,
				``,
				`  | EXPLANATION`,
				`  |     a command of a similar type was executed via a dogma.CommandExecutor`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • check the message type, should it be a pointer?`,
				`  | `,
				`  | MESSAGE DIFF`,
				`  |     [-*-]stubs.CommandStub[github.com/dogmatiq/enginekit/enginetest/stubs.TypeX]{<zero>}`,
			),
		),
		g.Entry(
			"similar command executed with a different value",
			executeCommandViaExecutor(CommandThatIsIgnored{Content: "<content>"}),
			ToExecuteCommand(CommandThatIsIgnored{Content: "<different>"}),
			expectFail,
			expectReport(
				`✗ execute a specific 'stubs.CommandStub[TypeX]' command`,
				``,
				`  | EXPLANATION`,
				`  |     a similar command was executed via a dogma.CommandExecutor`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • check the content of the message`,
				`  | `,
				`  | MESSAGE DIFF`,
				`  |     stubs.CommandStub[github.com/dogmatiq/enginekit/enginetest/stubs.TypeX]{`,
				`  |         Content:         "<[-differ-]{+cont+}ent>"`,
				`  |         ValidationError: ""`,
				`  |     }`,
			),
		),
	)
})
