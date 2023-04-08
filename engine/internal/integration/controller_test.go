package integration_test

import (
	"context"
	"errors"
	"time"

	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogma/fixtures"
	. "github.com/dogmatiq/testkit/engine/internal/integration"
	"github.com/dogmatiq/testkit/envelope"
	"github.com/dogmatiq/testkit/fact"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("type Controller", func() {
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
				c.Identity("<name>", "8cbb8bca-b5eb-4c94-a877-dfc8dc9968ca")
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

	Describe("func HandlerConfig()", func() {
		It("returns the handler config", func() {
			Expect(ctrl.HandlerConfig()).To(BeIdenticalTo(config))
		})
	})

	Describe("func Tick()", func() {
		It("does not return any envelopes", func() {
			envelopes, err := ctrl.Tick(
				context.Background(),
				fact.Ignore,
				time.Now(),
			)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(envelopes).To(BeEmpty())
		})

		It("does not record any facts", func() {
			buf := &fact.Buffer{}
			_, err := ctrl.Tick(
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
				Expect(m).To(Equal(MessageA1))
				return nil
			}

			_, err := ctrl.Handle(
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
				s.RecordEvent(MessageE1)
				s.RecordEvent(MessageE2)
				return nil
			}

			now := time.Now()
			events, err := ctrl.Handle(
				context.Background(),
				fact.Ignore,
				now,
				command,
			)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(events).To(ConsistOf(
				command.NewEvent(
					"1",
					MessageE1,
					now,
					envelope.Origin{
						Handler:     config,
						HandlerType: configkit.IntegrationHandlerType,
					},
				),
				command.NewEvent(
					"2",
					MessageE2,
					now,
					envelope.Origin{
						Handler:     config,
						HandlerType: configkit.IntegrationHandlerType,
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

			_, err := ctrl.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				command,
			)

			Expect(err).To(Equal(expected))
		})

		It("uses the handler's timeout hint", func() {
			hint := 3 * time.Second
			handler.TimeoutHintFunc = func(dogma.Message) time.Duration {
				return hint
			}

			handler.HandleCommandFunc = func(
				ctx context.Context,
				_ dogma.IntegrationCommandScope,
				_ dogma.Message,
			) error {
				dl, ok := ctx.Deadline()
				Expect(ok).To(BeTrue())
				Expect(dl).To(BeTemporally("~", time.Now().Add(hint)))
				return nil
			}

			_, err := ctrl.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				command,
			)

			Expect(err).ShouldNot(HaveOccurred())
		})

		It("provides more context to UnexpectedMessage panics from HandleCommand()", func() {
			handler.HandleCommandFunc = func(
				context.Context,
				dogma.IntegrationCommandScope,
				dogma.Message,
			) error {
				panic(dogma.UnexpectedMessage)
			}

			Expect(func() {
				ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					command,
				)
			}).To(PanicWith(
				MatchFields(
					IgnoreExtras,
					Fields{
						"Handler":   Equal(config),
						"Interface": Equal("IntegrationMessageHandler"),
						"Method":    Equal("HandleCommand"),
						"Message":   Equal(command.Message),
					},
				),
			))
		})

		It("provides more context to UnexpectedMessage panics from TimeoutHint()", func() {
			handler.TimeoutHintFunc = func(
				dogma.Message,
			) time.Duration {
				panic(dogma.UnexpectedMessage)
			}

			Expect(func() {
				ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					command,
				)
			}).To(PanicWith(
				MatchFields(
					IgnoreExtras,
					Fields{
						"Handler":   Equal(config),
						"Interface": Equal("IntegrationMessageHandler"),
						"Method":    Equal("TimeoutHint"),
						"Message":   Equal(command.Message),
					},
				),
			))
		})
	})

	Describe("func Reset()", func() {
		It("does nothing", func() {
			ctrl.Reset()
		})
	})
})
