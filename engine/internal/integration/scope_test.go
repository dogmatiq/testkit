package integration_test

import (
	"context"
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

var _ = g.Describe("type scope", func() {
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
				c.Identity("<name>", "24ec3839-5d51-4904-9b45-34b5282e7f24")
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

	g.Describe("func RecordEvent()", func() {
		g.BeforeEach(func() {
			handler.HandleCommandFunc = func(
				_ context.Context,
				s dogma.IntegrationCommandScope,
				_ dogma.Command,
			) error {
				s.RecordEvent(EventA1)
				return nil
			}
		})

		g.It("records a fact", func() {
			buf := &fact.Buffer{}
			now := time.Now()
			_, err := ctrl.Handle(
				context.Background(),
				buf,
				now,
				command,
			)

			gm.Expect(err).ShouldNot(gm.HaveOccurred())
			gm.Expect(buf.Facts()).To(gm.ContainElement(
				fact.EventRecordedByIntegration{
					Handler:  config,
					Envelope: command,
					EventEnvelope: command.NewEvent(
						"1",
						EventA1,
						now,
						envelope.Origin{
							Handler:     config,
							HandlerType: configkit.IntegrationHandlerType,
						},
					),
				},
			))
		})

		g.It("panics if the event type is not configured to be produced", func() {
			handler.HandleCommandFunc = func(
				_ context.Context,
				s dogma.IntegrationCommandScope,
				m dogma.Command,
			) error {
				s.RecordEvent(EventX1)
				return nil
			}

			gm.Expect(func() {
				ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					command,
				)
			}).To(gm.PanicWith(
				MatchAllFields(
					Fields{
						"Handler":        gm.Equal(config),
						"Interface":      gm.Equal("IntegrationMessageHandler"),
						"Method":         gm.Equal("HandleCommand"),
						"Implementation": gm.Equal(config.Handler()),
						"Message":        gm.Equal(command.Message),
						"Description":    gm.Equal("recorded an event of type stubs.EventStub[TypeX], which is not produced by this handler"),
						"Location": MatchAllFields(
							Fields{
								"Func": gm.Not(gm.BeEmpty()),
								"File": gm.HaveSuffix("/engine/internal/integration/scope_test.go"),
								"Line": gm.Not(gm.BeZero()),
							},
						),
					},
				),
			))
		})

		g.It("panics if the event is invalid", func() {
			handler.HandleCommandFunc = func(
				_ context.Context,
				s dogma.IntegrationCommandScope,
				_ dogma.Command,
			) error {
				s.RecordEvent(EventStub[TypeA]{
					ValidationError: "<invalid>",
				})
				return nil
			}

			gm.Expect(func() {
				ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					command,
				)
			}).To(gm.PanicWith(
				MatchAllFields(
					Fields{
						"Handler":        gm.Equal(config),
						"Interface":      gm.Equal("IntegrationMessageHandler"),
						"Method":         gm.Equal("HandleCommand"),
						"Implementation": gm.Equal(config.Handler()),
						"Message":        gm.Equal(command.Message),
						"Description":    gm.Equal("recorded an invalid stubs.EventStub[TypeA] event: <invalid>"),
						"Location": MatchAllFields(
							Fields{
								"Func": gm.Not(gm.BeEmpty()),
								"File": gm.HaveSuffix("/engine/internal/integration/scope_test.go"),
								"Line": gm.Not(gm.BeZero()),
							},
						),
					},
				),
			))
		})
	})

	g.Describe("func Log()", func() {
		g.BeforeEach(func() {
			handler.HandleCommandFunc = func(
				_ context.Context,
				s dogma.IntegrationCommandScope,
				_ dogma.Command,
			) error {
				s.Log("<format>", "<arg-1>", "<arg-2>")
				return nil
			}
		})

		g.It("records a fact", func() {
			buf := &fact.Buffer{}
			_, err := ctrl.Handle(
				context.Background(),
				buf,
				time.Now(),
				command,
			)

			gm.Expect(err).ShouldNot(gm.HaveOccurred())
			gm.Expect(buf.Facts()).To(gm.ContainElement(
				fact.MessageLoggedByIntegration{
					Handler:   config,
					Envelope:  command,
					LogFormat: "<format>",
					LogArguments: []any{
						"<arg-1>",
						"<arg-2>",
					},
				},
			))
		})
	})
})
