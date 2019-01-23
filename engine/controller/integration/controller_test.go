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
	"github.com/dogmatiq/dogmatest/internal/enginekit/fixtures"
	handlerkit "github.com/dogmatiq/dogmatest/internal/enginekit/handler"
	"github.com/dogmatiq/dogmatest/internal/enginekit/message"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ controller.Controller = &Controller{}

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

	Describe("func Tick()", func() {
		It("does not return an error", func() {
			err := controller.Tick(context.Background(), time.Now())
			Expect(err).ShouldNot(HaveOccurred())
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
				command,
			)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(events).To(ConsistOf(
				command.NewEvent(
					fixtures.MessageB1,
				),
				command.NewEvent(
					fixtures.MessageB2,
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
