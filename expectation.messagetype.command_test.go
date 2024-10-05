package testkit_test

import (
	"context"
	"time"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/enginekit/enginetest/stubs"
	. "github.com/dogmatiq/enginekit/enginetest/stubs"
	. "github.com/dogmatiq/testkit"
	"github.com/dogmatiq/testkit/engine"
	"github.com/dogmatiq/testkit/internal/testingmock"
	g "github.com/onsi/ginkgo/v2"
	gm "github.com/onsi/gomega"
)

var _ = g.Describe("func ToExecuteCommandType()", func() {
	var (
		testingT *testingmock.T
		app      dogma.Application
	)

	type (
		EventThatIsIgnored        = EventStub[TypeX]
		EventThatExecutesCommand  = EventStub[TypeC]
		EventThatSchedulesTimeout = EventStub[TypeT]

		CommandThatIsExecuted      = CommandStub[TypeC]
		CommandThatIsNeverExecuted = CommandStub[TypeX]
		CommandThatIsOnlyConsumed  = CommandStub[TypeO]

		TimeoutThatIsScheduled = TimeoutStub[TypeT]
	)

	g.BeforeEach(func() {
		testingT = &testingmock.T{
			FailSilently: true,
		}

		app = &ApplicationStub{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "936ab3fa-f379-42e7-9100-a4d28accc932")

				// Register a process that will execute the commands about which
				// we will make assertions using ToExecuteCommand().
				c.RegisterProcess(&ProcessMessageHandlerStub{
					ConfigureFunc: func(c dogma.ProcessConfigurer) {
						c.Identity("<process>", "72df8a82-b6ab-4fed-bfdc-1fedf6636041")
						c.Routes(
							dogma.HandlesEvent[EventThatIsIgnored](),

							dogma.HandlesEvent[EventThatExecutesCommand](),
							dogma.ExecutesCommand[CommandThatIsExecuted](),
							dogma.ExecutesCommand[CommandThatIsNeverExecuted](),

							dogma.HandlesEvent[EventThatSchedulesTimeout](),
							dogma.SchedulesTimeout[TimeoutThatIsScheduled](),
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
								CommandThatIsExecuted{
									Content: m.Content,
								},
							)
						case EventThatSchedulesTimeout:
							s.ScheduleTimeout(
								TimeoutThatIsScheduled{
									Content: m.Content,
								},
								time.Now().Add(1*time.Hour),
							)
						}

						return nil
					},
				})

				// Register an integration so that we can test what happens when
				// we expect execution of a command that is never executed by
				// any handler (only consumed).
				c.RegisterIntegration(&IntegrationMessageHandlerStub{
					ConfigureFunc: func(c dogma.IntegrationConfigurer) {
						c.Identity("<integration>", "bc84090e-270c-4ca9-bb4e-4b152031d996")
						c.Routes(
							dogma.HandlesCommand[CommandThatIsOnlyConsumed](),
						)
					},
				})
			},
		}
	})

	g.DescribeTable(
		"expectation behavior",
		func(
			a Action,
			e Expectation,
			ok bool,
			rm reportMatcher,
			options ...TestOption,
		) {
			test := Begin(testingT, app, options...)
			test.Expect(a, e)
			rm(testingT)
			gm.Expect(testingT.Failed()).To(gm.Equal(!ok))
		},
		g.Entry(
			"command type executed as expected",
			RecordEvent(EventThatExecutesCommand{}),
			ToExecuteCommandType[CommandThatIsExecuted](),
			expectPass,
			expectReport(
				`✓ execute any 'stubs.CommandStub[TypeC]' command`,
			),
		),
		g.Entry(
			"no matching command types executed",
			RecordEvent(EventThatExecutesCommand{}),
			ToExecuteCommandType[CommandThatIsNeverExecuted](),
			expectFail,
			expectReport(
				`✗ execute any 'stubs.CommandStub[TypeX]' command`,
				``,
				`  | EXPLANATION`,
				`  |     none of the engaged handlers executed a matching command`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • verify the logic within the '<process>' process message handler`,
			),
		),
		g.Entry(
			"no messages produced at all",
			RecordEvent(EventThatIsIgnored{}),
			ToExecuteCommandType[CommandThatIsExecuted](),
			expectFail,
			expectReport(
				`✗ execute any 'stubs.CommandStub[TypeC]' command`,
				``,
				`  | EXPLANATION`,
				`  |     no messages were produced at all`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • verify the logic within the '<process>' process message handler`,
			),
		),
		g.Entry(
			"no commands produced at all",
			RecordEvent(EventThatSchedulesTimeout{}),
			ToExecuteCommandType[CommandThatIsExecuted](),
			expectFail,
			expectReport(
				`✗ execute any 'stubs.CommandStub[TypeC]' command`,
				``,
				`  | EXPLANATION`,
				`  |     no commands were executed at all`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • verify the logic within the '<process>' process message handler`,
			),
		),
		g.Entry(
			"no matching command type executed and all relevant handler types disabled",
			RecordEvent(EventThatExecutesCommand{}),
			ToExecuteCommandType[CommandThatIsExecuted](),
			expectFail,
			expectReport(
				`✗ execute any 'stubs.CommandStub[TypeC]' command`,
				``,
				`  | EXPLANATION`,
				`  |     no relevant handler types were enabled`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • enable process handlers using the EnableHandlerType() option`,
			),
			WithUnsafeOperationOptions(
				engine.EnableProcesses(false),
			),
		),
		g.Entry(
			"does not include an explanation when negated and a sibling expectation passes",
			RecordEvent(EventThatExecutesCommand{}),
			NoneOf(
				ToExecuteCommandType[CommandThatIsExecuted](),
				ToExecuteCommandType[CommandThatIsNeverExecuted](),
			),
			expectFail,
			expectReport(
				`✗ none of (1 of the expectations passed unexpectedly)`,
				`    ✓ execute any 'stubs.CommandStub[TypeC]' command`,
				`    ✗ execute any 'stubs.CommandStub[TypeX]' command`,
			),
		),
	)

	g.It("fails the test if the message type is unrecognized", func() {
		test := Begin(testingT, app)
		test.Expect(
			noop,
			ToExecuteCommandType[stubs.CommandStub[TypeU]](),
		)

		gm.Expect(testingT.Failed()).To(gm.BeTrue())
		gm.Expect(testingT.Logs).To(gm.ContainElement(
			"a command of type stubs.CommandStub[TypeU] can never be executed, the application does not use this message type",
		))
	})

	g.It("fails the test if the message type is not produced by any handlers", func() {
		test := Begin(testingT, app)
		test.Expect(
			noop,
			ToExecuteCommandType[CommandThatIsOnlyConsumed](),
		)

		gm.Expect(testingT.Failed()).To(gm.BeTrue())
		gm.Expect(testingT.Logs).To(gm.ContainElement(
			"no handlers execute commands of type stubs.CommandStub[TypeO], it is only ever consumed",
		))
	})
})
