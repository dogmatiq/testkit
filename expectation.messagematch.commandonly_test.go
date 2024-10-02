package testkit_test

import (
	"context"
	"errors"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/enginekit/enginetest/stubs"
	. "github.com/dogmatiq/testkit"
	"github.com/dogmatiq/testkit/internal/testingmock"
	g "github.com/onsi/ginkgo/v2"
	gm "github.com/onsi/gomega"
)

var _ = g.Describe("func ToOnlyExecuteCommandsMatching()", func() {
	var (
		testingT *testingmock.T
		app      dogma.Application
	)

	type (
		EventThatExecutesCommands = EventStub[TypeC]

		CommandThatIsExecuted      = CommandStub[TypeC]
		CommandThatIsNeverExecuted = CommandStub[TypeX]
		CommandThatIsOnlyConsumed  = CommandStub[TypeO]
	)

	g.BeforeEach(func() {
		testingT = &testingmock.T{
			FailSilently: true,
		}

		app = &ApplicationStub{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "386480e5-4b83-4d3b-9b87-51e6d56e41e7")

				c.RegisterProcess(&ProcessMessageHandlerStub{
					ConfigureFunc: func(c dogma.ProcessConfigurer) {
						c.Identity("<process>", "39869c73-5ff0-4ae6-8317-eb494c87167b")
						c.Routes(
							dogma.HandlesEvent[EventThatExecutesCommands](),
							dogma.ExecutesCommand[CommandThatIsExecuted](),
							dogma.ExecutesCommand[CommandThatIsNeverExecuted](),
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
				})

				c.RegisterIntegration(&IntegrationMessageHandlerStub{
					ConfigureFunc: func(c dogma.IntegrationConfigurer) {
						c.Identity("<integration>", "20bf2831-1887-4148-9539-eb7c294e80b6")
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
			"all executed commands match",
			RecordEvent(EventThatExecutesCommands{}),
			ToOnlyExecuteCommandsMatching(
				func(m dogma.Command) error {
					return nil
				},
			),
			expectPass,
			expectReport(
				`✓ only execute commands that match the predicate near expectation.messagematch.commandonly_test.go:97`,
			),
		),
		g.Entry(
			"all executed commands match, using predicate with a more specific type",
			RecordEvent(EventThatExecutesCommands{}),
			ToOnlyExecuteCommandsMatching(
				func(m CommandThatIsExecuted) error {
					return nil
				},
			),
			expectPass,
			expectReport(
				`✓ only execute commands that match the predicate near expectation.messagematch.commandonly_test.go:110`,
			),
		),
		g.Entry(
			"no commands executed at all",
			noop,
			ToOnlyExecuteCommandsMatching(
				func(m dogma.Command) error {
					panic("unexpected call")
				},
			),
			expectPass,
			expectReport(
				`✓ only execute commands that match the predicate near expectation.messagematch.commandonly_test.go:122`,
			),
		),
		g.Entry(
			"none of the executed commands match",
			RecordEvent(EventThatExecutesCommands{}),
			ToOnlyExecuteCommandsMatching(
				func(m dogma.Command) error {
					return errors.New("<error>")
				},
			),
			expectFail,
			expectReport(
				`✗ only execute commands that match the predicate near expectation.messagematch.commandonly_test.go:135`,
				``,
				`  | EXPLANATION`,
				`  |     none of the 3 relevant commands matched the predicate`,
				`  | `,
				`  | FAILED MATCHES`,
				`  |     • stubs.CommandStub[TypeC]: <error> (repeated 3 times)`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • verify the logic within the predicate function`,
				`  |     • verify the logic within the '<process>' process message handler`,
			),
		),
		g.Entry(
			"some matching commands executed",
			RecordEvent(EventThatExecutesCommands{}),
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
				`  |     • stubs.CommandStub[TypeC]: <error>`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • verify the logic within the predicate function, it ignored 1 command`,
				`  |     • verify the logic within the '<process>' process message handler`,
			),
		),
		g.Entry(
			"no executed commands match, using predicate with a more specific type",
			RecordEvent(EventThatExecutesCommands{}),
			ToOnlyExecuteCommandsMatching(
				func(m CommandThatIsNeverExecuted) error {
					panic("unexpected call")
				},
			),
			expectFail,
			expectReport(
				`✗ only execute commands that match the predicate near expectation.messagematch.commandonly_test.go:188`,
				``,
				`  | EXPLANATION`,
				`  |     none of the 3 relevant commands matched the predicate`,
				`  | `,
				`  | FAILED MATCHES`,
				`  |     • stubs.CommandStub[TypeC]: predicate function expected stubs.CommandStub[TypeX] (repeated 3 times)`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • verify the logic within the predicate function`,
				`  |     • verify the logic within the '<process>' process message handler`,
			),
		),
	)

	g.It("fails the test if the message type is unrecognized", func() {
		test := Begin(testingT, app)
		test.Expect(
			noop,
			ToOnlyExecuteCommandsMatching(
				func(CommandStub[TypeU]) error {
					return nil
				},
			),
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
			ToOnlyExecuteCommandsMatching(
				func(EventThatExecutesCommands) error {
					return nil
				},
			),
		)

		gm.Expect(testingT.Failed()).To(gm.BeTrue())
		gm.Expect(testingT.Logs).To(gm.ContainElement(
			"stubs.EventStub[TypeC] is an event, it can never be executed as a command",
		))
	})

	g.It("fails the test if the message type is not produced by any handlers", func() {
		test := Begin(testingT, app)
		test.Expect(
			noop,
			ToOnlyExecuteCommandsMatching(
				func(CommandThatIsOnlyConsumed) error {
					return nil
				},
			),
		)

		gm.Expect(testingT.Failed()).To(gm.BeTrue())
		gm.Expect(testingT.Logs).To(gm.ContainElement(
			"no handlers execute commands of type stubs.CommandStub[TypeO], it is only ever consumed",
		))
	})

	g.It("panics if the function is nil", func() {
		gm.Expect(func() {
			var fn func(dogma.Command) error
			ToOnlyExecuteCommandsMatching(fn)
		}).To(gm.PanicWith("ToOnlyExecuteCommandsMatching(<nil>): function must not be nil"))
	})
})
