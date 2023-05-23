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

var _ = g.Describe("func ToRecordEvent() (when used with the Call() action)", func() {
	var (
		testingT *testingmock.T
		app      dogma.Application
		test     *Test
	)

	g.BeforeEach(func() {
		testingT = &testingmock.T{
			FailSilently: true,
		}

		app = &Application{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "58067ffe-c1d6-4097-8acb-c55a7936cb4b")

				c.RegisterAggregate(&AggregateMessageHandler{
					ConfigureFunc: func(c dogma.AggregateConfigurer) {
						c.Identity("<aggregate>", "eb9a40bd-6edf-4922-88dd-dfbdad42c88a")
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
						c.Identity("<process>", "c8e3b9de-c8e5-4e91-8658-58c9e989b048")
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

	g.DescribeTable(
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
		g.Entry(
			"event recorded as expected",
			recordEventViaRecorder(MessageE1),
			ToRecordEvent(MessageE1),
			expectPass,
			expectReport(
				`✓ record a specific 'fixtures.MessageE' event`,
			),
		),
		g.Entry(
			"no matching event recorded",
			executeCommandViaExecutor(MessageR1),
			ToRecordEvent(MessageX1),
			expectFail,
			expectReport(
				`✗ record a specific 'fixtures.MessageX' event`,
				``,
				`  | EXPLANATION`,
				`  |     nothing recorded a matching event`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • enable integration handlers using the EnableHandlerType() option`,
				`  |     • verify the logic within the '<aggregate>' aggregate message handler`,
				`  |     • verify the logic within the code that uses the dogma.EventRecorder`,
			),
		),
		g.Entry(
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
		g.Entry(
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
		g.Entry(
			"no matching event recorded and all relevant handler types disabled",
			recordEventViaRecorder(MessageA1),
			ToRecordEvent(MessageE1),
			expectFail,
			expectReport(
				`✗ record a specific 'fixtures.MessageE' event`,
				``,
				`  | EXPLANATION`,
				`  |     nothing recorded a matching event`,
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
		g.Entry(
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
		g.Entry(
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
