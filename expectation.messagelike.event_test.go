package testkit_test

import (
	"context"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/enginekit/enginetest/stubs"
	. "github.com/dogmatiq/testkit"
	"github.com/dogmatiq/testkit/internal/testingmock"
	g "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = g.Describe("func ToRecordEventLike()", func() {
	var (
		testingT *testingmock.T
		app      dogma.Application
	)

	type RecordAnEvent = CommandStub[dogma.Event]

	g.BeforeEach(func() {
		testingT = &testingmock.T{
			FailSilently: true,
		}

		app = &ApplicationStub{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "61cb6dbe-8473-4a9f-a1d5-dd13d6e104b4")

				c.RegisterIntegration(&IntegrationMessageHandlerStub{
					ConfigureFunc: func(c dogma.IntegrationConfigurer) {
						c.Identity("<integration>", "39cc8b78-fcde-49ef-b85f-d0d145b567de")
						c.Routes(
							dogma.HandlesCommand[RecordAnEvent](),
							dogma.RecordsEvent[EventStub[TypeA]](),
							dogma.RecordsEvent[EventStub[TypeB]](),
						)
					},
					HandleCommandFunc: func(
						_ context.Context,
						s dogma.IntegrationCommandScope,
						m dogma.Command,
					) error {
						if m, ok := m.(RecordAnEvent); ok {
							s.RecordEvent(m.Content)
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
		) {
			Begin(testingT, app).
				EnableHandlers("<integration>").
				Expect(a, e)

			rm(testingT)
			Expect(testingT.Failed()).To(Equal(!ok))
		},
		g.Entry(
			"exact match event recorded",
			ExecuteCommand(
				RecordAnEvent{
					Content: EventA1,
				},
			),
			ToRecordEventLike(EventA1),
			expectPass,
			expectReport(
				`✓ record an event that is a superset of a specific 'stubs.EventStub[TypeA]' event`,
			),
		),
		g.Entry(
			"no matching event recorded",
			ExecuteCommand(
				RecordAnEvent{
					Content: EventB1,
				},
			),
			ToRecordEventLike(EventA1),
			expectFail,
			expectReport(
				`✗ record an event that is a superset of a specific 'stubs.EventStub[TypeA]' event`,
				``,
				`  | EXPLANATION`,
				`  |     none of the engaged handlers recorded a matching event`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • enable integration handlers using the EnableHandlerType() option`,
				`  |     • verify the logic within the '<integration>' integration message handler`,
			),
		),
		g.Entry(
			"no messages produced at all",
			noop,
			ToRecordEventLike(EventA1),
			expectFail,
			expectReport(
				`✗ record an event that is a superset of a specific 'stubs.EventStub[TypeA]' event`,
				``,
				`  | EXPLANATION`,
				`  |     no messages were produced at all`,
			),
		),
	)

	g.It("panics if the message is nil", func() {
		Expect(func() {
			ToRecordEventLike(nil)
		}).To(PanicWith("ToRecordEventLike(<nil>): message must not be nil"))
	})

	g.It("does not panic if the message is invalid", func() {
		Expect(func() {
			ToRecordEventLike(EventStub[TypeA]{
				ValidationError: "<error>",
			})
		}).NotTo(Panic())
	})
})
