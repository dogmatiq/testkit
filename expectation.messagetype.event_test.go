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
				c.Identity("<app>", "ef25ca55-2ace-40b5-9c2d-c53f5a80908a")

				c.RegisterAggregate(&AggregateMessageHandler{
					ConfigureFunc: func(c dogma.AggregateConfigurer) {
						c.Identity("<aggregate>", "bef4f7fd-cd2a-4e10-9c52-691d389847b3")
						c.Routes(
							dogma.HandlesCommand[MessageR](), // R = record an event
							dogma.HandlesCommand[MessageN](), // N = do nothing
							dogma.RecordsEvent[MessageE](),
							dogma.RecordsEvent[*MessageE](), // pointer, used to test type similarity
							dogma.RecordsEvent[MessageX](),
						)
					},
					RouteCommandToInstanceFunc: func(dogma.Message) string {
						return "<instance>"
					},
					HandleCommandFunc: func(
						_ dogma.AggregateRoot,
						s dogma.AggregateCommandScope,
						m dogma.Message,
					) {
						if _, ok := m.(MessageR); ok {
							s.RecordEvent(MessageE1)
						}
					},
				})

				c.RegisterProcess(&ProcessMessageHandler{
					ConfigureFunc: func(c dogma.ProcessConfigurer) {
						c.Identity("<process>", "6a2146fa-00e5-49a3-acd8-fb0a6451a8ff")
						c.Routes(
							dogma.HandlesEvent[MessageE](), // E = execute a command
							dogma.HandlesEvent[MessageO](), // O = only consumed, never produced
							dogma.ExecutesCommand[MessageN](),
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
							s.ExecuteCommand(MessageN1)
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
			"event type recorded as expected",
			ExecuteCommand(MessageR1),
			ToRecordEventOfType(MessageE{}),
			expectPass,
			expectReport(
				`✓ record any 'fixtures.MessageE' event`,
			),
		),
		g.Entry(
			"no matching event type recorded",
			ExecuteCommand(MessageR1),
			ToRecordEventOfType(MessageX{}),
			expectFail,
			expectReport(
				`✗ record any 'fixtures.MessageX' event`,
				``,
				`  | EXPLANATION`,
				`  |     none of the engaged handlers recorded a matching event`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • enable integration handlers using the EnableHandlerType() option`,
				`  |     • verify the logic within the '<aggregate>' aggregate message handler`,
			),
		),
		g.Entry(
			"no matching event type recorded and all relevant handler types disabled",
			ExecuteCommand(MessageR1),
			ToRecordEventOfType(MessageX{}),
			expectFail,
			expectReport(
				`✗ record any 'fixtures.MessageX' event`,
				``,
				`  | EXPLANATION`,
				`  |     no relevant handler types were enabled`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • enable aggregate handlers using the EnableHandlerType() option`,
				`  |     • enable integration handlers using the EnableHandlerType() option`,
			),
			WithUnsafeOperationOptions(
				engine.EnableAggregates(false),
				engine.EnableIntegrations(false),
			),
		),
		g.Entry(
			"no matching event type recorded and no relevant handler types engaged",
			ExecuteCommand(MessageR1),
			ToRecordEventOfType(MessageX{}),
			expectFail,
			expectReport(
				`✗ record any 'fixtures.MessageX' event`,
				``,
				`  | EXPLANATION`,
				`  |     no relevant handlers (aggregate or integration) were engaged`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • enable aggregate handlers using the EnableHandlerType() option`,
				`  |     • check the application's routing configuration`,
			),
			WithUnsafeOperationOptions(
				engine.EnableAggregates(false),
				engine.EnableIntegrations(true),
			),
		),
		g.Entry(
			"no messages produced at all",
			ExecuteCommand(MessageN1),
			ToRecordEventOfType(MessageX{}),
			expectFail,
			expectReport(
				`✗ record any 'fixtures.MessageX' event`,
				``,
				`  | EXPLANATION`,
				`  |     no messages were produced at all`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • enable integration handlers using the EnableHandlerType() option`,
				`  |     • verify the logic within the '<aggregate>' aggregate message handler`,
			),
		),
		g.Entry(
			"no events recorded at all",
			RecordEvent(MessageE1),
			ToRecordEventOfType(MessageX{}),
			expectFail,
			expectReport(
				`✗ record any 'fixtures.MessageX' event`,
				``,
				`  | EXPLANATION`,
				`  |     no events were recorded at all`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • enable integration handlers using the EnableHandlerType() option`,
				`  |     • verify the logic within the '<aggregate>' aggregate message handler`,
			),
		),
		g.Entry(
			"event of a similar type recorded",
			ExecuteCommand(MessageR1),
			ToRecordEventOfType(&MessageE{}), // note: message type is pointer
			expectFail,
			expectReport(
				`✗ record any '*fixtures.MessageE' event`,
				``,
				`  | EXPLANATION`,
				`  |     an event of a similar type was recorded by the '<aggregate>' aggregate message handler`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • check the message type, should it be a pointer?`,
				`  | `,
				`  | MESSAGE TYPE DIFF`,
				`  |     [-*-]fixtures.MessageE`,
			),
		),
	)

	g.It("fails the test if the message type is unrecognized", func() {
		test := Begin(testingT, app)
		test.Expect(
			noop,
			ToRecordEventOfType(MessageU{}),
		)

		Expect(testingT.Failed()).To(BeTrue())
		Expect(testingT.Logs).To(ContainElement(
			"an event of type fixtures.MessageU can never be recorded, the application does not use this message type",
		))
	})

	g.It("fails the test if the message type is not an event", func() {
		test := Begin(testingT, app)
		test.Expect(
			noop,
			ToRecordEventOfType(MessageR{}),
		)

		Expect(testingT.Failed()).To(BeTrue())
		Expect(testingT.Logs).To(ContainElement(
			"fixtures.MessageR is a command, it can never be recorded as an event",
		))
	})

	g.It("fails the test if the message type is not produced by any handlers", func() {
		test := Begin(testingT, app)
		test.Expect(
			noop,
			ToRecordEventOfType(MessageO{}),
		)

		Expect(testingT.Failed()).To(BeTrue())
		Expect(testingT.Logs).To(ContainElement(
			"no handlers record events of type fixtures.MessageO, it is only ever consumed",
		))
	})

	g.It("panics if the message is nil", func() {
		Expect(func() {
			ToRecordEventOfType(nil)
		}).To(PanicWith("ToRecordEventOfType(<nil>): message must not be nil"))
	})
})
