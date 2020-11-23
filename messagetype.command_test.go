package testkit_test

import (
	"context"

	"github.com/dogmatiq/testkit/engine"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogma/fixtures"
	"github.com/dogmatiq/testkit"
	. "github.com/dogmatiq/testkit"
	"github.com/dogmatiq/testkit/internal/testingmock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("func ToExecuteCommandOfType() (commands produced by handlers)", func() {
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
						c.ProducesEventType(MessageN{})
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
						c.Identity("<process>", "<process-key>")
						c.ConsumesEventType(MessageE{})   // E = event
						c.ConsumesEventType(MessageN{})   // N = (do) nothing
						c.ProducesCommandType(MessageC{}) // C = command
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
							s.ExecuteCommand(MessageC1)
						}
						return nil
					},
				})
			},
		}
	})

	testExpectationBehavior := func(
		a Action,
		e Expectation,
		ok bool,
		rm reportMatcher,
		options ...TestOption,
	) {
		test = testkit.Begin(testingT, app, options...)
		test.Expect(a, e)
		rm(testingT)
		Expect(testingT.Failed()).To(Equal(!ok))
	}

	DescribeTable(
		"expectation behavior",
		testExpectationBehavior,
		Entry(
			"it passes if the expected command type is executed",
			RecordEvent(MessageE1),
			ToExecuteCommandOfType(MessageC{}),
			expectPass,
			expectReport(
				`✓ execute any 'fixtures.MessageC' command`,
			),
		),
		Entry(
			"it fails if none of the engaged handlers executed a command of the expected type",
			RecordEvent(MessageE1),
			ToExecuteCommandOfType(MessageX{}),
			expectFail,
			expectReport(
				`✗ execute any 'fixtures.MessageX' command`,
				``,
				`  | EXPLANATION`,
				`  |     none of the engaged handlers executed the expected command`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • verify the logic within the '<process>' process message handler`,
			),
		),
		Entry(
			"it fails if none of the engaged handlers produced any messages at all",
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
		Entry(
			"it fails if none of the engaged handlers execute any commands at all",
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
		Entry(
			"it fails if none of the relevant handler types are enabled",
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
		Entry(
			"it fails if a command of a similar type is executed",
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
		Entry(
			"expected message type recorded as an event rather than executed as a command",
			ExecuteCommand(MessageR1),
			ToExecuteCommandOfType(MessageN{}),
			expectFail,
			expectReport(
				`✗ execute any 'fixtures.MessageN' command`,
				``,
				`  | EXPLANATION`,
				`  |     a message of this type was recorded as an event by the '<aggregate>' aggregate message handler`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • verify that the '<aggregate>' aggregate message handler intended to record an event of this type`,
				`  |     • verify that ToExecuteCommandOfType() is the correct assertion, did you mean ToRecordEventOfType()?`,
			),
		),
		Entry(
			"a message with a similar type recorded as an event rather than executed as a command",
			ExecuteCommand(MessageR1),
			ToExecuteCommandOfType(&MessageN{}), // note: message type is pointer
			expectFail,
			expectReport(
				`✗ execute any '*fixtures.MessageN' command`,
				``,
				`  | EXPLANATION`,
				`  |     a message of a similar type was recorded as an event by the '<aggregate>' aggregate message handler`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • verify that the '<aggregate>' aggregate message handler intended to record an event of this type`,
				`  |     • verify that ToExecuteCommandOfType() is the correct assertion, did you mean ToRecordEventOfType()?`,
				`  |     • check the message type, should it be a pointer?`,
				`  | `,
				`  | MESSAGE TYPE DIFF`,
				`  |     [-*-]fixtures.MessageN`,
			),
		),
	)
})
