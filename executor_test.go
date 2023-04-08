package testkit_test

import (
	"context"
	"errors"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogma/fixtures"
	. "github.com/dogmatiq/testkit"
	"github.com/dogmatiq/testkit/internal/testingmock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("func InterceptCommandExecutor()", func() {
	var (
		testingT                     *testingmock.T
		app                          dogma.Application
		doNothing                    CommandExecutorInterceptor
		executeCommandAndReturnError CommandExecutorInterceptor
	)

	BeforeEach(func() {
		testingT = &testingmock.T{}

		app = &Application{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "b5453327-a0fa-4e94-bb46-8464e727c4fd")

				c.RegisterIntegration(&IntegrationMessageHandler{
					ConfigureFunc: func(c dogma.IntegrationConfigurer) {
						c.Identity("<handler-name>", "67c167a8-d09e-4827-beab-7c8c9817bb1a")
						c.ConsumesCommandType(MessageC{})
						c.ProducesEventType(MessageE{})
					},
					HandleCommandFunc: func(
						_ context.Context,
						s dogma.IntegrationCommandScope,
						_ dogma.Message,
					) error {
						s.RecordEvent(MessageE1)
						return nil
					},
				})
			},
		}

		doNothing = func(
			context.Context,
			dogma.Message,
			dogma.CommandExecutor,
		) error {
			return nil
		}

		executeCommandAndReturnError = func(
			ctx context.Context,
			m dogma.Message,
			e dogma.CommandExecutor,
		) error {
			Expect(m).To(Equal(MessageC1))

			err := e.ExecuteCommand(ctx, m)
			Expect(err).ShouldNot(HaveOccurred())

			return errors.New("<error>")
		}
	})

	It("panics if the interceptor function is nil", func() {
		Expect(func() {
			InterceptCommandExecutor(nil)
		}).To(PanicWith("InterceptCommandExecutor(<nil>): function must not be nil"))
	})

	When("used as a TestOption", func() {
		It("intercepts calls to ExecuteCommand()", func() {
			test := Begin(
				testingT,
				app,
				InterceptCommandExecutor(executeCommandAndReturnError),
			)

			test.EnableHandlers("<handler-name>")

			test.Expect(
				Call(func() {
					err := test.CommandExecutor().ExecuteCommand(
						context.Background(),
						MessageC1,
					)
					Expect(err).To(MatchError("<error>"))
				}),
				ToRecordEvent(MessageE1),
			)
		})
	})

	When("used as a CallOption", func() {
		It("intercepts calls to ExecuteCommand()", func() {
			test := Begin(
				&testingmock.T{},
				app,
			)

			test.EnableHandlers("<handler-name>")

			test.Expect(
				Call(
					func() {
						err := test.CommandExecutor().ExecuteCommand(
							context.Background(),
							MessageC1,
						)
						Expect(err).To(MatchError("<error>"))
					},
					InterceptCommandExecutor(executeCommandAndReturnError),
				),
				ToRecordEvent(MessageE1),
			)
		})

		It("uninstalls the interceptor upon completion of the Call() action", func() {
			test := Begin(
				&testingmock.T{},
				app,
			)

			test.Prepare(
				Call(
					func() {
						err := test.CommandExecutor().ExecuteCommand(
							context.Background(),
							MessageC1,
						)
						Expect(err).To(MatchError("<error>"))
					},
					InterceptCommandExecutor(executeCommandAndReturnError),
				),
				Call(
					func() {
						err := test.CommandExecutor().ExecuteCommand(
							context.Background(),
							MessageC1,
						)
						Expect(err).ShouldNot(HaveOccurred())
					},
				),
			)
		})

		It("re-installs the test-level interceptor upon completion of the Call() action", func() {
			test := Begin(
				&testingmock.T{},
				app,
				InterceptCommandExecutor(executeCommandAndReturnError),
			)

			test.Prepare(
				Call(
					func() {
						err := test.CommandExecutor().ExecuteCommand(
							context.Background(),
							MessageC1,
						)
						Expect(err).ShouldNot(HaveOccurred())
					},
					InterceptCommandExecutor(doNothing),
				),
				Call(
					func() {
						err := test.CommandExecutor().ExecuteCommand(
							context.Background(),
							MessageC1,
						)
						Expect(err).To(MatchError("<error>"))
					},
				),
			)
		})
	})
})
