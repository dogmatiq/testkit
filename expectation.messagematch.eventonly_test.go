package testkit_test

import (
	"errors"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/enginekit/enginetest/stubs"
	. "github.com/dogmatiq/testkit"
	"github.com/dogmatiq/testkit/internal/testingmock"
	g "github.com/onsi/ginkgo/v2"
	gm "github.com/onsi/gomega"
)

var _ = g.Describe("func ToOnlyRecordEventsMatching()", func() {
	var (
		testingT *testingmock.T
		app      dogma.Application
	)

	type (
		CommandThatRecordsEvent = CommandStub[TypeE]

		EventThatIsRecorded      = EventStub[TypeE]
		EventThatIsNeverRecorded = EventStub[TypeX]
		EventThatIsOnlyConsumed  = EventStub[TypeO]
	)

	g.BeforeEach(func() {
		testingT = &testingmock.T{
			FailSilently: true,
		}

		app = &ApplicationStub{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "94f425c5-339a-4213-8309-16234225480e")

				c.Routes(
					dogma.ViaAggregate(&AggregateMessageHandlerStub{
						ConfigureFunc: func(c dogma.AggregateConfigurer) {
							c.Identity("<aggregate>", "bc64cfe4-3339-4eee-a9d2-364d33dff47d")
							c.Routes(
								dogma.HandlesCommand[CommandThatRecordsEvent](),
								dogma.RecordsEvent[EventThatIsRecorded](),
								dogma.RecordsEvent[EventThatIsNeverRecorded](),
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
							s.RecordEvent(EventE1)
							s.RecordEvent(EventE2)
							s.RecordEvent(EventE3)
						},
					}),

					dogma.ViaProjection(&ProjectionMessageHandlerStub{
						ConfigureFunc: func(c dogma.ProjectionConfigurer) {
							c.Identity("<projection>", "de708f1d-3651-437e-91ae-275a423ecd15")
							c.Routes(
								dogma.HandlesEvent[EventThatIsOnlyConsumed](),
							)
						},
					}),
				)
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
			gm.Expect(testingT.Failed()).To(gm.Equal(!ok))
		},
		g.Entry(
			"all recorded events match",
			ExecuteCommand(CommandThatRecordsEvent{}),
			ToOnlyRecordEventsMatching(
				func(m dogma.Event) error {
					return nil
				},
			),
			expectPass,
			expectReport(
				`✓ only record events that match the predicate near expectation.messagematch.eventonly_test.go:91`,
			),
		),
		g.Entry(
			"all recorded events match, using predicate with a more specific type",
			ExecuteCommand(CommandThatRecordsEvent{}),
			ToOnlyRecordEventsMatching(
				func(m EventThatIsRecorded) error {
					return nil
				},
			),
			expectPass,
			expectReport(
				`✓ only record events that match the predicate near expectation.messagematch.eventonly_test.go:104`,
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
				`✓ only record events that match the predicate near expectation.messagematch.eventonly_test.go:116`,
			),
		),
		g.Entry(
			"none of the recorded events match",
			ExecuteCommand(CommandThatRecordsEvent{}),
			ToOnlyRecordEventsMatching(
				func(m dogma.Event) error {
					return errors.New("<error>")
				},
			),
			expectFail,
			expectReport(
				`✗ only record events that match the predicate near expectation.messagematch.eventonly_test.go:129`,
				``,
				`  | EXPLANATION`,
				`  |     none of the 3 relevant events matched the predicate`,
				`  | `,
				`  | FAILED MATCHES`,
				`  |     • stubs.EventStub[TypeE]: <error> (repeated 3 times)`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • verify the logic within the predicate function`,
				`  |     • enable integration handlers using the EnableHandlerType() option`,
				`  |     • verify the logic within the '<aggregate>' aggregate message handler`,
			),
		),
		g.Entry(
			"some matching events recorded",
			ExecuteCommand(CommandThatRecordsEvent{}),
			ToOnlyRecordEventsMatching(
				func(m dogma.Event) error {
					switch m {
					case EventE1:
						return errors.New("<error>")
					case EventE2:
						return IgnoreMessage
					default:
						return nil
					}
				},
			),
			expectFail,
			expectReport(
				`✗ only record events that match the predicate near expectation.messagematch.eventonly_test.go:153`,
				``,
				`  | EXPLANATION`,
				`  |     only 1 of 2 relevant events matched the predicate`,
				`  | `,
				`  | FAILED MATCHES`,
				`  |     • stubs.EventStub[TypeE]: <error>`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • verify the logic within the predicate function, it ignored 1 event`,
				`  |     • enable integration handlers using the EnableHandlerType() option`,
				`  |     • verify the logic within the '<aggregate>' aggregate message handler`,
			),
		),
		g.Entry(
			"no matching events recorded, using predicate with a more specific type",
			ExecuteCommand(CommandThatRecordsEvent{}),
			ToOnlyRecordEventsMatching(
				func(m EventThatIsNeverRecorded) error {
					panic("unexpected call")
				},
			),
			expectFail,
			expectReport(
				`✗ only record events that match the predicate near expectation.messagematch.eventonly_test.go:184`,
				``,
				`  | EXPLANATION`,
				`  |     none of the 3 relevant events matched the predicate`,
				`  | `,
				`  | FAILED MATCHES`,
				`  |     • stubs.EventStub[TypeE]: predicate function expected stubs.EventStub[TypeX] (repeated 3 times)`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • verify the logic within the predicate function`,
				`  |     • enable integration handlers using the EnableHandlerType() option`,
				`  |     • verify the logic within the '<aggregate>' aggregate message handler`,
			),
		),
	)

	g.It("fails the test if the message type is unrecognized", func() {
		test := Begin(testingT, app)
		test.Expect(
			noop,
			ToOnlyRecordEventsMatching(
				func(EventStub[TypeU]) error {
					return nil
				},
			),
		)

		gm.Expect(testingT.Failed()).To(gm.BeTrue())
		gm.Expect(testingT.Logs).To(gm.ContainElement(
			"an event of type stubs.EventStub[TypeU] can never be recorded, the application does not use this message type",
		))
	})

	g.It("fails the test if the message type is not produced by any handlers", func() {
		test := Begin(testingT, app)
		test.Expect(
			noop,
			ToOnlyRecordEventsMatching(
				func(EventThatIsOnlyConsumed) error {
					return nil
				},
			),
		)

		gm.Expect(testingT.Failed()).To(gm.BeTrue())
		gm.Expect(testingT.Logs).To(gm.ContainElement(
			"no handlers record events of type stubs.EventStub[TypeO], it is only ever consumed",
		))
	})

	g.It("panics if the function is nil", func() {
		gm.Expect(func() {
			var fn func(dogma.Event) error
			ToOnlyRecordEventsMatching(fn)
		}).To(gm.PanicWith("ToOnlyRecordEventsMatching(<nil>): function must not be nil"))
	})
})
