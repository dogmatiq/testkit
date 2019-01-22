package process_test

import (
	"context"
	"errors"
	"time"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogmatest/engine/controller/process"
	"github.com/dogmatiq/dogmatest/engine/envelope"
	"github.com/dogmatiq/dogmatest/engine/fact"
	"github.com/dogmatiq/dogmatest/internal/enginekit/fixtures"
	handlerkit "github.com/dogmatiq/dogmatest/internal/enginekit/handler"
	"github.com/dogmatiq/dogmatest/internal/enginekit/message"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("type Controller", func() {
	var (
		handler    *fixtures.ProcessMessageHandler
		controller *Controller
		event      = envelope.New(
			fixtures.MessageA1,
			message.EventRole,
		)
	)

	BeforeEach(func() {
		handler = &fixtures.ProcessMessageHandler{
			RouteEventToInstanceFunc: func(
				_ context.Context,
				m dogma.Message,
			) (string, bool, error) {
				switch m.(type) {
				case fixtures.MessageA:
					return "<instance>", true, nil
				default:
					panic(dogma.UnexpectedMessage)
				}
			},
		}
		controller = NewController("<name>", handler)
	})

	Describe("func Name()", func() {
		It("returns the handler name", func() {
			Expect(controller.Name()).To(Equal("<name>"))
		})
	})

	Describe("func Type()", func() {
		It("returns handler.ProcessType", func() {
			Expect(controller.Type()).To(Equal(handlerkit.ProcessType))
		})
	})

	Describe("func Handle()", func() {
		When("handling an event", func() {
			It("forwards the message to the handler", func() {
				called := false
				handler.HandleEventFunc = func(
					_ context.Context,
					_ dogma.ProcessEventScope,
					m dogma.Message,
				) error {
					called = true
					Expect(m).To(Equal(fixtures.MessageA1))
					return nil
				}

				_, err := controller.Handle(
					context.Background(),
					fact.Ignore,
					event,
				)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(called).To(BeTrue())
			})

			It("propagates handler errors", func() {
				expected := errors.New("<error>")

				handler.HandleEventFunc = func(
					_ context.Context,
					_ dogma.ProcessEventScope,
					_ dogma.Message,
				) error {
					return expected
				}

				_, err := controller.Handle(
					context.Background(),
					fact.Ignore,
					event,
				)

				Expect(err).To(Equal(expected))
			})

			It("returns the executed commands and scheduled timeouts", func() {
				t := time.Now()

				handler.HandleEventFunc = func(
					_ context.Context,
					s dogma.ProcessEventScope,
					_ dogma.Message,
				) error {
					s.Begin()
					s.ExecuteCommand(fixtures.MessageB1)
					s.ScheduleTimeout(fixtures.MessageB2, t)
					return nil
				}

				commands, err := controller.Handle(
					context.Background(),
					fact.Ignore,
					event,
				)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(commands).To(ConsistOf(
					event.NewCommand(
						fixtures.MessageB1,
					),
					event.NewTimeout(
						fixtures.MessageB2,
						t,
					),
				))
			})

			When("the event is not routed to an instance", func() {
				XIt("does not forward the message to the handler", func() {
				})

				XIt("records a fact", func() {
				})
			})
		})

		When("handling a timeout", func() {
			XIt("forwards the message to the handler", func() {
			})

			XIt("propagates handler errors", func() {
			})

			XIt("returns the executed commands and scheduled timeouts", func() {
			})
		})

		XIt("propagates routing errors", func() {
		})

		It("panics when the handler routes to an empty instance ID", func() {
			handler.RouteEventToInstanceFunc = func(
				context.Context,
				dogma.Message,
			) (string, bool, error) {
				return "", true, nil
			}

			Expect(func() {
				controller.Handle(
					context.Background(),
					fact.Ignore,
					event,
				)
			}).To(Panic())
		})

		When("the instance does not exist", func() {
			It("records a fact", func() {
				buf := &fact.Buffer{}
				_, err := controller.Handle(
					context.Background(),
					buf,
					event,
				)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(buf.Facts).To(ContainElement(
					fact.ProcessInstanceNotFound{
						HandlerName: "<name>",
						InstanceID:  "<instance>",
						Envelope:    event,
					},
				))
			})

			It("panics if New() returns nil", func() {
				handler.NewFunc = func() dogma.ProcessRoot {
					return nil
				}

				Expect(func() {
					controller.Handle(
						context.Background(),
						fact.Ignore,
						event,
					)
				}).To(Panic())
			})
		})

		When("the instance exists", func() {
			BeforeEach(func() {
				handler.HandleEventFunc = func(
					_ context.Context,
					s dogma.ProcessEventScope,
					_ dogma.Message,
				) error {
					s.Begin()
					return nil
				}

				_, err := controller.Handle(
					context.Background(),
					fact.Ignore,
					envelope.New(
						fixtures.MessageA2, // use a different message to begin the instance
						message.EventRole,
					),
				)

				Expect(err).ShouldNot(HaveOccurred())
			})

			It("records a fact", func() {
				buf := &fact.Buffer{}
				_, err := controller.Handle(
					context.Background(),
					buf,
					event,
				)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(buf.Facts).To(ContainElement(
					fact.ProcessInstanceLoaded{
						HandlerName: "<name>",
						InstanceID:  "<instance>",
						Root:        &fixtures.ProcessRoot{},
						Envelope:    event,
					},
				))
			})

			It("does not call New()", func() {
				handler.NewFunc = func() dogma.ProcessRoot {
					Fail("expected call to New()")
					return nil
				}

				controller.Handle(
					context.Background(),
					fact.Ignore,
					event,
				)
			})
		})
	})

	Describe("func Reset()", func() {
		BeforeEach(func() {
			handler.HandleEventFunc = func(
				_ context.Context,
				s dogma.ProcessEventScope,
				m dogma.Message,
			) error {
				s.Begin()
				return nil
			}

			_, err := controller.Handle(
				context.Background(),
				fact.Ignore,
				event,
			)

			Expect(err).ShouldNot(HaveOccurred())
		})

		It("removes all instances", func() {
			controller.Reset()

			buf := &fact.Buffer{}
			_, err := controller.Handle(
				context.Background(),
				buf,
				event,
			)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(buf.Facts).NotTo(ContainElement(
				BeAssignableToTypeOf(fact.ProcessInstanceLoaded{}),
			))
		})
	})
})
