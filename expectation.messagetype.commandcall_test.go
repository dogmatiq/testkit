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

var _ = g.Describe("func ToExecuteCommandType() (when used with the Call() action)", func() {
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
				c.Identity("<app>", "f38c3003-bbd0-4b4a-b1f8-6922e9545acd")

				c.RegisterAggregate(&AggregateMessageHandlerStub{
					ConfigureFunc: func(c dogma.AggregateConfigurer) {
						c.Identity("<aggregate>", "5c44e153-9e42-4816-87b0-59b4b4943dc4")
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
						c.Identity("<process>", "3994dd62-43f7-4569-813f-a616dc444486")
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
			"command type executed as expected",
			executeCommandViaExecutor(MessageR1),
			ToExecuteCommandType[MessageR](),
			expectPass,
			expectReport(
				`✓ execute any 'fixtures.MessageR' command`,
			),
		),
		g.Entry(
			"no messages produced at all",
			Call(func() {}),
			ToExecuteCommandType[MessageC](),
			expectFail,
			expectReport(
				`✗ execute any 'fixtures.MessageC' command`,
				``,
				`  | EXPLANATION`,
				`  |     no messages were produced at all`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • verify the logic within the code that uses the dogma.CommandExecutor`,
			),
		),
		g.Entry(
			"no matching command type executed and all relevant handler types disabled",
			executeCommandViaExecutor(MessageR1),
			ToExecuteCommandType[MessageC](),
			expectFail,
			expectReport(
				`✗ execute any 'fixtures.MessageC' command`,
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
			"command of a similar type executed",
			executeCommandViaExecutor(MessageR1),
			ToExecuteCommandType[*MessageR](), // note: message type is pointer
			expectFail,
			expectReport(
				`✗ execute any '*fixtures.MessageR' command`,
				``,
				`  | EXPLANATION`,
				`  |     a command of a similar type was executed via a dogma.CommandExecutor`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • check the message type, should it be a pointer?`,
				`  | `,
				`  | MESSAGE TYPE DIFF`,
				`  |     [-*-]fixtures.MessageR`,
			),
		),
	)
})
