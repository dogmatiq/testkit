package testkit_test

import (
	"context"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogma/fixtures"
	. "github.com/dogmatiq/enginekit/enginetest/stubs"
	. "github.com/dogmatiq/testkit"
	"github.com/dogmatiq/testkit/engine"
	"github.com/dogmatiq/testkit/internal/testingmock"
	g "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = g.Describe("func ToExecuteCommand() (when used with the Call() action)", func() {
	var (
		testingT *testingmock.T
		app      dogma.Application
		test     *Test
	)

	g.BeforeEach(func() {
		testingT = &testingmock.T{
			FailSilently: true,
		}

		app = &ApplicationStub{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "556be8fb-fa7b-4240-882e-86e735b7705d")

				c.RegisterAggregate(&AggregateMessageHandlerStub{
					ConfigureFunc: func(c dogma.AggregateConfigurer) {
						c.Identity("<aggregate>", "c94fc20f-f6fc-46c6-ad6f-45bb854aeb6c")
						c.Routes(
							dogma.HandlesCommand[MessageR](),  // R = record an event
							dogma.HandlesCommand[*MessageR](), // pointer, used to test type similarity
							dogma.HandlesCommand[MessageX](),
							dogma.RecordsEvent[MessageN](),
						)
					},
					RouteCommandToInstanceFunc: func(dogma.Command) string {
						return "<instance>"
					},
					HandleCommandFunc: func(
						_ dogma.AggregateRoot,
						s dogma.AggregateCommandScope,
						_ dogma.Command,
					) {
						s.RecordEvent(MessageN1)
					},
				})

				c.RegisterProcess(&ProcessMessageHandlerStub{
					ConfigureFunc: func(c dogma.ProcessConfigurer) {
						c.Identity("<process>", "e914b7b8-745b-4635-88a7-2d06628098a4")
						c.Routes(
							dogma.HandlesEvent[MessageE](),    // E = event
							dogma.HandlesEvent[MessageN](),    // N = (do) nothing
							dogma.ExecutesCommand[MessageC](), // C = command
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
							s.ExecuteCommand(MessageC1)
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
			"command executed as expected",
			executeCommandViaExecutor(MessageR1),
			ToExecuteCommand(MessageR1),
			expectPass,
			expectReport(
				`✓ execute a specific 'fixtures.MessageR' command`,
			),
		),
		g.Entry(
			"no messages produced at all",
			Call(func() {}),
			ToExecuteCommand(MessageC{}),
			expectFail,
			expectReport(
				`✗ execute a specific 'fixtures.MessageC' command`,
				``,
				`  | EXPLANATION`,
				`  |     no messages were produced at all`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • verify the logic within the code that uses the dogma.CommandExecutor`,
			),
		),
		g.Entry(
			"no matching command executed and all relevant handler types disabled",
			executeCommandViaExecutor(MessageR1),
			ToExecuteCommand(MessageC1),
			expectFail,
			expectReport(
				`✗ execute a specific 'fixtures.MessageC' command`,
				``,
				`  | EXPLANATION`,
				`  |     nothing executed a matching command`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • enable process handlers using the EnableHandlerType() option`,
				`  |     • verify the logic within the code that uses the dogma.CommandExecutor`,
			),
			WithUnsafeOperationOptions(
				engine.EnableProcesses(false),
			),
		),
		g.Entry(
			"similar command executed with a different type",
			executeCommandViaExecutor(MessageR1),
			ToExecuteCommand(&MessageR1), // note: message type is pointer
			expectFail,
			expectReport(
				`✗ execute a specific '*fixtures.MessageR' command`,
				``,
				`  | EXPLANATION`,
				`  |     a command of a similar type was executed via a dogma.CommandExecutor`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • check the message type, should it be a pointer?`,
				`  | `,
				`  | MESSAGE DIFF`,
				`  |     [-*-]fixtures.MessageR{`,
				`  |         Value: "R1"`,
				`  |     }`,
			),
		),
		g.Entry(
			"similar command executed with a different value",
			executeCommandViaExecutor(MessageR1),
			ToExecuteCommand(MessageR2), // note: message type is pointer
			expectFail,
			expectReport(
				`✗ execute a specific 'fixtures.MessageR' command`,
				``,
				`  | EXPLANATION`,
				`  |     a similar command was executed via a dogma.CommandExecutor`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • check the content of the message`,
				`  | `,
				`  | MESSAGE DIFF`,
				`  |     fixtures.MessageR{`,
				`  |         Value: "R[-2-]{+1+}"`,
				`  |     }`,
			),
		),
	)
})
