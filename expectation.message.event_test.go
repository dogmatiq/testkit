package testkit_test

import (
	"context"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/enginekit/enginetest/stubs"
	. "github.com/dogmatiq/testkit"
	"github.com/dogmatiq/testkit/engine"
	"github.com/dogmatiq/testkit/internal/testingmock"
	g "github.com/onsi/ginkgo/v2"
	gm "github.com/onsi/gomega"
)

var _ = g.Describe("func ToRecordEvent()", func() {
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
				c.Identity("<app>", "adb2ed1e-b1f4-4756-abfa-a5e3a3e08def")

				c.RegisterAggregate(&AggregateMessageHandlerStub{
					ConfigureFunc: func(c dogma.AggregateConfigurer) {
						c.Identity("<aggregate>", "8746651e-df4d-421c-9eea-177585e5b8eb")
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
						c.Identity("<process>", "209c7f0f-49ad-4419-88a6-4e9ee1cf204a")
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
			gm.Expect(testingT.Failed()).To(gm.Equal(!ok))
		},
		g.Entry(
			"event recorded as expected",
			ExecuteCommand(CommandThatRecordsEvent{}),
			ToRecordEvent(EventThatIsRecorded{}),
			expectPass,
			expectReport(
				`✓ record a specific 'stubs.EventStub[TypeE]' event`,
			),
		),
		g.Entry(
			"no matching event recorded",
			ExecuteCommand(CommandThatRecordsEvent{}),
			ToRecordEvent(EventThatIsNeverRecorded{}),
			expectFail,
			expectReport(
				`✗ record a specific 'stubs.EventStub[TypeX]' event`,
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
			ExecuteCommand(CommandThatRecordsEvent{}),
			ToRecordEvent(EventThatIsRecorded{}),
			expectFail,
			expectReport(
				`✗ record a specific 'stubs.EventStub[TypeE]' event`,
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
			ExecuteCommand(CommandThatRecordsEvent{}),
			ToRecordEvent(EventThatIsRecorded{}),
			expectFail,
			expectReport(
				`✗ record a specific 'stubs.EventStub[TypeE]' event`,
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
			ToRecordEvent(EventThatIsRecorded{}),
			expectFail,
			expectReport(
				`✗ record a specific 'stubs.EventStub[TypeE]' event`,
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
			ToRecordEvent(EventThatIsRecorded{}),
			expectFail,
			expectReport(
				`✗ record a specific 'stubs.EventStub[TypeE]' event`,
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
			ExecuteCommand(CommandThatRecordsEvent{}),
			ToRecordEvent(&EventThatIsRecorded{}), // note: message type is pointer
			expectFail,
			expectReport(
				`✗ record a specific '*stubs.EventStub[TypeE]' event`,
				``,
				`  | EXPLANATION`,
				`  |     an event of a similar type was recorded by the '<aggregate>' aggregate message handler`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • check the message type, should it be a pointer?`,
				`  | `,
				`  | MESSAGE DIFF`,
				`  |     [-*-]stubs.EventStub[github.com/dogmatiq/enginekit/enginetest/stubs.TypeE]{<zero>}`,
			),
		),
		g.Entry(
			"similar event recorded with a different value",
			ExecuteCommand(CommandThatRecordsEvent{Content: "<content>"}),
			ToRecordEvent(EventThatIsRecorded{Content: "<different>"}),
			expectFail,
			expectReport(
				`✗ record a specific 'stubs.EventStub[TypeE]' event`,
				``,
				`  | EXPLANATION`,
				`  |     a similar event was recorded by the '<aggregate>' aggregate message handler`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • check the content of the message`,
				`  | `,
				`  | MESSAGE DIFF`,
				`  |     stubs.EventStub[github.com/dogmatiq/enginekit/enginetest/stubs.TypeE]{`,
				`  |         Content:         "<[-differ-]{+cont+}ent>"`,
				`  |         ValidationError: ""`,
				`  |     }`,
			),
		),
		g.Entry(
			"does not include an explanation when negated and a sibling expectation passes",
			ExecuteCommand(CommandThatRecordsEvent{}),
			NoneOf(
				ToRecordEvent(EventThatIsRecorded{}),
				ToRecordEvent(EventThatIsNeverRecorded{}),
			),
			expectFail,
			expectReport(
				`✗ none of (1 of the expectations passed unexpectedly)`,
				`    ✓ record a specific 'stubs.EventStub[TypeE]' event`,
				`    ✗ record a specific 'stubs.EventStub[TypeX]' event`,
			),
		),
	)

	g.It("fails the test if the message type is unrecognized", func() {
		test := Begin(testingT, app)
		test.Expect(
			noop,
			ToRecordEvent(EventU1),
		)

		gm.Expect(testingT.Failed()).To(gm.BeTrue())
		gm.Expect(testingT.Logs).To(gm.ContainElement(
			"an event of type stubs.EventStub[TypeU] can never be recorded, the application does not use this message type",
		))
	})

	g.It("fails the test if the message type is not an event", func() {
		test := Begin(testingT, app)
		test.Expect(
			noop,
			ToRecordEvent(CommandThatIsIgnored{}),
		)

		gm.Expect(testingT.Failed()).To(gm.BeTrue())
		gm.Expect(testingT.Logs).To(gm.ContainElement(
			"stubs.CommandStub[TypeX] is a command, it can never be recorded as an event",
		))
	})

	g.It("fails the test if the message type is not produced by any handlers", func() {
		test := Begin(testingT, app)
		test.Expect(
			noop,
			ToRecordEvent(EventThatIsOnlyConsumed{}),
		)

		gm.Expect(testingT.Failed()).To(gm.BeTrue())
		gm.Expect(testingT.Logs).To(gm.ContainElement(
			"no handlers record events of type stubs.EventStub[TypeO], it is only ever consumed",
		))
	})

	g.It("panics if the message is nil", func() {
		gm.Expect(func() {
			ToRecordEvent(nil)
		}).To(gm.PanicWith("ToRecordEvent(<nil>): message must not be nil"))
	})
})
