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

var _ = Describe("func ToRecordEvent() (when used with the Call() action)", func() {
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
						c.ProducesEventType(&MessageE{}) // pointer, used to test type similarity
						c.ProducesEventType(MessageX{})
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
						c.ConsumesEventType(MessageA{}) // A = also execute a command
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

	executeCommandViaExecutor := func(m dogma.Message) Action {
		return Call(func() {
			err := test.CommandExecutor().ExecuteCommand(context.Background(), m)
			Expect(err).ShouldNot(HaveOccurred())
		})
	}

	recordEventViaRecorder := func(m dogma.Message) Action {
		return Call(func() {
			err := test.EventRecorder().RecordEvent(context.Background(), m)
			Expect(err).ShouldNot(HaveOccurred())
		})
	}

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
			recordEventViaRecorder(MessageE1),
			ToRecordEvent(MessageE1),
			expectPass,
			expectReport(
				`✓ record a specific 'fixtures.MessageE' event`,
			),
		),
		Entry(
			"no matching event recorded",
			executeCommandViaExecutor(MessageR1),
			ToRecordEvent(MessageX1),
			expectFail,
			expectReport(
				`✗ record a specific 'fixtures.MessageX' event`,
				``,
				`  | EXPLANATION`,
				`  |     nothing recorded the expected event`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • enable integration handlers using the EnableHandlerType() option`,
				`  |     • verify the logic within the '<aggregate>' aggregate message handler`,
				`  |     • verify the logic within the code that uses the dogma.EventRecorder`,
			),
		),
		Entry(
			"no messages produced at all",
			Call(func() {}),
			ToRecordEvent(MessageE1),
			expectFail,
			expectReport(
				`✗ record a specific 'fixtures.MessageE' event`,
				``,
				`  | EXPLANATION`,
				`  |     no messages were produced at all`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • verify the logic within the code that uses the dogma.EventRecorder`,
			),
		),
		Entry(
			"no events recorded at all",
			executeCommandViaExecutor(MessageN1),
			ToRecordEvent(MessageE1),
			expectFail,
			expectReport(
				`✗ record a specific 'fixtures.MessageE' event`,
				``,
				`  | EXPLANATION`,
				`  |     no events were recorded at all`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • enable integration handlers using the EnableHandlerType() option`,
				`  |     • verify the logic within the '<aggregate>' aggregate message handler`,
				`  |     • verify the logic within the code that uses the dogma.EventRecorder`,
			),
		),
		Entry(
			"no matching event recorded and all relevant handler types disabled",
			recordEventViaRecorder(MessageA1),
			ToRecordEvent(MessageE1),
			expectFail,
			expectReport(
				`✗ record a specific 'fixtures.MessageE' event`,
				``,
				`  | EXPLANATION`,
				`  |     nothing recorded the expected event`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • enable aggregate handlers using the EnableHandlerType() option`,
				`  |     • enable integration handlers using the EnableHandlerType() option`,
				`  |     • verify the logic within the code that uses the dogma.EventRecorder`,
			),
			WithUnsafeOperationOptions(
				engine.EnableAggregates(false),
			),
		),
		Entry(
			"similar event recorded with a different type",
			recordEventViaRecorder(MessageE1),
			ToRecordEvent(&MessageE1), // note: message type is pointer
			expectFail,
			expectReport(
				`✗ record a specific '*fixtures.MessageE' event`,
				``,
				`  | EXPLANATION`,
				`  |     an event of a similar type was recorded via a dogma.EventRecorder`,
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
			recordEventViaRecorder(MessageE1),
			ToRecordEvent(MessageE2),
			expectFail,
			expectReport(
				`✗ record a specific 'fixtures.MessageE' event`,
				``,
				`  | EXPLANATION`,
				`  |     a similar event was recorded via a dogma.EventRecorder`,
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
	)
})
