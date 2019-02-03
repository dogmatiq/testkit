package integration_test

import (
	"context"
	"errors"
	"time"

	"github.com/dogmatiq/dogmatest/engine/controller"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogmatest/engine/controller/integration"
	"github.com/dogmatiq/dogmatest/engine/envelope"
	"github.com/dogmatiq/dogmatest/engine/fact"
	"github.com/dogmatiq/enginekit/fixtures"
	handlerkit "github.com/dogmatiq/enginekit/handler"
	"github.com/dogmatiq/enginekit/message"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ controller.Controller = &Controller{}

var _ = Describe("type Controller", func() {
	var (
		messageIDs envelope.MessageIDGenerator
		handler    *fixtures.IntegrationMessageHandler
		controller *Controller
		command    *envelope.Envelope
	)

	BeforeEach(func() {
		command = envelope.New(
			"1000",
			fixtures.MessageA1,
			message.CommandRole,
		)

		handler = &fixtures.IntegrationMessageHandler{}

		controller = NewController("<name>", handler, &messageIDs)

		messageIDs.Reset() // reset after setup for a predictable ID.
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

	Describe("func Tick()", func() {
		It("does not return any envelopes", func() {
			envelopes, err := controller.Tick(
				context.Background(),
				fact.Ignore,
				time.Now(),
			)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(envelopes).To(BeEmpty())
		})

		It("does not record any facts", func() {
			buf := &fact.Buffer{}
			_, err := controller.Tick(
				context.Background(),
				buf,
				time.Now(),
			)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(buf.Facts()).To(BeEmpty())
		})
	})

	Describe("func Handle()", func() {
		It("forwards the message to the handler", func() {
			called := false
			handler.HandleCommandFunc = func(
				_ context.Context,
				_ dogma.IntegrationCommandScope,
				m dogma.Message,
			) error {
				called = true
				Expect(m).To(Equal(fixtures.MessageA1))
				return nil
			}

			_, err := controller.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				command,
			)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(called).To(BeTrue())
		})

		It("returns the recorded events", func() {
			handler.HandleCommandFunc = func(
				_ context.Context,
				s dogma.IntegrationCommandScope,
				_ dogma.Message,
			) error {
				s.RecordEvent(fixtures.MessageB1)
				s.RecordEvent(fixtures.MessageB2)
				return nil
			}

			events, err := controller.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				command,
			)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(events).To(ConsistOf(
				command.NewEvent(
					"1",
					fixtures.MessageB1,
					envelope.Origin{
						HandlerName: "<name>",
						HandlerType: handlerkit.IntegrationType,
					},
				),
				command.NewEvent(
					"2",
					fixtures.MessageB2,
					envelope.Origin{
						HandlerName: "<name>",
						HandlerType: handlerkit.IntegrationType,
					},
				),
			))
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
				time.Now(),
				command,
			)

			Expect(err).To(Equal(expected))
		})
	})

	Describe("func Reset()", func() {
		It("does nothing", func() {
			controller.Reset()
		})
	})
})
