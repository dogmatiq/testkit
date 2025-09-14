package testkit_test

import (
	"context"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/enginekit/enginetest/stubs"
	. "github.com/dogmatiq/testkit"
	"github.com/dogmatiq/testkit/internal/testingmock"
	g "github.com/onsi/ginkgo/v2"
	gm "github.com/onsi/gomega"
)

var _ = g.Describe("func ToRecordEvent() (when used with the Call() action)", func() {
	var (
		testingT *testingmock.T
		app      dogma.Application
		test     *Test
	)

	type (
		CommandThatIsIgnored    = CommandStub[TypeX]
		CommandThatRecordsEvent = CommandStub[TypeE]

		EventThatIsRecorded      = EventStub[TypeE]
		EventThatIsNeverRecorded = EventStub[TypeX]
	)

	g.BeforeEach(func() {
		testingT = &testingmock.T{
			FailSilently: true,
		}

		app = &ApplicationStub{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "58067ffe-c1d6-4097-8acb-c55a7936cb4b")

				c.Routes(
					dogma.ViaAggregate(&AggregateMessageHandlerStub{
						ConfigureFunc: func(c dogma.AggregateConfigurer) {
							c.Identity("<aggregate>", "8746651e-df4d-421c-9eea-177585e5b8eb")
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
								s.RecordEvent(EventThatIsRecorded{
									Content: m.Content,
								})
							}
						},
					}),
				)
			},
		}
	})

	executeCommandViaExecutor := func(m dogma.Command) Action {
		return Call(func() {
			err := test.CommandExecutor().ExecuteCommand(context.Background(), m)
			gm.Expect(err).ShouldNot(gm.HaveOccurred())
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
			gm.Expect(testingT.Failed()).To(gm.Equal(!ok))
		},
		g.Entry(
			"event recorded as expected",
			executeCommandViaExecutor(CommandThatRecordsEvent{}),
			ToRecordEvent(EventThatIsRecorded{}),
			expectPass,
			expectReport(
				`✓ record a specific 'stubs.EventStub[TypeE]' event`,
			),
		),
		g.Entry(
			"no matching event recorded",
			executeCommandViaExecutor(CommandThatRecordsEvent{}),
			ToRecordEvent(EventThatIsNeverRecorded{}),
			expectFail,
			expectReport(
				`✗ record a specific 'stubs.EventStub[TypeX]' event`,
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
			ToRecordEvent(EventThatIsRecorded{}),
			expectFail,
			expectReport(
				`✗ record a specific 'stubs.EventStub[TypeE]' event`,
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
			executeCommandViaExecutor(CommandThatIsIgnored{}),
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
				`  |     • verify the logic within the code that uses the dogma.EventRecorder`,
			),
		),
	)
})
