package integration_test

import (
	"context"
	"errors"
	"time"

	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/enginekit/enginetest/stubs"
	. "github.com/dogmatiq/testkit/engine/internal/integration"
	"github.com/dogmatiq/testkit/envelope"
	"github.com/dogmatiq/testkit/fact"
	g "github.com/onsi/ginkgo/v2"
	gm "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = g.Describe("type Controller", func() {
	var (
		messageIDs envelope.MessageIDGenerator
		handler    *IntegrationMessageHandlerStub
		config     configkit.RichIntegration
		ctrl       *Controller
		command    *envelope.Envelope
	)

	g.BeforeEach(func() {
		command = envelope.NewCommand(
			"1000",
			CommandA1,
			time.Now(),
		)

		handler = &IntegrationMessageHandlerStub{
			ConfigureFunc: func(c dogma.IntegrationConfigurer) {
				c.Identity("<name>", "8cbb8bca-b5eb-4c94-a877-dfc8dc9968ca")
				c.Routes(
					dogma.HandlesCommand[CommandStub[TypeA]](),
					dogma.RecordsEvent[EventStub[TypeA]](),
				)
			},
		}

		config = configkit.FromIntegration(handler)

		ctrl = &Controller{
			Config:     config,
			MessageIDs: &messageIDs,
		}

		messageIDs.Reset() // reset after setup for a predictable ID.
	})

	g.Describe("func HandlerConfig()", func() {
		g.It("returns the handler config", func() {
			gm.Expect(ctrl.HandlerConfig()).To(gm.BeIdenticalTo(config))
		})
	})

	g.Describe("func Tick()", func() {
		g.It("does not return any envelopes", func() {
			envelopes, err := ctrl.Tick(
				context.Background(),
				fact.Ignore,
				time.Now(),
			)
			gm.Expect(err).ShouldNot(gm.HaveOccurred())
			gm.Expect(envelopes).To(gm.BeEmpty())
		})

		g.It("does not record any facts", func() {
			buf := &fact.Buffer{}
			_, err := ctrl.Tick(
				context.Background(),
				buf,
				time.Now(),
			)
			gm.Expect(err).ShouldNot(gm.HaveOccurred())
			gm.Expect(buf.Facts()).To(gm.BeEmpty())
		})
	})

	g.Describe("func Handle()", func() {
		g.It("forwards the message to the handler", func() {
			called := false
			handler.HandleCommandFunc = func(
				_ context.Context,
				_ dogma.IntegrationCommandScope,
				m dogma.Command,
			) error {
				called = true
				gm.Expect(m).To(gm.Equal(CommandA1))
				return nil
			}

			_, err := ctrl.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				command,
			)

			gm.Expect(err).ShouldNot(gm.HaveOccurred())
			gm.Expect(called).To(gm.BeTrue())
		})

		g.It("returns the recorded events", func() {
			handler.HandleCommandFunc = func(
				_ context.Context,
				s dogma.IntegrationCommandScope,
				_ dogma.Command,
			) error {
				s.RecordEvent(EventA1)
				s.RecordEvent(EventA2)
				return nil
			}

			now := time.Now()
			events, err := ctrl.Handle(
				context.Background(),
				fact.Ignore,
				now,
				command,
			)

			gm.Expect(err).ShouldNot(gm.HaveOccurred())
			gm.Expect(events).To(gm.ConsistOf(
				command.NewEvent(
					"1",
					EventA1,
					now,
					envelope.Origin{
						Handler:     config,
						HandlerType: configkit.IntegrationHandlerType,
					},
				),
				command.NewEvent(
					"2",
					EventA2,
					now,
					envelope.Origin{
						Handler:     config,
						HandlerType: configkit.IntegrationHandlerType,
					},
				),
			))
		})

		g.It("propagates handler errors", func() {
			expected := errors.New("<error>")

			handler.HandleCommandFunc = func(
				_ context.Context,
				_ dogma.IntegrationCommandScope,
				_ dogma.Command,
			) error {
				return expected
			}

			_, err := ctrl.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				command,
			)

			gm.Expect(err).To(gm.Equal(expected))
		})

		g.It("provides more context to UnexpectedMessage panics from HandleCommand()", func() {
			handler.HandleCommandFunc = func(
				context.Context,
				dogma.IntegrationCommandScope,
				dogma.Command,
			) error {
				panic(dogma.UnexpectedMessage)
			}

			gm.Expect(func() {
				ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					command,
				)
			}).To(gm.PanicWith(
				MatchFields(
					IgnoreExtras,
					Fields{
						"Handler":   gm.Equal(config),
						"Interface": gm.Equal("IntegrationMessageHandler"),
						"Method":    gm.Equal("HandleCommand"),
						"Message":   gm.Equal(command.Message),
					},
				),
			))
		})
	})

	g.Describe("func Reset()", func() {
		g.It("does nothing", func() {
			ctrl.Reset()
		})
	})
})
