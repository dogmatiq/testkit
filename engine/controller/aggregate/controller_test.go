package aggregate_test

import (
	"context"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogmatest/engine/controller/aggregate"
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
		handler    *fixtures.AggregateMessageHandler
		controller *Controller
		command    = envelope.New(
			fixtures.MessageA1,
			message.CommandRole,
		)
	)

	BeforeEach(func() {
		handler = &fixtures.AggregateMessageHandler{
			RouteCommandToInstanceFunc: func(m dogma.Message) string {
				switch m.(type) {
				case fixtures.MessageA:
					return "<instance>"
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
		It("returns handler.AggregateType", func() {
			Expect(controller.Type()).To(Equal(handlerkit.AggregateType))
		})
	})

	Describe("func Handle()", func() {
		It("forwards the message to the handler", func() {
			called := false
			handler.HandleCommandFunc = func(
				_ dogma.AggregateCommandScope,
				hm dogma.Message,
			) {
				called = true
				Expect(hm).To(Equal(fixtures.MessageA1))
			}

			_, err := controller.Handle(
				context.Background(),
				fact.Ignore,
				command,
			)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(called).To(BeTrue())
		})

		When("the handler returns an empty instance ID", func() {
			It("panics when the handler routes to an empty instance ID", func() {
				handler.RouteCommandToInstanceFunc = func(m dogma.Message) string {
					return ""
				}

				Expect(func() {
					controller.Handle(
						context.Background(),
						fact.Ignore,
						command,
					)
				}).To(Panic())
			})
		})

		When("the instance does not exist", func() {
			It("records a fact", func() {
				buf := &fact.Buffer{}
				_, err := controller.Handle(
					context.Background(),
					buf,
					command,
				)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(buf.Facts).To(ContainElement(
					fact.AggregateInstanceNotFound{
						HandlerName: "<name>",
						InstanceID:  "<instance>",
						Envelope:    command,
					},
				))
			})

			It("panics if New() returns nil", func() {
				handler.NewFunc = func() dogma.AggregateRoot {
					return nil
				}

				Expect(func() {
					controller.Handle(
						context.Background(),
						fact.Ignore,
						command,
					)
				}).To(Panic())
			})
		})

		When("the instance exists", func() {
			BeforeEach(func() {
				handler.HandleCommandFunc = func(
					s dogma.AggregateCommandScope,
					m dogma.Message,
				) {
					s.Create()
					s.RecordEvent(fixtures.MessageE1) // event must be recorded when creating
				}

				_, err := controller.Handle(
					context.Background(),
					fact.Ignore,
					envelope.New(
						fixtures.MessageA2, // use a different message to create the instance
						message.CommandRole,
					),
				)

				Expect(err).ShouldNot(HaveOccurred())
			})

			It("records a fact", func() {
				buf := &fact.Buffer{}
				_, err := controller.Handle(
					context.Background(),
					buf,
					command,
				)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(buf.Facts).To(ContainElement(
					fact.AggregateInstanceLoaded{
						HandlerName: "<name>",
						InstanceID:  "<instance>",
						Root:        &fixtures.AggregateRoot{},
						Envelope:    command,
					},
				))
			})

			It("does not call New()", func() {
				handler.NewFunc = func() dogma.AggregateRoot {
					Fail("expected call to New()")
					return nil
				}

				controller.Handle(
					context.Background(),
					fact.Ignore,
					command,
				)
			})
		})
	})

	Describe("func Reset()", func() {
		BeforeEach(func() {
			handler.HandleCommandFunc = func(
				s dogma.AggregateCommandScope,
				m dogma.Message,
			) {
				s.Create()
				s.RecordEvent(fixtures.MessageE1) // event must be recorded when creating
			}

			_, err := controller.Handle(
				context.Background(),
				fact.Ignore,
				command,
			)

			Expect(err).ShouldNot(HaveOccurred())
		})

		It("removes all instances", func() {
			controller.Reset()

			buf := &fact.Buffer{}
			_, err := controller.Handle(
				context.Background(),
				buf,
				command,
			)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(buf.Facts).NotTo(ContainElement(
				BeAssignableToTypeOf(fact.AggregateInstanceLoaded{}),
			))
		})
	})
})
