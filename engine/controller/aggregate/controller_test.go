package aggregate_test

import (
	"context"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogmatest/engine/controller/aggregate"
	"github.com/dogmatiq/dogmatest/engine/envelope"
	"github.com/dogmatiq/dogmatest/engine/fact"
	handlerkit "github.com/dogmatiq/dogmatest/internal/enginekit/handler"
	"github.com/dogmatiq/dogmatest/internal/enginekit/message"
	"github.com/dogmatiq/dogmatest/internal/fixtures"
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

		When("the handler records an event", func() {
			event := command.NewEvent(
				fixtures.MessageB1,
			)

			BeforeEach(func() {
				handler.HandleCommandFunc = func(
					s dogma.AggregateCommandScope,
					_ dogma.Message,
				) {
					s.Create() // must be created first
					s.RecordEvent(fixtures.MessageB1)
				}
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
					fact.EventRecordedByAggregate{
						HandlerName:   "<name>",
						InstanceID:    "<instance>",
						Root:          &fixtures.AggregateRoot{},
						Envelope:      command,
						EventEnvelope: event,
					},
				))
			})
		})

		When("the handler logs a message", func() {
			BeforeEach(func() {
				handler.HandleCommandFunc = func(
					s dogma.AggregateCommandScope,
					_ dogma.Message,
				) {
					s.Log("<format>", "<arg-1>", "<arg-2>")
				}
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
					fact.MessageLoggedByAggregate{
						HandlerName: "<name>",
						InstanceID:  "<instance>",
						Root:        &fixtures.AggregateRoot{},
						Envelope:    command,
						LogFormat:   "<format>",
						LogArguments: []interface{}{
							"<arg-1>",
							"<arg-2>",
						},
					},
				))
			})
		})
	})
})
