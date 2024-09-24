package testkit_test

import (
	"context"
	"errors"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/enginekit/enginetest/stubs"
	. "github.com/dogmatiq/testkit"
	"github.com/dogmatiq/testkit/internal/testingmock"
	g "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = g.Describe("func ToOnlyExecuteCommandsMatching()", func() {
	var (
		testingT *testingmock.T
		app      dogma.Application
	)

	type (
		EventThatExecutesCommands = EventStub[TypeC]
		CommandThatIsExecuted     = CommandStub[TypeC]
	)

	g.BeforeEach(func() {
		testingT = &testingmock.T{
			FailSilently: true,
		}

		app = &ApplicationStub{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "386480e5-4b83-4d3b-9b87-51e6d56e41e7")

				c.RegisterProcess(&ProcessMessageHandlerStub{
					ConfigureFunc: func(c dogma.ProcessConfigurer) {
						c.Identity("<process>", "39869c73-5ff0-4ae6-8317-eb494c87167b")
						c.Routes(
							dogma.HandlesEvent[EventThatExecutesCommands](),
							dogma.ExecutesCommand[CommandThatIsExecuted](),
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
						s.ExecuteCommand(CommandC1)
						s.ExecuteCommand(CommandC2)
						s.ExecuteCommand(CommandC3)
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
			"all executed commands match",
			RecordEvent(EventThatExecutesCommands{}),
			ToOnlyExecuteCommandsMatching(
				func(m dogma.Command) error {
					return nil
				},
			),
			expectPass,
			expectReport(
				`✓ only execute commands that match the predicate near expectation.messagematch.commandonly_test.go:84`,
			),
		),
		g.Entry(
			"all executed commands match, using predicate with a more specific type",
			RecordEvent(EventThatExecutesCommands{}),
			ToOnlyExecuteCommandsMatching(
				func(m CommandThatIsExecuted) error {
					return nil
				},
			),
			expectPass,
			expectReport(
				`✓ only execute commands that match the predicate near expectation.messagematch.commandonly_test.go:97`,
			),
		),
		g.Entry(
			"no commands executed at all",
			noop,
			ToOnlyExecuteCommandsMatching(
				func(m dogma.Command) error {
					panic("unexpected call")
				},
			),
			expectPass,
			expectReport(
				`✓ only execute commands that match the predicate near expectation.messagematch.commandonly_test.go:109`,
			),
		),
		g.Entry(
			"none of the executed commands match",
			RecordEvent(EventThatExecutesCommands{}),
			ToOnlyExecuteCommandsMatching(
				func(m dogma.Command) error {
					return errors.New("<error>")
				},
			),
			expectFail,
			expectReport(
				`✗ only execute commands that match the predicate near expectation.messagematch.commandonly_test.go:122`,
				``,
				`  | EXPLANATION`,
				`  |     none of the 3 relevant commands matched the predicate`,
				`  | `,
				`  | FAILED MATCHES`,
				`  |     • stubs.CommandStub[TypeC]: <error> (repeated 3 times)`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • verify the logic within the predicate function`,
				`  |     • verify the logic within the '<process>' process message handler`,
			),
		),
		g.Entry(
			"some matching commands executed",
			RecordEvent(EventThatExecutesCommands{}),
			ToOnlyExecuteCommandsMatching(
				func(m dogma.Command) error {
					switch m {
					case CommandC1:
						return errors.New("<error>")
					case CommandC2:
						return IgnoreMessage
					default:
						return nil
					}
				},
			),
			expectFail,
			expectReport(
				`✗ only execute commands that match the predicate near expectation.messagematch.commandonly_test.go:145`,
				``,
				`  | EXPLANATION`,
				`  |     only 1 of 2 relevant commands matched the predicate`,
				`  | `,
				`  | FAILED MATCHES`,
				`  |     • stubs.CommandStub[TypeC]: <error>`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • verify the logic within the predicate function, it ignored 1 command`,
				`  |     • verify the logic within the '<process>' process message handler`,
			),
		),
		g.Entry(
			"no executed commands match, using predicate with a more specific type",
			RecordEvent(EventThatExecutesCommands{}),
			ToOnlyExecuteCommandsMatching(
				func(m CommandStub[TypeX]) error {
					panic("unexpected call")
				},
			),
			expectFail,
			expectReport(
				`✗ only execute commands that match the predicate near expectation.messagematch.commandonly_test.go:175`,
				``,
				`  | EXPLANATION`,
				`  |     none of the 3 relevant commands matched the predicate`,
				`  | `,
				`  | FAILED MATCHES`,
				`  |     • stubs.CommandStub[TypeC]: predicate function expected stubs.CommandStub[TypeX] (repeated 3 times)`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • verify the logic within the predicate function`,
				`  |     • verify the logic within the '<process>' process message handler`,
			),
		),
	)

	g.It("panics if the function is nil", func() {
		Expect(func() {
			var fn func(dogma.Command) error
			ToOnlyExecuteCommandsMatching(fn)
		}).To(PanicWith("ToOnlyExecuteCommandsMatching(<nil>): function must not be nil"))
	})
})
