package testkit_test

import (
	"context"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/enginekit/enginetest/stubs"
	. "github.com/dogmatiq/testkit"
	"github.com/dogmatiq/testkit/engine"
	"github.com/dogmatiq/testkit/internal/testingmock"
	g "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = g.Describe("func ToRecordEventType()", func() {
	var (
		testingT *testingmock.T
		app      dogma.Application
	)

	type (
		CommandThatIsIgnored    = CommandStub[TypeX]
		CommandThatRecordsEvent = CommandStub[TypeE]

		EventThatIsRecorded      = EventStub[TypeE]
		EventThatIsNeverRecorded = EventStub[TypeX]
		EventThatExecutesCommand = EventStub[TypeC]
		EventThatIsOnlyConsumed  = EventStub[TypeO]
	)

	g.BeforeEach(func() {
		testingT = &testingmock.T{
			FailSilently: true,
		}

		app = &ApplicationStub{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "ef25ca55-2ace-40b5-9c2d-c53f5a80908a")

				c.RegisterAggregate(&AggregateMessageHandlerStub{
					ConfigureFunc: func(c dogma.AggregateConfigurer) {
						c.Identity("<aggregate>", "2cc50fd0-3d22-4f96-81c6-5e28d6abe735")
						c.Routes(
							dogma.HandlesCommand[CommandThatIsIgnored](),

							dogma.HandlesCommand[CommandThatRecordsEvent](),
							dogma.RecordsEvent[EventThatIsRecorded](),
							dogma.RecordsEvent[*EventThatIsRecorded](), // pointer, used to test type similarity
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
						switch m := m.(type) {
						case CommandThatRecordsEvent:
							s.RecordEvent(EventThatIsRecorded{
								Content: m.Content,
							})
						}
					},
				})

				c.RegisterProcess(&ProcessMessageHandlerStub{
					ConfigureFunc: func(c dogma.ProcessConfigurer) {
						c.Identity("<process>", "94de2bb7-c115-494d-ad15-bdfedbe4aec3")
						c.Routes(
							dogma.HandlesEvent[EventThatExecutesCommand](),
							dogma.HandlesEvent[EventThatIsOnlyConsumed](),
							dogma.ExecutesCommand[CommandThatIsIgnored](),
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
						switch m.(type) {
						case EventThatExecutesCommand:
							s.ExecuteCommand(CommandThatIsIgnored{})
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
			ExecuteCommand(CommandThatRecordsEvent{}),
			ToRecordEventType[EventThatIsRecorded](),
			expectPass,
			expectReport(
				`✓ record any 'stubs.EventStub[TypeE]' event`,
			),
		),
		g.Entry(
			"no matching event type recorded",
			ExecuteCommand(CommandThatRecordsEvent{}),
			ToRecordEventType[EventThatIsNeverRecorded](),
			expectFail,
			expectReport(
				`✗ record any 'stubs.EventStub[TypeX]' event`,
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
			ExecuteCommand(CommandThatRecordsEvent{}),
			ToRecordEventType[EventThatIsNeverRecorded](),
			expectFail,
			expectReport(
				`✗ record any 'stubs.EventStub[TypeX]' event`,
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
			ExecuteCommand(CommandThatRecordsEvent{}),
			ToRecordEventType[EventThatIsNeverRecorded](),
			expectFail,
			expectReport(
				`✗ record any 'stubs.EventStub[TypeX]' event`,
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
			ExecuteCommand(CommandThatIsIgnored{}),
			ToRecordEventType[EventThatIsNeverRecorded](),
			expectFail,
			expectReport(
				`✗ record any 'stubs.EventStub[TypeX]' event`,
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
			RecordEvent(EventThatExecutesCommand{}),
			ToRecordEventType[EventThatIsRecorded](),
			expectFail,
			expectReport(
				`✗ record any 'stubs.EventStub[TypeE]' event`,
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
			ExecuteCommand(CommandThatRecordsEvent{}),
			ToRecordEventType[*EventThatIsRecorded](), // note: message type is pointer
			expectFail,
			expectReport(
				`✗ record any '*stubs.EventStub[TypeE]' event`,
				``,
				`  | EXPLANATION`,
				`  |     an event of a similar type was recorded by the '<aggregate>' aggregate message handler`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • check the message type, should it be a pointer?`,
				`  | `,
				`  | MESSAGE TYPE DIFF`,
				`  |     [-*-]stubs.EventStub[TypeE]`,
			),
		),
		g.Entry(
			"does not include an explanation when negated and a sibling expectation passes",
			ExecuteCommand(CommandThatRecordsEvent{}),
			NoneOf(
				ToRecordEventType[EventThatIsRecorded](),
				ToRecordEventType[EventThatIsNeverRecorded](),
			),
			expectFail,
			expectReport(
				`✗ none of (1 of the expectations passed unexpectedly)`,
				`    ✓ record any 'stubs.EventStub[TypeE]' event`,
				`    ✗ record any 'stubs.EventStub[TypeX]' event`,
			),
		),
	)

	g.It("fails the test if the message type is unrecognized", func() {
		test := Begin(testingT, app)
		test.Expect(
			noop,
			ToRecordEventType[EventStub[TypeU]](),
		)

		Expect(testingT.Failed()).To(BeTrue())
		Expect(testingT.Logs).To(ContainElement(
			"an event of type stubs.EventStub[TypeU] can never be recorded, the application does not use this message type",
		))
	})

	g.It("fails the test if the message type is not an event", func() {
		test := Begin(testingT, app)
		test.Expect(
			noop,
			ToRecordEventType[CommandThatRecordsEvent](),
		)

		Expect(testingT.Failed()).To(BeTrue())
		Expect(testingT.Logs).To(ContainElement(
			"stubs.CommandStub[TypeE] is a command, it can never be recorded as an event",
		))
	})

	g.It("fails the test if the message type is not produced by any handlers", func() {
		test := Begin(testingT, app)
		test.Expect(
			noop,
			ToRecordEventType[EventThatIsOnlyConsumed](),
		)

		Expect(testingT.Failed()).To(BeTrue())
		Expect(testingT.Logs).To(ContainElement(
			"no handlers record events of type stubs.EventStub[TypeO], it is only ever consumed",
		))
	})
})
