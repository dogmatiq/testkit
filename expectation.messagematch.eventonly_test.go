package testkit_test

import (
	"errors"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/dogma/fixtures"
	. "github.com/dogmatiq/dogma/fixtures"
	. "github.com/dogmatiq/enginekit/enginetest/stubs"
	. "github.com/dogmatiq/testkit"
	"github.com/dogmatiq/testkit/internal/testingmock"
	g "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = g.Describe("func ToOnlyRecordEventsMatching()", func() {
	var (
		testingT *testingmock.T
		app      dogma.Application
	)

	g.BeforeEach(func() {
		testingT = &testingmock.T{
			FailSilently: true,
		}

		app = &ApplicationStub{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "94f425c5-339a-4213-8309-16234225480e")

				c.RegisterAggregate(&AggregateMessageHandlerStub{
					ConfigureFunc: func(c dogma.AggregateConfigurer) {
						c.Identity("<aggregate>", "bc64cfe4-3339-4eee-a9d2-364d33dff47d")
						c.Routes(
							dogma.HandlesCommand[MessageC](), // C = command
							dogma.RecordsEvent[MessageE](),   // E = event
						)
					},
					RouteCommandToInstanceFunc: func(dogma.Command) string {
						return "<instance>"
					},
					HandleCommandFunc: func(
						_ dogma.AggregateRoot,
						s dogma.AggregateCommandScope,
						m dogma.Command,
					) {
						s.RecordEvent(MessageE1)
						s.RecordEvent(MessageE2)
						s.RecordEvent(MessageE3)
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
			"all recorded events match",
			ExecuteCommand(MessageC1),
			ToOnlyRecordEventsMatching(
				func(m dogma.Event) error {
					return nil
				},
			),
			expectPass,
			expectReport(
				`✓ only record events that match the predicate near expectation.messagematch.eventonly_test.go:75`,
			),
		),
		g.Entry(
			"all recorded events match, using predicate with application-defined type parameter",
			ExecuteCommand(MessageC1),
			ToOnlyRecordEventsMatching(
				func(m MessageE) error {
					return nil
				},
			),
			expectPass,
			expectReport(
				`✓ only record events that match the predicate near expectation.messagematch.eventonly_test.go:88`,
			),
		),
		g.Entry(
			"no events recorded at all",
			noop,
			ToOnlyRecordEventsMatching(
				func(m dogma.Event) error {
					panic("unexpected call")
				},
			),
			expectPass,
			expectReport(
				`✓ only record events that match the predicate near expectation.messagematch.eventonly_test.go:100`,
			),
		),
		g.Entry(
			"none of the recorded events match",
			ExecuteCommand(MessageC1),
			ToOnlyRecordEventsMatching(
				func(m dogma.Event) error {
					return errors.New("<error>")
				},
			),
			expectFail,
			expectReport(
				`✗ only record events that match the predicate near expectation.messagematch.eventonly_test.go:113`,
				``,
				`  | EXPLANATION`,
				`  |     none of the 3 relevant events matched the predicate`,
				`  | `,
				`  | FAILED MATCHES`,
				`  |     • fixtures.MessageE: <error> (repeated 3 times)`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • verify the logic within the predicate function`,
				`  |     • enable integration handlers using the EnableHandlerType() option`,
				`  |     • verify the logic within the '<aggregate>' aggregate message handler`,
			),
		),
		g.Entry(
			"some matching events recorded",
			ExecuteCommand(MessageC1),
			ToOnlyRecordEventsMatching(
				func(m dogma.Event) error {
					switch m {
					case fixtures.MessageE1:
						return errors.New("<error>")
					case fixtures.MessageE2:
						return IgnoreMessage
					default:
						return nil
					}
				},
			),
			expectFail,
			expectReport(
				`✗ only record events that match the predicate near expectation.messagematch.eventonly_test.go:137`,
				``,
				`  | EXPLANATION`,
				`  |     only 1 of 2 relevant events matched the predicate`,
				`  | `,
				`  | FAILED MATCHES`,
				`  |     • fixtures.MessageE: <error>`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • verify the logic within the predicate function, it ignored 1 event`,
				`  |     • enable integration handlers using the EnableHandlerType() option`,
				`  |     • verify the logic within the '<aggregate>' aggregate message handler`,
			),
		),
		g.Entry(
			"no matching events recorded, using predicate with application-defined type parameter",
			ExecuteCommand(MessageC1),
			ToOnlyRecordEventsMatching(
				func(m MessageX) error {
					panic("unexpected call")
				},
			),
			expectFail,
			expectReport(
				`✗ only record events that match the predicate near expectation.messagematch.eventonly_test.go:168`,
				``,
				`  | EXPLANATION`,
				`  |     none of the 3 relevant events matched the predicate`,
				`  | `,
				`  | FAILED MATCHES`,
				`  |     • fixtures.MessageE: predicate function expected fixtures.MessageX (repeated 3 times)`,
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
			var fn func(dogma.Event) error
			ToOnlyRecordEventsMatching(fn)
		}).To(PanicWith("ToOnlyRecordEventsMatching(<nil>): function must not be nil"))
	})
})
