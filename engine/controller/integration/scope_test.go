package integration_test

import (
	"context"
	"time"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/enginekit/fixtures"
	handlerkit "github.com/dogmatiq/enginekit/handler"
	"github.com/dogmatiq/enginekit/message"
	. "github.com/dogmatiq/testkit/engine/controller/integration"
	"github.com/dogmatiq/testkit/engine/envelope"
	"github.com/dogmatiq/testkit/engine/fact"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("type scope", func() {
	var (
		messageIDs envelope.MessageIDGenerator
		handler    *fixtures.IntegrationMessageHandler
		controller *Controller
		command    *envelope.Envelope
	)

	BeforeEach(func() {
		command = envelope.NewCommand(
			"1000",
			fixtures.MessageA1,
			time.Now(),
		)

		handler = &fixtures.IntegrationMessageHandler{}

		controller = NewController(
			"<name>",
			handler,
			&messageIDs,
			message.NewTypeSet(
				fixtures.MessageBType,
				fixtures.MessageEType,
			),
		)

		messageIDs.Reset() // reset after setup for a predictable ID.
	})

	Describe("func RecordEvent", func() {
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
			now := time.Now()
			_, err := controller.Handle(
				context.Background(),
				buf,
				now,
				command,
			)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(buf.Facts()).To(ContainElement(
				fact.EventRecordedByIntegration{
					HandlerName: "<name>",
					Handler:     handler,
					Envelope:    command,
					EventEnvelope: command.NewEvent(
						"1",
						fixtures.MessageB1,
						now,
						envelope.Origin{
							HandlerName: "<name>",
							HandlerType: handlerkit.IntegrationType,
						},
					),
				},
			))
		})

		It("panics if the event type is not configured to be produced", func() {
			handler.HandleCommandFunc = func(
				_ context.Context,
				s dogma.IntegrationCommandScope,
				m dogma.Message,
			) error {
				s.RecordEvent(fixtures.MessageZ1)
				return nil
			}

			Expect(func() {
				controller.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					command,
				)
			}).To(Panic())
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
			_, err := controller.Handle(
				context.Background(),
				buf,
				time.Now(),
				command,
			)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(buf.Facts()).To(ContainElement(
				fact.MessageLoggedByIntegration{
					HandlerName: "<name>",
					Handler:     handler,
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
