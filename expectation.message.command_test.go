package testkit_test

import (
	"context"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogma/fixtures"
	. "github.com/dogmatiq/testkit"
	"github.com/dogmatiq/testkit/engine"
	"github.com/dogmatiq/testkit/internal/testingmock"
	g "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = g.Describe("func ToExecuteCommand()", func() {
	var (
		testingT *testingmock.T
		app      dogma.Application
	)

	g.BeforeEach(func() {
		testingT = &testingmock.T{
			FailSilently: true,
		}

		app = &Application{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "ce773269-4ad7-4c7f-a0ff-cda2e5899743")

				c.RegisterAggregate(&AggregateMessageHandler{
					ConfigureFunc: func(c dogma.AggregateConfigurer) {
						c.Identity("<aggregate>", "49fa7c5f-7682-4743-bf8a-ed96dee2d81a")
						c.Routes(
							dogma.HandlesCommand[MessageR](),
							dogma.RecordsEvent[MessageN](),
						)
					},
					RouteCommandToInstanceFunc: func(dogma.Message) string {
						return "<instance>"
					},
					HandleCommandFunc: func(
						_ dogma.AggregateRoot,
						s dogma.AggregateCommandScope,
						_ dogma.Message,
					) {
						s.RecordEvent(MessageN1)
					},
				})

				c.RegisterProcess(&ProcessMessageHandler{
					ConfigureFunc: func(c dogma.ProcessConfigurer) {
						c.Identity("<process>", "8b4c4701-be92-4b28-83b6-0d69b97fb451")
						c.Routes(
							dogma.HandlesEvent[MessageE](),     // E = event
							dogma.HandlesEvent[MessageN](),     // N = (do) nothing
							dogma.ExecutesCommand[MessageC](),  // C = command
							dogma.ExecutesCommand[*MessageC](), // pointer, used to test type similarity
							dogma.ExecutesCommand[MessageX](),
						)

					},
					RouteEventToInstanceFunc: func(
						context.Context,
						dogma.Message,
					) (string, bool, error) {
						return "<instance>", true, nil
					},
					HandleEventFunc: func(
						_ context.Context,
						_ dogma.ProcessRoot,
						s dogma.ProcessEventScope,
						m dogma.Message,
					) error {
						if _, ok := m.(MessageE); ok {
							s.ExecuteCommand(MessageC1)
						}
						return nil
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
			Expect(testingT.Failed()).To(Equal(!ok))
		},
		g.Entry(
			"command executed as expected",
			RecordEvent(MessageE1),
			ToExecuteCommand(MessageC1),
			expectPass,
			expectReport(
				`✓ execute a specific 'fixtures.MessageC' command`,
			),
		),
		g.Entry(
			"no matching command executed",
			RecordEvent(MessageE1),
			ToExecuteCommand(MessageX1),
			expectFail,
			expectReport(
				`✗ execute a specific 'fixtures.MessageX' command`,
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
			RecordEvent(MessageN1),
			ToExecuteCommand(MessageX1),
			expectFail,
			expectReport(
				`✗ execute a specific 'fixtures.MessageX' command`,
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
			ExecuteCommand(MessageR1),
			ToExecuteCommand(MessageC1),
			expectFail,
			expectReport(
				`✗ execute a specific 'fixtures.MessageC' command`,
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
			ExecuteCommand(MessageR1),
			ToExecuteCommand(MessageX1),
			expectFail,
			expectReport(
				`✗ execute a specific 'fixtures.MessageX' command`,
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
			RecordEvent(MessageE1),
			ToExecuteCommand(&MessageC1), // note: message type is pointer
			expectFail,
			expectReport(
				`✗ execute a specific '*fixtures.MessageC' command`,
				``,
				`  | EXPLANATION`,
				`  |     a command of a similar type was executed by the '<process>' process message handler`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • check the message type, should it be a pointer?`,
				`  | `,
				`  | MESSAGE DIFF`,
				`  |     [-*-]fixtures.MessageC{`,
				`  |         Value: "C1"`,
				`  |     }`,
			),
		),
		g.Entry(
			"similar command executed with a different value",
			RecordEvent(MessageE1),
			ToExecuteCommand(MessageC2),
			expectFail,
			expectReport(
				`✗ execute a specific 'fixtures.MessageC' command`,
				``,
				`  | EXPLANATION`,
				`  |     a similar command was executed by the '<process>' process message handler`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • check the content of the message`,
				`  | `,
				`  | MESSAGE DIFF`,
				`  |     fixtures.MessageC{`,
				`  |         Value: "C[-2-]{+1+}"`,
				`  |     }`,
			),
		),
	)

	g.It("fails the test if the message type is unrecognized", func() {
		test := Begin(testingT, app)
		test.Expect(
			noop,
			ToExecuteCommand(MessageU1),
		)

		Expect(testingT.Failed()).To(BeTrue())
		Expect(testingT.Logs).To(ContainElement(
			"a command of type fixtures.MessageU can never be executed, the application does not use this message type",
		))
	})

	g.It("fails the test if the message type is not a command", func() {
		test := Begin(testingT, app)
		test.Expect(
			noop,
			ToExecuteCommand(MessageE1),
		)

		Expect(testingT.Failed()).To(BeTrue())
		Expect(testingT.Logs).To(ContainElement(
			"fixtures.MessageE is an event, it can never be executed as a command",
		))
	})

	g.It("fails the test if the message type is not produced by any handlers", func() {
		test := Begin(testingT, app)
		test.Expect(
			noop,
			ToExecuteCommand(MessageR1),
		)

		Expect(testingT.Failed()).To(BeTrue())
		Expect(testingT.Logs).To(ContainElement(
			"no handlers execute commands of type fixtures.MessageR, it is only ever consumed",
		))
	})

	g.It("panics if the message is invalid", func() {
		Expect(func() {
			ToExecuteCommand(nil)
		}).To(PanicWith("ToExecuteCommand(<nil>): message must not be nil"))
	})
})
