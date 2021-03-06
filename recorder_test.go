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

var _ = Describe("func InterceptEventRecorder()", func() {
	var (
		testingT                  *testingmock.T
		app                       dogma.Application
		doNothing                 EventRecorderInterceptor
		recordEventAndReturnError EventRecorderInterceptor
	)

	BeforeEach(func() {
		testingT = &testingmock.T{}

		app = &Application{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "<app-key>")

				c.RegisterProcess(&ProcessMessageHandler{
					ConfigureFunc: func(c dogma.ProcessConfigurer) {
						c.Identity("<handler-name>", "<handler-key>")
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

	It("panics if the interceptor function is nil", func() {
		Expect(func() {
			InterceptEventRecorder(nil)
		}).To(PanicWith("InterceptEventRecorder(<nil>): function must not be nil"))
	})

	When("used as a TestOption", func() {
		It("intercepts calls to RecordEvent()", func() {
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

	When("used as a CallOption", func() {
		It("intercepts calls to RecordEvent()", func() {
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

		It("uninstalls the interceptor upon completion of the Call() action", func() {
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

		It("re-installs the test-level interceptor upon completion of the Call() action", func() {
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
