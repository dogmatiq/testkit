package testkit_test

import (
	"context"
	"errors"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/enginekit/enginetest/stubs"
	. "github.com/dogmatiq/testkit"
	"github.com/dogmatiq/testkit/internal/testingmock"
	g "github.com/onsi/ginkgo/v2"
	gm "github.com/onsi/gomega"
)

var _ = g.Describe("func InterceptCommandExecutor()", func() {
	var (
		testingT                     *testingmock.T
		app                          dogma.Application
		doNothing                    CommandExecutorInterceptor
		executeCommandAndReturnError CommandExecutorInterceptor
	)

	g.BeforeEach(func() {
		testingT = &testingmock.T{}

		app = &ApplicationStub{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "b5453327-a0fa-4e94-bb46-8464e727c4fd")
				c.Routes(
					dogma.ViaIntegration(&IntegrationMessageHandlerStub{
						ConfigureFunc: func(c dogma.IntegrationConfigurer) {
							c.Identity("<handler-name>", "67c167a8-d09e-4827-beab-7c8c9817bb1a")
							c.Routes(
								dogma.HandlesCommand[*CommandStub[TypeA]](),
								dogma.RecordsEvent[*EventStub[TypeA]](),
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

		doNothing = func(
			context.Context,
			dogma.Command,
			dogma.CommandExecutor,
		) error {
			return nil
		}

		executeCommandAndReturnError = func(
			ctx context.Context,
			m dogma.Command,
			e dogma.CommandExecutor,
		) error {
			gm.Expect(m).To(gm.Equal(CommandA1))

			err := e.ExecuteCommand(ctx, m)
			gm.Expect(err).ShouldNot(gm.HaveOccurred())

			return errors.New("<error>")
		}
	})

	g.It("panics if the interceptor function is nil", func() {
		gm.Expect(func() {
			InterceptCommandExecutor(nil)
		}).To(gm.PanicWith("InterceptCommandExecutor(<nil>): function must not be nil"))
	})

	g.When("used as a TestOption", func() {
		g.It("intercepts calls to ExecuteCommand()", func() {
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
						CommandA1,
					)
					gm.Expect(err).To(gm.MatchError("<error>"))
				}),
				ToRecordEvent(EventA1),
			)
		})
	})

	g.When("used as a CallOption", func() {
		g.It("intercepts calls to ExecuteCommand()", func() {
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
							CommandA1,
						)
						gm.Expect(err).To(gm.MatchError("<error>"))
					},
					InterceptCommandExecutor(executeCommandAndReturnError),
				),
				ToRecordEvent(EventA1),
			)
		})

		g.It("uninstalls the interceptor upon completion of the Call() action", func() {
			test := Begin(
				&testingmock.T{},
				app,
			)

			test.Prepare(
				Call(
					func() {
						err := test.CommandExecutor().ExecuteCommand(
							context.Background(),
							CommandA1,
						)
						gm.Expect(err).To(gm.MatchError("<error>"))
					},
					InterceptCommandExecutor(executeCommandAndReturnError),
				),
				Call(
					func() {
						err := test.CommandExecutor().ExecuteCommand(
							context.Background(),
							CommandA1,
						)
						gm.Expect(err).ShouldNot(gm.HaveOccurred())
					},
				),
			)
		})

		g.It("re-installs the test-level interceptor upon completion of the Call() action", func() {
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
							CommandA1,
						)
						gm.Expect(err).ShouldNot(gm.HaveOccurred())
					},
					InterceptCommandExecutor(doNothing),
				),
				Call(
					func() {
						err := test.CommandExecutor().ExecuteCommand(
							context.Background(),
							CommandA1,
						)
						gm.Expect(err).To(gm.MatchError("<error>"))
					},
				),
			)
		})
	})
})
