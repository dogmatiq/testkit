package testkit_test

import (
	"context"
	"errors"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogma/fixtures"
	. "github.com/dogmatiq/testkit"
	"github.com/dogmatiq/testkit/engine"
	"github.com/dogmatiq/testkit/internal/testingmock"
	g "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = g.Describe("func ToRecordEventMatching()", func() {
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
				c.Identity("<app>", "43962b88-b25c-4a59-938e-64540c473a7c")

				c.RegisterAggregate(&AggregateMessageHandler{
					ConfigureFunc: func(c dogma.AggregateConfigurer) {
						c.Identity("<aggregate>", "8305181b-b87f-4446-8f56-a3c38d5bd32f")
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
						if m, ok := m.(MessageR); ok {
							s.RecordEvent(MessageE1)

							if m.Value == "<multiple>" {
								s.RecordEvent(MessageE1)
							}
						}
					},
				})

				c.RegisterProcess(&ProcessMessageHandler{
					ConfigureFunc: func(c dogma.ProcessConfigurer) {
						c.Identity("<process>", "0bd19cfa-c910-453f-a6cc-959fdce9c34f")
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
			"matching event recorded as expected",
			ExecuteCommand(MessageR1),
			ToRecordEventMatching(
				func(m dogma.Message) error {
					if m == MessageE1 {
						return nil
					}

					return errors.New("<error>")
				},
			),
			expectPass,
			expectReport(
				`✓ record an event that matches the predicate near expectation.messagematch.event_test.go:109`,
			),
		),
		g.Entry(
			"no matching event recorded",
			ExecuteCommand(MessageR1),
			ToRecordEventMatching(
				func(m dogma.Message) error {
					if m == MessageX1 {
						return nil
					}

					return errors.New("<error>")
				},
			),
			expectFail,
			expectReport(
				`✗ record an event that matches the predicate near expectation.messagematch.event_test.go:126`,
				``,
				`  | EXPLANATION`,
				`  |     none of the engaged handlers recorded a matching event`,
				`  | `,
				`  | FAILED MATCHES`,
				`  |     • fixtures.MessageE: <error>`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • verify the logic within the predicate function`,
				`  |     • enable integration handlers using the EnableHandlerType() option`,
				`  |     • verify the logic within the '<aggregate>' aggregate message handler`,
			),
		),
		g.Entry(
			"no messages produced at all",
			ExecuteCommand(MessageN1),
			ToRecordEventMatching(
				func(m dogma.Message) error {
					panic("unexpected call")
				},
			),
			expectFail,
			expectReport(
				`✗ record an event that matches the predicate near expectation.messagematch.event_test.go:154`,
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
			ToRecordEventMatching(
				func(m dogma.Message) error {
					return IgnoreMessage
				},
			),
			expectFail,
			expectReport(
				`✗ record an event that matches the predicate near expectation.messagematch.event_test.go:174`,
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
			"no matching event recorded and all relevant handler types disabled",
			ExecuteCommand(MessageR1),
			ToRecordEventMatching(
				func(m dogma.Message) error {
					panic("unexpected call")
				},
			),
			expectFail,
			expectReport(
				`✗ record an event that matches the predicate near expectation.messagematch.event_test.go:194`,
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
			"no matching event recorded and events were ignored",
			ExecuteCommand(MessageR1),
			ToRecordEventMatching(
				func(m dogma.Message) error {
					return IgnoreMessage
				},
			),
			expectFail,
			expectReport(
				`✗ record an event that matches the predicate near expectation.messagematch.event_test.go:219`,
				``,
				`  | EXPLANATION`,
				`  |     none of the engaged handlers recorded a matching event`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • verify the logic within the predicate function, it ignored 1 event`,
				`  |     • enable integration handlers using the EnableHandlerType() option`,
				`  |     • verify the logic within the '<aggregate>' aggregate message handler`,
			),
		),
		g.Entry(
			"no matching event recorded and error messages were repeated",
			ExecuteCommand(MessageR{
				Value: "<multiple>", // trigger multiple MessageE events
			}),
			ToRecordEventMatching(
				func(m dogma.Message) error {
					return errors.New("<error>")
				},
			),
			expectFail,
			expectReport(
				`✗ record an event that matches the predicate near expectation.messagematch.event_test.go:242`,
				``,
				`  | EXPLANATION`,
				`  |     none of the engaged handlers recorded a matching event`,
				`  | `,
				`  | FAILED MATCHES`,
				`  |     • fixtures.MessageE: <error> (repeated 2 times)`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • verify the logic within the predicate function`,
				`  |     • enable integration handlers using the EnableHandlerType() option`,
				`  |     • verify the logic within the '<aggregate>' aggregate message handler`,
			),
		),
	)

	g.It("panics if the function is nil", func() {
		Expect(func() {
			ToRecordEventMatching(nil)
		}).To(PanicWith("ToRecordEventMatching(<nil>): function must not be nil"))
	})
})
