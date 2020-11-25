package testkit_test

import (
	"context"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogma/fixtures"
	. "github.com/dogmatiq/testkit"
	"github.com/dogmatiq/testkit/engine"
	"github.com/dogmatiq/testkit/internal/testingmock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("func ToExecuteCommandOfType()", func() {
	var (
		testingT *testingmock.T
		app      dogma.Application
		test     *Test
	)

	BeforeEach(func() {
		testingT = &testingmock.T{
			FailSilently: true,
		}

		app = &Application{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "<app-key>")

				c.RegisterAggregate(&AggregateMessageHandler{
					ConfigureFunc: func(c dogma.AggregateConfigurer) {
						c.Identity("<aggregate>", "<aggregate-key>")
						c.ConsumesCommandType(MessageR{}) // R = record an event
						c.ConsumesCommandType(MessageN{}) // N = do nothing
						c.ProducesEventType(MessageE{})
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
						c.Identity("<process>", "<process-key>")
						c.ConsumesEventType(MessageE{}) // E = execute a command
						c.ProducesCommandType(MessageN{})
					},
					RouteEventToInstanceFunc: func(
						context.Context,
						dogma.Message,
					) (string, bool, error) {
						return "<instance>", true, nil
					},
					HandleEventFunc: func(
						_ context.Context,
						s dogma.ProcessEventScope,
						m dogma.Message,
					) error {
						if _, ok := m.(MessageE); ok {
							s.Begin()
							s.ExecuteCommand(MessageN1)
						}
						return nil
					},
				})
			},
		}
	})

	DescribeTable(
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
		Entry(
			"event recorded as expected",
			ExecuteCommand(MessageR1),
			ToRecordEvent(MessageE1),
			expectPass,
			expectReport(
				`✓ record a specific 'fixtures.MessageE' event`,
			),
		),
		Entry(
			"no matching event recorded",
			ExecuteCommand(MessageR1),
			ToRecordEvent(MessageX1),
			expectFail,
			expectReport(
				`✗ record a specific 'fixtures.MessageX' event`,
				``,
				`  | EXPLANATION`,
				`  |     none of the engaged handlers recorded the expected event`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • enable integration handlers using the EnableHandlerType() option`,
				`  |     • verify the logic within the '<aggregate>' aggregate message handler`,
			),
		),
		Entry(
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
		Entry(
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
		Entry(
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
		Entry(
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
		Entry(
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
		Entry(
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
		Entry(
			"expected message executed as a command rather than recorded as an event",
			RecordEvent(MessageE1),
			ToRecordEvent(MessageN1),
			expectFail,
			expectReport(
				`✗ record a specific 'fixtures.MessageN' event`,
				``,
				`  | EXPLANATION`,
				`  |     the expected message was executed as a command by the '<process>' process message handler`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • verify that the '<process>' process message handler intended to execute a command of this type`,
				`  |     • verify that ToRecordEvent() is the correct expectation, did you mean ToExecuteCommand()?`,
			),
		),
		Entry(
			"similar message with a different value executed as a command rather than recorded as an event",
			ExecuteCommand(MessageR1),
			ToRecordEvent(MessageN2),
			expectFail,
			expectReport(
				`✗ record a specific 'fixtures.MessageN' event`,
				``,
				`  | EXPLANATION`,
				`  |     a similar message was executed as a command by the '<process>' process message handler`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • verify that the '<process>' process message handler intended to execute a command of this type`,
				`  |     • verify that ToRecordEvent() is the correct expectation, did you mean ToExecuteCommand()?`,
				`  | `,
				`  | MESSAGE DIFF`,
				`  |     fixtures.MessageN{`,
				`  |         Value: "N[-2-]{+1+}"`,
				`  |     }`,
			),
		),
		Entry(
			"similar message with a different type executed as a command rather than recorded as an event",
			ExecuteCommand(MessageR1),
			ToRecordEvent(&MessageN1), // note: message type is pointer
			expectFail,
			expectReport(
				`✗ record a specific '*fixtures.MessageN' event`,
				``,
				`  | EXPLANATION`,
				`  |     a message of a similar type was executed as a command by the '<process>' process message handler`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • verify that the '<process>' process message handler intended to execute a command of this type`,
				`  |     • verify that ToRecordEvent() is the correct expectation, did you mean ToExecuteCommand()?`,
				`  |     • check the message type, should it be a pointer?`,
				`  | `,
				`  | MESSAGE DIFF`,
				`  |     [-*-]fixtures.MessageN{`,
				`  |         Value: "N1"`,
				`  |     }`,
			),
		),
	)

	It("panics if the message is nil", func() {
		Expect(func() {
			ToRecordEvent(nil)
		}).To(PanicWith("ToRecordEvent(<nil>): message must not be nil"))
	})
})
