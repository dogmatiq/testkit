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

var _ = g.Describe("func ToExecuteCommandOfType()", func() {
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
				c.Identity("<app>", "936ab3fa-f379-42e7-9100-a4d28accc932")

				c.RegisterAggregate(&AggregateMessageHandler{
					ConfigureFunc: func(c dogma.AggregateConfigurer) {
						c.Identity("<aggregate>", "c0094d00-d270-4fa0-bb4b-3cb24060cbe2")
						c.Routes(
							dogma.HandlesCommand[MessageR](), // R = record an event
							dogma.RecordsEvent[MessageN](),
						)
					},
					RouteCommandToInstanceFunc: func(dogma.Command) string {
						return "<instance>"
					},
					HandleCommandFunc: func(
						_ dogma.AggregateRoot,
						s dogma.AggregateCommandScope,
						_ dogma.Command,
					) {
						s.RecordEvent(MessageN1)
					},
				})

				c.RegisterProcess(&ProcessMessageHandler{
					ConfigureFunc: func(c dogma.ProcessConfigurer) {
						c.Identity("<process>", "551c1b77-4c4c-4b0c-b97f-b7c1ff6eada2")
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
			"command type executed as expected",
			RecordEvent(MessageE1),
			ToExecuteCommandOfType(MessageC{}),
			expectPass,
			expectReport(
				`✓ execute any 'fixtures.MessageC' command`,
			),
		),
		g.Entry(
			"no matching command types executed",
			RecordEvent(MessageE1),
			ToExecuteCommandOfType(MessageX{}),
			expectFail,
			expectReport(
				`✗ execute any 'fixtures.MessageX' command`,
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
			ToExecuteCommandOfType(MessageC{}),
			expectFail,
			expectReport(
				`✗ execute any 'fixtures.MessageC' command`,
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
			ToExecuteCommandOfType(MessageC{}),
			expectFail,
			expectReport(
				`✗ execute any 'fixtures.MessageC' command`,
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
			RecordEvent(MessageN1),
			ToExecuteCommandOfType(MessageC{}),
			expectFail,
			expectReport(
				`✗ execute any 'fixtures.MessageC' command`,
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
			"command of a similar type executed",
			RecordEvent(MessageE1),
			ToExecuteCommandOfType(&MessageC{}), // note: message type is pointer
			expectFail,
			expectReport(
				`✗ execute any '*fixtures.MessageC' command`,
				``,
				`  | EXPLANATION`,
				`  |     a command of a similar type was executed by the '<process>' process message handler`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • check the message type, should it be a pointer?`,
				`  | `,
				`  | MESSAGE TYPE DIFF`,
				`  |     [-*-]fixtures.MessageC`,
			),
		),
		g.Entry(
			"does not include an explanation when negated and a sibling expectation passes",
			RecordEvent(MessageE1),
			NoneOf(
				ToExecuteCommandOfType(MessageC{}),
				ToExecuteCommandOfType(MessageX{}),
			),
			expectFail,
			expectReport(
				`✗ none of (1 of the expectations passed unexpectedly)`,
				`    ✓ execute any 'fixtures.MessageC' command`,
				`    ✗ execute any 'fixtures.MessageX' command`,
			),
		),
	)

	g.It("fails the test if the message type is unrecognized", func() {
		test := Begin(testingT, app)
		test.Expect(
			noop,
			ToExecuteCommandOfType(MessageU{}),
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
			ToExecuteCommandOfType(MessageE{}),
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
			ToExecuteCommandOfType(MessageR{}),
		)

		Expect(testingT.Failed()).To(BeTrue())
		Expect(testingT.Logs).To(ContainElement(
			"no handlers execute commands of type fixtures.MessageR, it is only ever consumed",
		))
	})

	g.It("panics if the message is nil", func() {
		Expect(func() {
			ToExecuteCommandOfType(nil)
		}).To(PanicWith("ToExecuteCommandOfType(<nil>): message must not be nil"))
	})
})
