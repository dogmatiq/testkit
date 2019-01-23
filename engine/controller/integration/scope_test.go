package integration_test

import (
	"context"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogmatest/engine/controller/integration"
	"github.com/dogmatiq/dogmatest/engine/envelope"
	"github.com/dogmatiq/dogmatest/engine/fact"
	"github.com/dogmatiq/dogmatest/internal/enginekit/fixtures"
	"github.com/dogmatiq/dogmatest/internal/enginekit/message"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("type commandScope", func() {
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

	Describe("func RecordEvent", func() {
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

		It("records a fact", func() {
			buf := &fact.Buffer{}
			_, _, err := controller.Handle(
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

	Describe("func Log", func() {
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
			_, _, err := controller.Handle(
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
