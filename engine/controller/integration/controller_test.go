package integration_test

import (
	"context"
	"errors"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogmatest/engine/controller/integration"
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
		handler    *fixtures.IntegrationMessageHandler
		controller *Controller
		command    = envelope.New(
			fixtures.MessageA1,
			message.CommandRole,
		)
	)

	BeforeEach(func() {
		handler = &fixtures.IntegrationMessageHandler{}
		controller = NewController("<name>", handler)
	})

	Describe("func Name()", func() {
		It("returns the handler name", func() {
			Expect(controller.Name()).To(Equal("<name>"))
		})
	})

	Describe("func Type()", func() {
		It("returns handler.IntegrationType", func() {
			Expect(controller.Type()).To(Equal(handlerkit.IntegrationType))
		})
	})

	Describe("func Handle()", func() {
		It("forwards the message to the handler", func() {
			called := false
			handler.HandleCommandFunc = func(
				_ context.Context,
				_ dogma.IntegrationCommandScope,
				hm dogma.Message,
			) error {
				called = true
				Expect(hm).To(Equal(fixtures.MessageA1))
				return nil
			}

			_, err := controller.Handle(
				context.Background(),
				fact.Ignore,
				command,
			)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(called).To(BeTrue())
		})

		It("propagates handler errors", func() {
			expected := errors.New("<error>")

			handler.HandleCommandFunc = func(
				_ context.Context,
				_ dogma.IntegrationCommandScope,
				_ dogma.Message,
			) error {
				return expected
			}

			_, err := controller.Handle(
				context.Background(),
				fact.Ignore,
				command,
			)

			Expect(err).To(Equal(expected))
		})

		When("the handler records an event", func() {
			event := command.NewEvent(
				fixtures.MessageB1,
			)

			BeforeEach(func() {
				handler.HandleCommandFunc = func(
					_ context.Context,
					s dogma.IntegrationCommandScope,
					_ dogma.Message,
				) error {
					s.RecordEvent(fixtures.MessageB1)
					return nil
				}
			})

			It("returns the recorded event", func() {
				events, err := controller.Handle(
					context.Background(),
					fact.Ignore,
					command,
				)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(events).To(ConsistOf(event))
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
					fact.EventRecordedByIntegration{
						HandlerName:   "<name>",
						Envelope:      command,
						EventEnvelope: event,
					},
				))
			})
		})

		When("the handler logs a message", func() {
			BeforeEach(func() {
				handler.HandleCommandFunc = func(
					_ context.Context,
					s dogma.IntegrationCommandScope,
					_ dogma.Message,
				) error {
					s.Log("<format>", "<arg-1>", "<arg-2>")
					return nil
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
					fact.MessageLoggedByIntegration{
						HandlerName: "<name>",
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
