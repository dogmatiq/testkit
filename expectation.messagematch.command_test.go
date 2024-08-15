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

var _ = g.Describe("func ToExecuteCommandMatching()", func() {
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
				c.Identity("<app>", "95d4b9b2-a0ec-4dfb-aa57-c7e5ef5b1f02")

				c.RegisterAggregate(&AggregateMessageHandler{
					ConfigureFunc: func(c dogma.AggregateConfigurer) {
						c.Identity("<aggregate>", "0c7df570-4c3b-4d88-b6aa-68f8716bd93b")
						c.Routes(
							dogma.HandlesCommand[MessageR](), // R = record an event
							dogma.RecordsEvent[MessageN](),
						)
					},
					RouteCommandToInstanceFunc: func(dogma.Message) string {
						return "<instance>"
					},
					HandleCommandFunc: func(
						_ dogma.AggregateRoot,
						s dogma.AggregateCommandScope,
						_ dogma.Message,
					) {
						s.RecordEvent(MessageN1)
					},
				})

				c.RegisterProcess(&ProcessMessageHandler{
					ConfigureFunc: func(c dogma.ProcessConfigurer) {
						c.Identity("<process>", "d28572f4-ecce-48eb-92c5-d3968ab8636c")
						c.Routes(
							dogma.HandlesEvent[MessageE](),    // E = event
							dogma.HandlesEvent[MessageN](),    // N = [do)]()thing
							dogma.ExecutesCommand[MessageC](), // C = command
							dogma.ExecutesCommand[MessageX](),
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
						if m, ok := m.(MessageE); ok {
							s.ExecuteCommand(MessageC1)

							if m.Value == "<multiple>" {
								s.ExecuteCommand(MessageC1)
							}
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
			"matching command executed as expected",
			RecordEvent(MessageE1),
			ToExecuteCommandMatching(
				func(m dogma.Message) error {
					if m == MessageC1 {
						return nil
					}

					return errors.New("<error>")
				},
			),
			expectPass,
			expectReport(
				`✓ execute a command that matches the predicate near expectation.messagematch.command_test.go:105`,
			),
		),
		g.Entry(
			"no matching command executed",
			RecordEvent(MessageE1),
			ToExecuteCommandMatching(
				func(m dogma.Message) error {
					if m == MessageX1 {
						return nil
					}

					return errors.New("<error>")
				},
			),
			expectFail,
			expectReport(
				`✗ execute a command that matches the predicate near expectation.messagematch.command_test.go:122`,
				``,
				`  | EXPLANATION`,
				`  |     none of the engaged handlers executed a matching command`,
				`  | `,
				`  | FAILED MATCHES`,
				`  |     • fixtures.MessageC: <error>`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • verify the logic within the predicate function`,
				`  |     • verify the logic within the '<process>' process message handler`,
			),
		),
		g.Entry(
			"no messages produced at all",
			RecordEvent(MessageN1),
			ToExecuteCommandMatching(
				func(m dogma.Message) error {
					panic("unexpected call")
				},
			),
			expectFail,
			expectReport(
				`✗ execute a command that matches the predicate near expectation.messagematch.command_test.go:149`,
				``,
				`  | EXPLANATION`,
				`  |     no messages were produced at all`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • verify the logic within the '<process>' process message handler`,
			),
		),
		g.Entry(
			"no commands produced at all",
			ExecuteCommand(MessageR1),
			ToExecuteCommandMatching(
				func(m dogma.Message) error {
					return IgnoreMessage
				},
			),
			expectFail,
			expectReport(
				`✗ execute a command that matches the predicate near expectation.messagematch.command_test.go:168`,
				``,
				`  | EXPLANATION`,
				`  |     no commands were executed at all`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • verify the logic within the '<process>' process message handler`,
			),
		),
		g.Entry(
			"no matching command executed and all relevant handler types disabled",
			ExecuteCommand(MessageR1),
			ToExecuteCommandMatching(
				func(m dogma.Message) error {
					if m == MessageX1 {
						return nil
					}

					return errors.New("<error>")
				},
			),
			expectFail,
			expectReport(
				`✗ execute a command that matches the predicate near expectation.messagematch.command_test.go:187`,
				``,
				`  | EXPLANATION`,
				`  |     no relevant handler types were enabled`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • enable process handlers using the EnableHandlerType() option`,
			),
			WithUnsafeOperationOptions(
				engine.EnableProcesses(false),
			),
		),
		g.Entry(
			"no matching command executed and commands were ignored",
			RecordEvent(MessageE1),
			ToExecuteCommandMatching(
				func(m dogma.Message) error {
					return IgnoreMessage
				},
			),
			expectFail,
			expectReport(
				`✗ execute a command that matches the predicate near expectation.messagematch.command_test.go:213`,
				``,
				`  | EXPLANATION`,
				`  |     none of the engaged handlers executed a matching command`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • verify the logic within the predicate function, it ignored 1 command`,
				`  |     • verify the logic within the '<process>' process message handler`,
			),
		),
		g.Entry(
			"no matching command executed and error messages were repeated",
			RecordEvent(MessageE{
				Value: "<multiple>", // trigger multiple MessageC commands
			}),
			ToExecuteCommandMatching(
				func(m dogma.Message) error {
					return errors.New("<error>")
				},
			),
			expectFail,
			expectReport(
				`✗ execute a command that matches the predicate near expectation.messagematch.command_test.go:235`,
				``,
				`  | EXPLANATION`,
				`  |     none of the engaged handlers executed a matching command`,
				`  | `,
				`  | FAILED MATCHES`,
				`  |     • fixtures.MessageC: <error> (repeated 2 times)`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • verify the logic within the predicate function`,
				`  |     • verify the logic within the '<process>' process message handler`,
			),
		),
		g.Entry(
			"does not include an explanation when negated and a sibling expectation passes",
			RecordEvent(MessageE1),
			NoneOf(
				ToExecuteCommandMatching(
					func(m dogma.Message) error {
						if m == MessageC1 {
							return nil
						}

						return errors.New("<error>")
					},
				),
				ToExecuteCommandMatching(
					func(m dogma.Message) error {
						if m == MessageX1 {
							return nil
						}

						return errors.New("<error>")
					},
				),
			),
			expectFail,
			expectReport(
				`✗ none of (1 of the expectations passed unexpectedly)`,
				`    ✓ execute a command that matches the predicate near expectation.messagematch.command_test.go:259`,
				`    ✗ execute a command that matches the predicate near expectation.messagematch.command_test.go:268`,
			),
		),
	)

	g.It("panics if the function is nil", func() {
		Expect(func() {
			ToExecuteCommandMatching(nil)
		}).To(PanicWith("ToExecuteCommandMatching(<nil>): function must not be nil"))
	})
})
