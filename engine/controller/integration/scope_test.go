package integration_test

import (
	"context"
	"errors"
	"time"

	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogma/fixtures"
	. "github.com/dogmatiq/testkit/engine/controller/integration"
	"github.com/dogmatiq/testkit/engine/envelope"
	"github.com/dogmatiq/testkit/engine/fact"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("type scope", func() {
	var (
		messageIDs envelope.MessageIDGenerator
		handler    *IntegrationMessageHandler
		config     configkit.RichIntegration
		ctrl       *Controller
		command    *envelope.Envelope
	)

	BeforeEach(func() {
		command = envelope.NewCommand(
			"1000",
			MessageA1,
			time.Now(),
		)

		handler = &IntegrationMessageHandler{
			ConfigureFunc: func(c dogma.IntegrationConfigurer) {
				c.Identity("<name>", "<key>")
				c.ConsumesCommandType(MessageC{})
				c.ProducesEventType(MessageE{})
			},
		}

		config = configkit.FromIntegration(handler)

		ctrl = &Controller{
			Config:     config,
			MessageIDs: &messageIDs,
		}

		messageIDs.Reset() // reset after setup for a predictable ID.
	})

	Describe("func RecordEvent()", func() {
		BeforeEach(func() {
			handler.HandleCommandFunc = func(
				_ context.Context,
				s dogma.IntegrationCommandScope,
				_ dogma.Message,
			) error {
				s.RecordEvent(MessageE1)
				return nil
			}
		})

		It("records a fact", func() {
			buf := &fact.Buffer{}
			now := time.Now()
			_, err := ctrl.Handle(
				context.Background(),
				buf,
				now,
				command,
			)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(buf.Facts()).To(ContainElement(
				fact.EventRecordedByIntegration{
					Handler:  config,
					Envelope: command,
					EventEnvelope: command.NewEvent(
						"1",
						MessageE1,
						now,
						envelope.Origin{
							Handler:     config,
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
				s.RecordEvent(MessageX1)
				return nil
			}

			Expect(func() {
				ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					command,
				)
			}).To(PanicWith(
				MatchAllFields(
					Fields{
						"Handler":        Equal(config),
						"Interface":      Equal("IntegrationMessageHandler"),
						"Method":         Equal("HandleCommand"),
						"Implementation": Equal(config.Handler()),
						"Message":        Equal(command.Message),
						"Description":    Equal("recorded an event of type fixtures.MessageX, which is not produced by this handler"),
						"Location": MatchAllFields(
							Fields{
								"Func": Not(BeEmpty()),
								"File": HaveSuffix("/engine/controller/integration/scope_test.go"),
								"Line": Not(BeZero()),
							},
						),
					},
				),
			))
		})

		It("panics if the event is invalid", func() {
			handler.HandleCommandFunc = func(
				_ context.Context,
				s dogma.IntegrationCommandScope,
				_ dogma.Message,
			) error {
				s.RecordEvent(MessageE{
					Value: errors.New("<invalid>"),
				})
				return nil
			}

			Expect(func() {
				ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					command,
				)
			}).To(PanicWith(
				MatchAllFields(
					Fields{
						"Handler":        Equal(config),
						"Interface":      Equal("IntegrationMessageHandler"),
						"Method":         Equal("HandleCommand"),
						"Implementation": Equal(config.Handler()),
						"Message":        Equal(command.Message),
						"Description":    Equal("recorded an invalid fixtures.MessageE event: <invalid>"),
						"Location": MatchAllFields(
							Fields{
								"Func": Not(BeEmpty()),
								"File": HaveSuffix("/engine/controller/integration/scope_test.go"),
								"Line": Not(BeZero()),
							},
						),
					},
				),
			))
		})
	})

	Describe("func Log()", func() {
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
			_, err := ctrl.Handle(
				context.Background(),
				buf,
				time.Now(),
				command,
			)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(buf.Facts()).To(ContainElement(
				fact.MessageLoggedByIntegration{
					Handler:   config,
					Envelope:  command,
					LogFormat: "<format>",
					LogArguments: []interface{}{
						"<arg-1>",
						"<arg-2>",
					},
				},
			))
		})
	})
})
