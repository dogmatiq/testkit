package testkit_test

import (
	"context"
	"errors"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/enginekit/enginetest/stubs"
	. "github.com/dogmatiq/testkit"
	"github.com/dogmatiq/testkit/engine"
	"github.com/dogmatiq/testkit/internal/testingmock"
	g "github.com/onsi/ginkgo/v2"
	gm "github.com/onsi/gomega"
)

var _ = g.Describe("func ToRecordEventMatching()", func() {
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
	)

	g.BeforeEach(func() {
		testingT = &testingmock.T{
			FailSilently: true,
		}

		app = &ApplicationStub{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "43962b88-b25c-4a59-938e-64540c473a7c")

				c.Routes(
					dogma.ViaAggregate(&AggregateMessageHandlerStub{
						ConfigureFunc: func(c dogma.AggregateConfigurer) {
							c.Identity("<aggregate>", "6a182af9-4659-49cf-8c44-85dfd47dbefc")
							c.Routes(
								dogma.HandlesCommand[CommandThatIsIgnored](),

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
							switch m := m.(type) {
							case CommandThatRecordsEvent:
								s.RecordEvent(
									EventThatIsRecorded{
										Content: m.Content,
									},
								)

								if m.Content == "<multiple>" {
									s.RecordEvent(
										EventThatIsRecorded{
											Content: m.Content,
										},
									)
								}
							}
						},
					}),

					dogma.ViaProcess(&ProcessMessageHandlerStub{
						ConfigureFunc: func(c dogma.ProcessConfigurer) {
							c.Identity("<process>", "7651ad19-1526-48c0-a53d-676286a34ca6")
							c.Routes(
								dogma.HandlesEvent[EventThatExecutesCommand](),
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
			"matching event recorded as expected",
			ExecuteCommand(CommandThatRecordsEvent{}),
			ToRecordEventMatching(
				func(m dogma.Event) error {
					if m == (EventThatIsRecorded{}) {
						return nil
					}

					return errors.New("<error>")
				},
			),
			expectPass,
			expectReport(
				`✓ record an event that matches the predicate near expectation.messagematch.event_test.go:127`,
			),
		),
		g.Entry(
			"matching event recorded as expected, using predicate with a more specific type",
			ExecuteCommand(CommandThatRecordsEvent{}),
			ToRecordEventMatching(
				func(m EventThatIsRecorded) error {
					if m == (EventThatIsRecorded{}) {
						return nil
					}

					return errors.New("<error>")
				},
			),
			expectPass,
			expectReport(
				`✓ record an event that matches the predicate near expectation.messagematch.event_test.go:144`,
			),
		),
		g.Entry(
			"no matching event recorded",
			ExecuteCommand(CommandThatRecordsEvent{}),
			ToRecordEventMatching(
				func(m dogma.Event) error {
					return errors.New("<error>")
				},
			),
			expectFail,
			expectReport(
				`✗ record an event that matches the predicate near expectation.messagematch.event_test.go:161`,
				``,
				`  | EXPLANATION`,
				`  |     none of the engaged handlers recorded a matching event`,
				`  | `,
				`  | FAILED MATCHES`,
				`  |     • stubs.EventStub[TypeE]: <error>`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • verify the logic within the predicate function`,
				`  |     • enable integration handlers using the EnableHandlerType() option`,
				`  |     • verify the logic within the '<aggregate>' aggregate message handler`,
			),
		),
		g.Entry(
			"no matching event recorded, using predicate with a more specific type",
			ExecuteCommand(CommandThatRecordsEvent{}),
			ToRecordEventMatching(
				func(m EventThatIsNeverRecorded) error {
					panic("unexpected call")
				},
			),
			expectFail,
			expectReport(
				`✗ record an event that matches the predicate near expectation.messagematch.event_test.go:185`,
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
			"no messages produced at all",
			ExecuteCommand(CommandThatIsIgnored{}),
			ToRecordEventMatching(
				func(m dogma.Event) error {
					panic("unexpected call")
				},
			),
			expectFail,
			expectReport(
				`✗ record an event that matches the predicate near expectation.messagematch.event_test.go:206`,
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
			ToRecordEventMatching(
				func(m dogma.Event) error {
					panic("unexpected call")
				},
			),
			expectFail,
			expectReport(
				`✗ record an event that matches the predicate near expectation.messagematch.event_test.go:226`,
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
			ExecuteCommand(CommandThatRecordsEvent{}),
			ToRecordEventMatching(
				func(m dogma.Event) error {
					panic("unexpected call")
				},
			),
			expectFail,
			expectReport(
				`✗ record an event that matches the predicate near expectation.messagematch.event_test.go:246`,
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
			ExecuteCommand(CommandThatRecordsEvent{}),
			ToRecordEventMatching(
				func(m dogma.Event) error {
					return IgnoreMessage
				},
			),
			expectFail,
			expectReport(
				`✗ record an event that matches the predicate near expectation.messagematch.event_test.go:270`,
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
			ExecuteCommand(CommandThatRecordsEvent{
				Content: "<multiple>",
			}),
			ToRecordEventMatching(
				func(m dogma.Event) error {
					return errors.New("<error>")
				},
			),
			expectFail,
			expectReport(
				`✗ record an event that matches the predicate near expectation.messagematch.event_test.go:293`,
				``,
				`  | EXPLANATION`,
				`  |     none of the engaged handlers recorded a matching event`,
				`  | `,
				`  | FAILED MATCHES`,
				`  |     • stubs.EventStub[TypeE]: <error> (repeated 2 times)`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • verify the logic within the predicate function`,
				`  |     • enable integration handlers using the EnableHandlerType() option`,
				`  |     • verify the logic within the '<aggregate>' aggregate message handler`,
			),
		),
		g.Entry(
			"does not include an explanation when negated and a sibling expectation passes",
			ExecuteCommand(CommandThatRecordsEvent{}),
			NoneOf(
				ToRecordEventMatching(
					func(m dogma.Event) error {
						return nil
					},
				),
				ToRecordEventMatching(
					func(m dogma.Event) error {
						return errors.New("<error>")
					},
				),
			),
			expectFail,
			expectReport(
				`✗ none of (1 of the expectations passed unexpectedly)`,
				`    ✓ record an event that matches the predicate near expectation.messagematch.event_test.go:319`,
				`    ✗ record an event that matches the predicate near expectation.messagematch.event_test.go:323`,
			),
		),
	)

	g.It("fails the test if the message type is unrecognized", func() {
		test := Begin(testingT, app)
		test.Expect(
			noop,
			ToRecordEventMatching(
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
			ToRecordEventMatching(
				func(EventThatExecutesCommand) error {
					return nil
				},
			),
		)

		gm.Expect(testingT.Failed()).To(gm.BeTrue())
		gm.Expect(testingT.Logs).To(gm.ContainElement(
			"no handlers record events of type stubs.EventStub[TypeC], it is only ever consumed",
		))
	})

	g.It("panics if the function is nil", func() {
		gm.Expect(func() {
			var fn func(dogma.Event) error
			ToRecordEventMatching(fn)
		}).To(gm.PanicWith("ToRecordEventMatching(<nil>): function must not be nil"))
	})
})
