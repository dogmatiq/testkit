package testkit_test

import (
	"context"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogma/fixtures"
	. "github.com/dogmatiq/testkit"
	"github.com/dogmatiq/testkit/internal/testingmock"
	g "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = g.Describe("func ToRecordEventOfType() (when used with the Call() action)", func() {
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
				c.Identity("<app>", "9d18eaa1-3721-464d-8957-59a2349c3fbc")

				c.RegisterAggregate(&AggregateMessageHandler{
					ConfigureFunc: func(c dogma.AggregateConfigurer) {
						c.Identity("<aggregate>", "23c7a071-0a52-4aae-8dbf-d2bc06e32079")
						c.Routes(
							dogma.HandlesCommand[MessageR](), // R = record an event
							dogma.HandlesCommand[MessageN](), // N = do nothing
							dogma.RecordsEvent[MessageE](),
							dogma.RecordsEvent[*MessageE](), // pointer, used to test type similarity
							dogma.RecordsEvent[MessageX](),
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
						if _, ok := m.(MessageR); ok {
							s.RecordEvent(MessageE1)
						}
					},
				})

				c.RegisterProcess(&ProcessMessageHandler{
					ConfigureFunc: func(c dogma.ProcessConfigurer) {
						c.Identity("<process>", "d2fdcbb0-b600-487c-97b4-f885b997cad3")
						c.Routes(
							dogma.HandlesEvent[MessageE](), // E = execute a command
							dogma.HandlesEvent[MessageA](), // A = also execute a command
							dogma.ExecutesCommand[MessageN](),
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
						if _, ok := m.(MessageE); ok {
							s.ExecuteCommand(MessageN1)
						}
						return nil
					},
				})
			},
		}
	})

	executeCommandViaExecutor := func(m dogma.Command) Action {
		return Call(func() {
			err := test.CommandExecutor().ExecuteCommand(context.Background(), m)
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
			"no matching event types recorded",
			executeCommandViaExecutor(MessageR1),
			ToRecordEventOfType(MessageX{}),
			expectFail,
			expectReport(
				`✗ record any 'fixtures.MessageX' event`,
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
			ToRecordEventOfType(MessageE{}),
			expectFail,
			expectReport(
				`✗ record any 'fixtures.MessageE' event`,
				``,
				`  | EXPLANATION`,
				`  |     no messages were produced at all`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • verify the logic within the code that uses the dogma.EventRecorder`,
			),
		),
		g.Entry(
			"no events produced at all",
			executeCommandViaExecutor(MessageN1),
			ToRecordEventOfType(MessageE{}),
			expectFail,
			expectReport(
				`✗ record any 'fixtures.MessageE' event`,
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
