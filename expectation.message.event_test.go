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

var _ = g.Describe("func ToExecuteCommandType()", func() {
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
				c.Identity("<app>", "adb2ed1e-b1f4-4756-abfa-a5e3a3e08def")

				c.RegisterAggregate(&AggregateMessageHandler{
					ConfigureFunc: func(c dogma.AggregateConfigurer) {
						c.Identity("<aggregate>", "8746651e-df4d-421c-9eea-177585e5b8eb")
						c.Routes(
							dogma.HandlesCommand[MessageR](), // R = record an event
							dogma.HandlesCommand[MessageN](), // N = do nothing
							dogma.RecordsEvent[MessageE](),
							dogma.RecordsEvent[*MessageE](), // pointer, used to test type similarity
							dogma.RecordsEvent[MessageX](),
						)
					},
					RouteCommandToInstanceFunc: func(dogma.Event) string {
						return "<instance>"
					},
					HandleCommandFunc: func(
						_ dogma.AggregateRoot,
						s dogma.AggregateCommandScope,
						m dogma.Event,
					) {
						if _, ok := m.(MessageR); ok {
							s.RecordEvent(MessageE1)
						}
					},
				})

				c.RegisterProcess(&ProcessMessageHandler{
					ConfigureFunc: func(c dogma.ProcessConfigurer) {
						c.Identity("<process>", "209c7f0f-49ad-4419-88a6-4e9ee1cf204a")
						c.Routes(
							dogma.HandlesEvent[MessageE](), // E = execute a command
							dogma.HandlesEvent[MessageO](), // O = only consumed, never produced
							dogma.ExecutesCommand[MessageN](),
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
			"event recorded as expected",
			ExecuteCommand(MessageR1),
			ToRecordEvent(MessageE1),
			expectPass,
			expectReport(
				`✓ record a specific 'fixtures.MessageE' event`,
			),
		),
		g.Entry(
			"no matching event recorded",
			ExecuteCommand(MessageR1),
			ToRecordEvent(MessageX1),
			expectFail,
			expectReport(
				`✗ record a specific 'fixtures.MessageX' event`,
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
			"no matching event recorded and all relevant handler types disabled",
			ExecuteCommand(MessageR1),
			ToRecordEvent(MessageX1),
			expectFail,
			expectReport(
				`✗ record a specific 'fixtures.MessageX' event`,
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
			"no matching event recorded and no relevant handler types engaged",
			ExecuteCommand(MessageR1),
			ToRecordEvent(MessageX1),
			expectFail,
			expectReport(
				`✗ record a specific 'fixtures.MessageX' event`,
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
			ToRecordEvent(MessageX1),
			expectFail,
			expectReport(
				`✗ record a specific 'fixtures.MessageX' event`,
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
			ToRecordEvent(MessageX1),
			expectFail,
			expectReport(
				`✗ record a specific 'fixtures.MessageX' event`,
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
			"similar event recorded with a different type",
			ExecuteCommand(MessageR1),
			ToRecordEvent(&MessageE1), // note: message type is pointer
			expectFail,
			expectReport(
				`✗ record a specific '*fixtures.MessageE' event`,
				``,
				`  | EXPLANATION`,
				`  |     an event of a similar type was recorded by the '<aggregate>' aggregate message handler`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • check the message type, should it be a pointer?`,
				`  | `,
				`  | MESSAGE DIFF`,
				`  |     [-*-]fixtures.MessageE{`,
				`  |         Value: "E1"`,
				`  |     }`,
			),
		),
		g.Entry(
			"similar event recorded with a different value",
			ExecuteCommand(MessageR1),
			ToRecordEvent(MessageE2),
			expectFail,
			expectReport(
				`✗ record a specific 'fixtures.MessageE' event`,
				``,
				`  | EXPLANATION`,
				`  |     a similar event was recorded by the '<aggregate>' aggregate message handler`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • check the content of the message`,
				`  | `,
				`  | MESSAGE DIFF`,
				`  |     fixtures.MessageE{`,
				`  |         Value: "E[-2-]{+1+}"`,
				`  |     }`,
			),
		),
		g.Entry(
			"does not include an explanation when negated and a sibling expectation passes",
			ExecuteCommand(MessageR1),
			NoneOf(
				ToRecordEvent(MessageE1),
				ToRecordEvent(MessageE2),
			),
			expectFail,
			expectReport(
				`✗ none of (1 of the expectations passed unexpectedly)`,
				`    ✓ record a specific 'fixtures.MessageE' event`,
				`    ✗ record a specific 'fixtures.MessageE' event`,
			),
		),
	)

	g.It("fails the test if the message type is unrecognized", func() {
		test := Begin(testingT, app)
		test.Expect(
			noop,
			ToRecordEvent(MessageU1),
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
			ToRecordEvent(MessageR1),
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
			ToRecordEvent(MessageO1),
		)

		Expect(testingT.Failed()).To(BeTrue())
		Expect(testingT.Logs).To(ContainElement(
			"no handlers record events of type fixtures.MessageO, it is only ever consumed",
		))
	})

	g.It("panics if the message is nil", func() {
		Expect(func() {
			ToRecordEvent(nil)
		}).To(PanicWith("ToRecordEvent(<nil>): message must not be nil"))
	})
})
