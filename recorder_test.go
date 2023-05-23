package testkit_test

import (
	"context"
	"errors"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogma/fixtures"
	. "github.com/dogmatiq/testkit"
	"github.com/dogmatiq/testkit/internal/testingmock"
	g "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = g.Describe("func InterceptEventRecorder()", func() {
	var (
		testingT                  *testingmock.T
		app                       dogma.Application
		doNothing                 EventRecorderInterceptor
		recordEventAndReturnError EventRecorderInterceptor
	)

	g.BeforeEach(func() {
		testingT = &testingmock.T{}

		app = &Application{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "01fdba5c-9010-4b6e-9cc4-bdb63f95c423")

				c.RegisterProcess(&ProcessMessageHandler{
					ConfigureFunc: func(c dogma.ProcessConfigurer) {
						c.Identity("<handler-name>", "531c9909-ebcc-4214-8dba-b0a3a93ae5b4")
						c.ConsumesEventType(MessageE{})
						c.ProducesCommandType(MessageC{})
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
						_ dogma.Message,
					) error {
						s.ExecuteCommand(MessageC1)
						return nil
					},
				})
			},
		}

		doNothing = func(
			context.Context,
			dogma.Message,
			dogma.EventRecorder,
		) error {
			return nil
		}

		recordEventAndReturnError = func(
			ctx context.Context,
			m dogma.Message,
			e dogma.EventRecorder,
		) error {
			Expect(m).To(Equal(MessageE1))

			err := e.RecordEvent(ctx, m)
			Expect(err).ShouldNot(HaveOccurred())

			return errors.New("<error>")
		}
	})

	g.It("panics if the interceptor function is nil", func() {
		Expect(func() {
			InterceptEventRecorder(nil)
		}).To(PanicWith("InterceptEventRecorder(<nil>): function must not be nil"))
	})

	g.When("used as a TestOption", func() {
		g.It("intercepts calls to RecordEvent()", func() {
			test := Begin(
				testingT,
				app,
				InterceptEventRecorder(recordEventAndReturnError),
			)

			test.EnableHandlers("<handler-name>")

			test.Expect(
				Call(func() {
					err := test.EventRecorder().RecordEvent(
						context.Background(),
						MessageE1,
					)
					Expect(err).To(MatchError("<error>"))
				}),
				ToExecuteCommand(MessageC1),
			)
		})
	})

	g.When("used as a CallOption", func() {
		g.It("intercepts calls to RecordEvent()", func() {
			test := Begin(
				&testingmock.T{},
				app,
			)

			test.EnableHandlers("<handler-name>")

			test.Expect(
				Call(
					func() {
						err := test.EventRecorder().RecordEvent(
							context.Background(),
							MessageE1,
						)
						Expect(err).To(MatchError("<error>"))
					},
					InterceptEventRecorder(recordEventAndReturnError),
				),
				ToExecuteCommand(MessageC1),
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
						err := test.EventRecorder().RecordEvent(
							context.Background(),
							MessageE1,
						)
						Expect(err).To(MatchError("<error>"))
					},
					InterceptEventRecorder(recordEventAndReturnError),
				),
				Call(
					func() {
						err := test.EventRecorder().RecordEvent(
							context.Background(),
							MessageE1,
						)
						Expect(err).ShouldNot(HaveOccurred())
					},
				),
			)
		})

		g.It("re-installs the test-level interceptor upon completion of the Call() action", func() {
			test := Begin(
				&testingmock.T{},
				app,
				InterceptEventRecorder(recordEventAndReturnError),
			)

			test.Prepare(
				Call(
					func() {
						err := test.EventRecorder().RecordEvent(
							context.Background(),
							MessageE1,
						)
						Expect(err).ShouldNot(HaveOccurred())
					},
					InterceptEventRecorder(doNothing),
				),
				Call(
					func() {
						err := test.EventRecorder().RecordEvent(
							context.Background(),
							MessageE1,
						)
						Expect(err).To(MatchError("<error>"))
					},
				),
			)
		})
	})
})
