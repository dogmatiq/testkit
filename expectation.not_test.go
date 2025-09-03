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

var _ = g.Context("not expectation", func() {
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
				c.Identity("<app>", "00df8612-2fd4-4ae3-9acf-afc2b4daf272")
				c.Routes(
					dogma.ViaIntegration(&IntegrationMessageHandlerStub{
						ConfigureFunc: func(c dogma.IntegrationConfigurer) {
							c.Identity("<integration>", "12dfb90b-e47b-4d49-b834-294b01992ad0")
							c.Routes(
								dogma.HandlesCommand[CommandStub[TypeA]](),
								dogma.RecordsEvent[EventStub[TypeA]](),
							)
						},
						HandleCommandFunc: func(
							_ context.Context,
							s dogma.IntegrationCommandScope,
							_ dogma.Command,
						) error {
							s.RecordEvent(EventA1)
							return nil
						},
					}),
				)
			},
		}

		test = Begin(testingT, app).
			EnableHandlers("<integration>")
	})

	testExpectationBehavior := func(
		e Expectation,
		ok bool,
		rm reportMatcher,
	) {
		test.Expect(ExecuteCommand(CommandA1), e)
		rm(testingT)
		gm.Expect(testingT.Failed()).To(gm.Equal(!ok))
	}

	g.Describe("func Not()", func() {
		g.DescribeTable(
			"expectation behavior",
			testExpectationBehavior,
			g.Entry(
				"it fails when the child expectation passes",
				Not(ToRecordEvent(EventA1)),
				expectFail,
				expectReport(
					`✗ do not record a specific 'stubs.EventStub[TypeA]' event`,
				),
			),
			g.Entry(
				"it passes when the child expectation fails",
				Not(ToRecordEvent(EventA2)),
				expectPass,
				expectReport(
					`✓ do not record a specific 'stubs.EventStub[TypeA]' event`,
				),
			),
		)

		g.It("produces the expected caption", func() {
			test.Expect(
				noop,
				Not(ToRecordEvent(EventA2)),
			)

			gm.Expect(testingT.Logs).To(gm.ContainElement(
				"--- expect [no-op] not to record a specific 'stubs.EventStub[TypeA]' event ---",
			))
		})

		g.It("fails the test if the child cannot construct a predicate", func() {
			test.Expect(
				noop,
				Not(failBeforeAction),
			)

			gm.Expect(testingT.Logs).To(gm.ContainElement("<always fail before action>"))
			gm.Expect(testingT.Failed()).To(gm.BeTrue())
		})
	})
})
