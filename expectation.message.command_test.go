package testkit_test

import (
	"context"
	"time"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/enginekit/enginetest/stubs"
	. "github.com/dogmatiq/testkit"
	"github.com/dogmatiq/testkit/engine"
	"github.com/dogmatiq/testkit/internal/testingmock"
	g "github.com/onsi/ginkgo/v2"
	gm "github.com/onsi/gomega"
)

var _ = g.Describe("func ToExecuteCommand()", func() {
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
				c.Identity("<app>", "ce773269-4ad7-4c7f-a0ff-cda2e5899743")

				c.RegisterProcess(&ProcessMessageHandlerStub{
					ConfigureFunc: func(c dogma.ProcessConfigurer) {
						c.Identity("<process>", "8b4c4701-be92-4b28-83b6-0d69b97fb451")
						c.Routes(
							dogma.HandlesEvent[EventThatIsIgnored](),

							dogma.HandlesEvent[EventThatExecutesCommand](),
							dogma.ExecutesCommand[CommandThatIsExecuted](),
							dogma.ExecutesCommand[*CommandThatIsExecuted](), // pointer, used to test type similarity
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

				c.RegisterIntegration(&IntegrationMessageHandlerStub{
					ConfigureFunc: func(c dogma.IntegrationConfigurer) {
						c.Identity("<integration>", "49fa7c5f-7682-4743-bf8a-ed96dee2d81a")
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
			"command executed as expected",
			RecordEvent(EventThatExecutesCommand{}),
			ToExecuteCommand(CommandThatIsExecuted{}),
			expectPass,
			expectReport(
				`✓ execute a specific 'stubs.CommandStub[TypeC]' command`,
			),
		),
		g.Entry(
			"no matching command executed",
			RecordEvent(EventThatExecutesCommand{}),
			ToExecuteCommand(CommandThatIsNeverExecuted{}),
			expectFail,
			expectReport(
				`✗ execute a specific 'stubs.CommandStub[TypeX]' command`,
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
			ToExecuteCommand(CommandThatIsExecuted{}),
			expectFail,
			expectReport(
				`✗ execute a specific 'stubs.CommandStub[TypeC]' command`,
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
			ToExecuteCommand(CommandThatIsExecuted{}),
			expectFail,
			expectReport(
				`✗ execute a specific 'stubs.CommandStub[TypeC]' command`,
				``,
				`  | EXPLANATION`,
				`  |     no commands were executed at all`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • verify the logic within the '<process>' process message handler`,
			),
		),
		g.Entry(
			"no matching command executed and all relevant handler types disabled",
			RecordEvent(EventThatExecutesCommand{}),
			ToExecuteCommand(CommandThatIsExecuted{}),
			expectFail,
			expectReport(
				`✗ execute a specific 'stubs.CommandStub[TypeC]' command`,
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
			"similar command executed with a different type",
			RecordEvent(EventThatExecutesCommand{}),
			ToExecuteCommand(&CommandThatIsExecuted{}), // note: message type is pointer
			expectFail,
			expectReport(
				`✗ execute a specific '*stubs.CommandStub[TypeC]' command`,
				``,
				`  | EXPLANATION`,
				`  |     a command of a similar type was executed by the '<process>' process message handler`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • check the message type, should it be a pointer?`,
				`  | `,
				`  | MESSAGE DIFF`,
				`  |     [-*-]stubs.CommandStub[github.com/dogmatiq/enginekit/enginetest/stubs.TypeC]{<zero>}`,
			),
		),
		g.Entry(
			"similar command executed with a different value",
			RecordEvent(EventThatExecutesCommand{Content: "<content>"}),
			ToExecuteCommand(CommandThatIsExecuted{Content: "<different>"}),
			expectFail,
			expectReport(
				`✗ execute a specific 'stubs.CommandStub[TypeC]' command`,
				``,
				`  | EXPLANATION`,
				`  |     a similar command was executed by the '<process>' process message handler`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • check the content of the message`,
				`  | `,
				`  | MESSAGE DIFF`,
				`  |     stubs.CommandStub[github.com/dogmatiq/enginekit/enginetest/stubs.TypeC]{`,
				`  |         Content:         "<[-differ-]{+cont+}ent>"`,
				`  |         ValidationError: ""`,
				`  |     }`,
			),
		),
		g.Entry(
			"does not include an explanation when negated and a sibling expectation passes",
			RecordEvent(EventThatExecutesCommand{}),
			NoneOf(
				ToExecuteCommand(CommandThatIsExecuted{}),
				ToExecuteCommand(CommandThatIsNeverExecuted{}),
			),
			expectFail,
			expectReport(
				`✗ none of (1 of the expectations passed unexpectedly)`,
				`    ✓ execute a specific 'stubs.CommandStub[TypeC]' command`,
				`    ✗ execute a specific 'stubs.CommandStub[TypeX]' command`,
			),
		),
	)

	g.It("fails the test if the message type is unrecognized", func() {
		test := Begin(testingT, app)
		test.Expect(
			noop,
			ToExecuteCommand(CommandU1),
		)

		gm.Expect(testingT.Failed()).To(gm.BeTrue())
		gm.Expect(testingT.Logs).To(gm.ContainElement(
			"a command of type stubs.CommandStub[TypeU] can never be executed, the application does not use this message type",
		))
	})

	g.It("fails the test if the message type is not a command", func() {
		test := Begin(testingT, app)
		test.Expect(
			noop,
			ToExecuteCommand(EventThatIsIgnored{}),
		)

		gm.Expect(testingT.Failed()).To(gm.BeTrue())
		gm.Expect(testingT.Logs).To(gm.ContainElement(
			"stubs.EventStub[TypeX] is an event, it can never be executed as a command",
		))
	})

	g.It("fails the test if the message type is not produced by any handlers", func() {
		test := Begin(testingT, app)
		test.Expect(
			noop,
			ToExecuteCommand(CommandThatIsOnlyConsumed{}),
		)

		gm.Expect(testingT.Failed()).To(gm.BeTrue())
		gm.Expect(testingT.Logs).To(gm.ContainElement(
			"no handlers execute commands of type stubs.CommandStub[TypeO], it is only ever consumed",
		))
	})

	g.It("panics if the message is nil", func() {
		gm.Expect(func() {
			ToExecuteCommand(nil)
		}).To(gm.PanicWith("ToExecuteCommand(<nil>): message must not be nil"))
	})

	g.It("panics if the message is invalid", func() {
		gm.Expect(func() {
			ToExecuteCommand(CommandStub[TypeA]{
				ValidationError: "<invalid>",
			})
		}).To(gm.PanicWith("ToExecuteCommand(stubs.CommandStub[TypeA]): <invalid>"))
	})
})
