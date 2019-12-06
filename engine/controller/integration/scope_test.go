package integration_test

import (
	"context"
	"time"

	"github.com/dogmatiq/configkit"
	. "github.com/dogmatiq/configkit/fixtures"
	"github.com/dogmatiq/configkit/message"
	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogma/fixtures"
	"github.com/dogmatiq/enginekit/identity"
	. "github.com/dogmatiq/testkit/engine/controller/integration"
	"github.com/dogmatiq/testkit/engine/envelope"
	"github.com/dogmatiq/testkit/engine/fact"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("type scope", func() {
	var (
		messageIDs envelope.MessageIDGenerator
		handler    *IntegrationMessageHandler
		controller *Controller
		command    *envelope.Envelope
	)

	BeforeEach(func() {
		command = envelope.NewCommand(
			"1000",
			MessageA1,
			time.Now(),
		)

		handler = &IntegrationMessageHandler{}

		controller = NewController(
			identity.MustNew("<name>", "<key>"),
			handler,
			&messageIDs,
			message.NewTypeSet(
				MessageBType,
				MessageEType,
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
				s.RecordEvent(MessageB1)
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
						MessageB1,
						now,
						envelope.Origin{
							HandlerName: "<name>",
							HandlerType: configkit.IntegrationHandlerType,
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
				s.RecordEvent(MessageZ1)
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
