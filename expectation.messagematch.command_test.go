package testkit_test

import (
	"context"
	"errors"
	"time"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/enginekit/enginetest/stubs"
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

	type (
		EventThatIsIgnored        = EventStub[TypeX]
		EventThatExecutesCommand  = EventStub[TypeC]
		EventThatSchedulesTimeout = EventStub[TypeT]

		CommandThatIsExecuted      = CommandStub[TypeC]
		CommandThatIsNeverExecuted = CommandStub[TypeX]

		TimeoutThatIsScheduled = TimeoutStub[TypeT]
	)

	g.BeforeEach(func() {
		testingT = &testingmock.T{
			FailSilently: true,
		}

		app = &ApplicationStub{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "95d4b9b2-a0ec-4dfb-aa57-c7e5ef5b1f02")

				// Register a process that will execute the commands about which
				// we will make assertions using ToExecuteCommand().
				c.RegisterProcess(&ProcessMessageHandlerStub{
					ConfigureFunc: func(c dogma.ProcessConfigurer) {
						c.Identity("<process>", "8b4c4701-be92-4b28-83b6-0d69b97fb451")
						c.Routes(
							dogma.HandlesEvent[EventThatIsIgnored](),

							dogma.HandlesEvent[EventThatExecutesCommand](),
							dogma.ExecutesCommand[CommandThatIsExecuted](),
							dogma.ExecutesCommand[CommandThatIsNeverExecuted](),

							dogma.HandlesEvent[EventThatSchedulesTimeout](),
							dogma.SchedulesTimeout[TimeoutThatIsScheduled](),
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
						switch m := m.(type) {
						case EventThatExecutesCommand:
							s.ExecuteCommand(
								CommandThatIsExecuted{
									Content: m.Content,
								},
							)

							if m.Content == "<multiple>" {
								s.ExecuteCommand(
									CommandThatIsExecuted{
										Content: m.Content,
									},
								)
							}

						case EventThatSchedulesTimeout:
							s.ScheduleTimeout(
								TimeoutThatIsScheduled{
									Content: m.Content,
								},
								time.Now().Add(1*time.Hour),
							)
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
			RecordEvent(EventThatExecutesCommand{}),
			ToExecuteCommandMatching(
				func(m dogma.Command) error {
					if m == (CommandThatIsExecuted{}) {
						return nil
					}

					return errors.New("<error>")
				},
			),
			expectPass,
			expectReport(
				`✓ execute a command that matches the predicate near expectation.messagematch.command_test.go:122`,
			),
		),
		g.Entry(
			"matching command executed as expected, using predicate with a more specific type",
			RecordEvent(EventThatExecutesCommand{}),
			ToExecuteCommandMatching(
				func(m CommandThatIsExecuted) error {
					if m == (CommandThatIsExecuted{}) {
						return nil
					}

					return errors.New("<error>")
				},
			),
			expectPass,
			expectReport(
				`✓ execute a command that matches the predicate near expectation.messagematch.command_test.go:139`,
			),
		),
		g.Entry(
			"no matching command executed",
			RecordEvent(EventThatExecutesCommand{}),
			ToExecuteCommandMatching(
				func(m dogma.Command) error {
					return errors.New("<error>")
				},
			),
			expectFail,
			expectReport(
				`✗ execute a command that matches the predicate near expectation.messagematch.command_test.go:156`,
				``,
				`  | EXPLANATION`,
				`  |     none of the engaged handlers executed a matching command`,
				`  | `,
				`  | FAILED MATCHES`,
				`  |     • stubs.CommandStub[TypeC]: <error>`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • verify the logic within the predicate function`,
				`  |     • verify the logic within the '<process>' process message handler`,
			),
		),
		g.Entry(
			"no matching command executed, using predicate with a more specific type",
			RecordEvent(EventThatExecutesCommand{}),
			ToExecuteCommandMatching(
				func(m CommandThatIsNeverExecuted) error {
					panic("unexpected call")
				},
			),
			expectFail,
			expectReport(
				`✗ execute a command that matches the predicate near expectation.messagematch.command_test.go:179`,
				``,
				`  | EXPLANATION`,
				`  |     none of the engaged handlers executed a matching command`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • verify the logic within the predicate function, it ignored 1 command`,
				`  |     • verify the logic within the '<process>' process message handler`),
		),
		g.Entry(
			"no messages produced at all",
			RecordEvent(EventThatIsIgnored{}),
			ToExecuteCommandMatching(
				func(m dogma.Command) error {
					panic("unexpected call")
				},
			),
			expectFail,
			expectReport(
				`✗ execute a command that matches the predicate near expectation.messagematch.command_test.go:198`,
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
			RecordEvent(EventThatSchedulesTimeout{}),
			ToExecuteCommandMatching(
				func(m dogma.Command) error {
					panic("unexpected call")
				},
			),
			expectFail,
			expectReport(
				`✗ execute a command that matches the predicate near expectation.messagematch.command_test.go:217`,
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
			RecordEvent(EventThatExecutesCommand{}),
			ToExecuteCommandMatching(
				func(m dogma.Command) error {
					return errors.New("<error>")
				},
			),
			expectFail,
			expectReport(
				`✗ execute a command that matches the predicate near expectation.messagematch.command_test.go:236`,
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
			RecordEvent(EventThatExecutesCommand{}),
			ToExecuteCommandMatching(
				func(m dogma.Command) error {
					return IgnoreMessage
				},
			),
			expectFail,
			expectReport(
				`✗ execute a command that matches the predicate near expectation.messagematch.command_test.go:258`,
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
			RecordEvent(EventThatExecutesCommand{
				Content: "<multiple>",
			}),
			ToExecuteCommandMatching(
				func(m dogma.Command) error {
					return errors.New("<error>")
				},
			),
			expectFail,
			expectReport(
				`✗ execute a command that matches the predicate near expectation.messagematch.command_test.go:280`,
				``,
				`  | EXPLANATION`,
				`  |     none of the engaged handlers executed a matching command`,
				`  | `,
				`  | FAILED MATCHES`,
				`  |     • stubs.CommandStub[TypeC]: <error> (repeated 2 times)`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • verify the logic within the predicate function`,
				`  |     • verify the logic within the '<process>' process message handler`,
			),
		),
		g.Entry(
			"does not include an explanation when negated and a sibling expectation passes",
			RecordEvent(EventThatExecutesCommand{}),
			NoneOf(
				ToExecuteCommandMatching(
					func(m dogma.Command) error {
						return nil
					},
				),
				ToExecuteCommandMatching(
					func(m dogma.Command) error {
						return errors.New("<error>")
					},
				),
			),
			expectFail,
			expectReport(
				`✗ none of (1 of the expectations passed unexpectedly)`,
				`    ✓ execute a command that matches the predicate near expectation.messagematch.command_test.go:305`,
				`    ✗ execute a command that matches the predicate near expectation.messagematch.command_test.go:309`,
			),
		),
	)

	g.It("panics if the function is nil", func() {
		Expect(func() {
			var fn func(dogma.Command) error
			ToExecuteCommandMatching(fn)
		}).To(PanicWith("ToExecuteCommandMatching(<nil>): function must not be nil"))
	})
})
