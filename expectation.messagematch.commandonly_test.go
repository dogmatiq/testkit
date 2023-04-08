package testkit_test

import (
	"context"
	"errors"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/dogma/fixtures"
	. "github.com/dogmatiq/dogma/fixtures"
	. "github.com/dogmatiq/testkit"
	"github.com/dogmatiq/testkit/internal/testingmock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("func ToOnlyExecuteCommandsMatching()", func() {
	var (
		testingT *testingmock.T
		app      dogma.Application
	)

	BeforeEach(func() {
		testingT = &testingmock.T{
			FailSilently: true,
		}

		app = &Application{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "386480e5-4b83-4d3b-9b87-51e6d56e41e7")

				c.RegisterProcess(&ProcessMessageHandler{
					ConfigureFunc: func(c dogma.ProcessConfigurer) {
						c.Identity("<process>", "39869c73-5ff0-4ae6-8317-eb494c87167b")
						c.ConsumesEventType(MessageE{})   // E = event
						c.ProducesCommandType(MessageC{}) // C = command
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
						s.ExecuteCommand(MessageC1)
						s.ExecuteCommand(MessageC2)
						s.ExecuteCommand(MessageC3)
						return nil
					},
				})
			},
		}
	})

	DescribeTable(
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
		Entry(
			"all executed commands match",
			RecordEvent(MessageE1),
			ToOnlyExecuteCommandsMatching(
				func(m dogma.Message) error {
					return nil
				},
			),
			expectPass,
			expectReport(
				`✓ only execute commands that match the predicate near expectation.messagematch.commandonly_test.go:79`,
			),
		),
		Entry(
			"no commands executed at all",
			noop,
			ToOnlyExecuteCommandsMatching(
				func(m dogma.Message) error {
					panic("unexpected call")
				},
			),
			expectPass,
			expectReport(
				`✓ only execute commands that match the predicate near expectation.messagematch.commandonly_test.go:91`,
			),
		),
		Entry(
			"some matching commands executed",
			RecordEvent(MessageE1),
			ToOnlyExecuteCommandsMatching(
				func(m dogma.Message) error {
					switch m {
					case fixtures.MessageC1:
						return errors.New("<error>")
					case fixtures.MessageC2:
						return IgnoreMessage
					default:
						return nil
					}
				},
			),
			expectFail,
			expectReport(
				`✗ only execute commands that match the predicate near expectation.messagematch.commandonly_test.go:104`,
				``,
				`  | EXPLANATION`,
				`  |     only 1 of 2 relevant commands matched the predicate`,
				`  | `,
				`  | FAILED MATCHES`,
				`  |     • fixtures.MessageC: <error>`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • verify the logic within the predicate function, it ignored 1 command`,
				`  |     • verify the logic within the '<process>' process message handler`,
			),
		),
	)

	It("panics if the function is nil", func() {
		Expect(func() {
			ToOnlyExecuteCommandsMatching(nil)
		}).To(PanicWith("ToOnlyExecuteCommandsMatching(<nil>): function must not be nil"))
	})
})
