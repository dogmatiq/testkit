package engine_test

import (
	"context"
	"errors"
	"time"

	"github.com/dogmatiq/configkit"
	. "github.com/dogmatiq/configkit/fixtures"
	"github.com/dogmatiq/configkit/message"
	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogma/fixtures"
	. "github.com/dogmatiq/testkit/engine"
	"github.com/dogmatiq/testkit/envelope"
	"github.com/dogmatiq/testkit/fact"
	g "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = g.Describe("type Engine", func() {
	var (
		aggregate   *AggregateMessageHandler
		process     *ProcessMessageHandler
		integration *IntegrationMessageHandler
		projection  *ProjectionMessageHandler
		app         *Application
		config      configkit.RichApplication
		engine      *Engine
	)

	g.BeforeEach(func() {
		aggregate = &AggregateMessageHandler{
			ConfigureFunc: func(c dogma.AggregateConfigurer) {
				c.Identity("<aggregate>", "c72c106b-771e-42f8-b3e6-05452d4002ed")
				c.ConsumesCommandType(MessageA{})
				c.ProducesEventType(MessageE{})
			},
			RouteCommandToInstanceFunc: func(dogma.Message) string {
				return "<instance>"
			},
		}

		process = &ProcessMessageHandler{
			ConfigureFunc: func(c dogma.ProcessConfigurer) {
				c.Identity("<process>", "4721492d-7fa3-4cfa-9f0f-a3cb1f95933e")
				c.ConsumesEventType(MessageB{})
				c.ConsumesEventType(MessageE{}) // shared with <projection>
				c.ProducesCommandType(MessageC{})
			},
			RouteEventToInstanceFunc: func(context.Context, dogma.Message) (string, bool, error) {
				return "<instance>", true, nil
			},
		}

		integration = &IntegrationMessageHandler{
			ConfigureFunc: func(c dogma.IntegrationConfigurer) {
				c.Identity("<integration>", "8b840c55-0b04-4107-bd4c-c69052c9fca3")
				c.ConsumesCommandType(MessageC{})
				c.ProducesEventType(MessageF{})
			},
		}

		projection = &ProjectionMessageHandler{
			ConfigureFunc: func(c dogma.ProjectionConfigurer) {
				c.Identity("<projection>", "f2b324d6-74f1-409e-8b28-8e44454037a9")
				c.ConsumesEventType(MessageD{})
				c.ConsumesEventType(MessageE{}) // shared with <process>
			},
		}

		app = &Application{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "9bc07eeb-5821-4649-941a-d931c8c88cb9")
				c.RegisterAggregate(aggregate)
				c.RegisterProcess(process)
				c.RegisterIntegration(integration)
				c.RegisterProjection(projection)
			},
		}

		config = configkit.FromApplication(app)
		engine = MustNew(config)
	})

	g.Describe("func Dispatch()", func() {
		g.It("skips handlers that are disabled by type", func() {
			aggregate.HandleCommandFunc = func(
				dogma.AggregateRoot,
				dogma.AggregateCommandScope,
				dogma.Message,
			) {
				g.Fail("unexpected call")
			}

			now := time.Now()
			buf := &fact.Buffer{}
			err := engine.Dispatch(
				context.Background(),
				MessageA1,
				WithCurrentTime(now),
				WithObserver(buf),
				EnableAggregates(false),
			)
			Expect(err).ShouldNot(HaveOccurred())

			h, _ := config.RichHandlers().ByName("<aggregate>")
			Expect(buf.Facts()).To(ContainElement(
				fact.HandlingSkipped{
					Handler: h,
					Envelope: &envelope.Envelope{
						MessageID:     "1",
						CausationID:   "1",
						CorrelationID: "1",
						Message:       MessageA1,
						Type:          MessageAType,
						Role:          message.CommandRole,
						CreatedAt:     now,
					},
					Reason: fact.HandlerTypeDisabled,
				},
			))
		})

		g.It("skips handlers that are disabled by name", func() {
			aggregate.HandleCommandFunc = func(
				dogma.AggregateRoot,
				dogma.AggregateCommandScope,
				dogma.Message,
			) {
				g.Fail("unexpected call")
			}

			now := time.Now()
			buf := &fact.Buffer{}
			err := engine.Dispatch(
				context.Background(),
				MessageA1,
				WithCurrentTime(now),
				WithObserver(buf),
				EnableHandler("<aggregate>", false),
			)
			Expect(err).ShouldNot(HaveOccurred())

			h, _ := config.RichHandlers().ByName("<aggregate>")
			Expect(buf.Facts()).To(ContainElement(
				fact.HandlingSkipped{
					Handler: h,
					Envelope: &envelope.Envelope{
						MessageID:     "1",
						CausationID:   "1",
						CorrelationID: "1",
						Message:       MessageA1,
						Type:          MessageAType,
						Role:          message.CommandRole,
						CreatedAt:     now,
					},
					Reason: fact.IndividualHandlerDisabled,
				},
			))
		})

		g.It("does not skip handlers that are enabled by name", func() {
			called := false
			aggregate.HandleCommandFunc = func(
				dogma.AggregateRoot,
				dogma.AggregateCommandScope,
				dogma.Message,
			) {
				called = true
			}

			err := engine.Dispatch(
				context.Background(),
				MessageA1,
				EnableHandler("<aggregate>", false),
				EnableHandler("<aggregate>", true),
			)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(called).To(BeTrue())
		})

		g.It("always returns context errors even if other errors occur", func() {
			ctx, cancel := context.WithCancel(context.Background())

			integration.HandleCommandFunc = func(
				context.Context,
				dogma.IntegrationCommandScope,
				dogma.Message,
			) error {
				cancel()
				return errors.New("<error>")
			}

			err := engine.Dispatch(ctx, MessageC1)
			Expect(err).To(Equal(context.Canceled))
		})

		g.It("adds handler details to controller errors", func() {
			integration.HandleCommandFunc = func(
				context.Context,
				dogma.IntegrationCommandScope,
				dogma.Message,
			) error {
				return errors.New("<error>")
			}

			err := engine.Dispatch(context.Background(), MessageC1)
			Expect(err).To(MatchError("<integration> integration: <error>"))
		})

		g.It("panics if the message is invalid", func() {
			Expect(func() {
				engine.Dispatch(
					context.Background(),
					MessageA{
						Value: errors.New("<invalid>"),
					},
				)
			}).To(PanicWith("cannot dispatch invalid fixtures.MessageA message: <invalid>"))
		})

		g.It("panics if the message type is unrecognized", func() {
			Expect(func() {
				engine.Dispatch(context.Background(), MessageX1)
			}).To(PanicWith("the fixtures.MessageX message type is not consumed by any handlers"))
		})
	})

	g.Describe("func Tick()", func() {
		g.It("skips handlers that are disabled by type", func() {
			buf := &fact.Buffer{}
			err := engine.Tick(
				context.Background(),
				WithObserver(buf),
				EnableAggregates(false),
			)
			Expect(err).ShouldNot(HaveOccurred())

			h, _ := config.RichHandlers().ByName("<aggregate>")
			Expect(buf.Facts()).To(ContainElement(
				fact.TickSkipped{
					Handler: h,
					Reason:  fact.HandlerTypeDisabled,
				},
			))
		})

		g.It("skips handlers that are disabled by name", func() {
			buf := &fact.Buffer{}
			err := engine.Tick(
				context.Background(),
				WithObserver(buf),
				EnableHandler("<aggregate>", false),
			)
			Expect(err).ShouldNot(HaveOccurred())

			h, _ := config.RichHandlers().ByName("<aggregate>")
			Expect(buf.Facts()).To(ContainElement(
				fact.TickSkipped{
					Handler: h,
					Reason:  fact.IndividualHandlerDisabled,
				},
			))
		})

		g.It("does not skip handlers that are enabled by name", func() {
			buf := &fact.Buffer{}
			err := engine.Tick(
				context.Background(),
				WithObserver(buf),
				EnableHandler("<aggregate>", false),
				EnableHandler("<aggregate>", true),
			)
			Expect(err).ShouldNot(HaveOccurred())

			h, _ := config.RichHandlers().ByName("<aggregate>")
			Expect(buf.Facts()).To(ContainElement(
				fact.TickBegun{
					Handler: h,
				},
			))
		})

		g.It("always returns context errors even if other errors occur", func() {
			ctx, cancel := context.WithCancel(context.Background())

			projection.CompactFunc = func(
				context.Context,
				dogma.ProjectionCompactScope,
			) error {
				cancel()
				return errors.New("<error>")
			}

			err := engine.Tick(ctx)
			Expect(err).To(Equal(context.Canceled))
		})

		g.It("adds handler details to controller errors", func() {
			projection.CompactFunc = func(
				context.Context,
				dogma.ProjectionCompactScope,
			) error {
				return errors.New("<error>")
			}

			err := engine.Tick(context.Background())
			Expect(err).To(MatchError("<projection> projection: <error>"))
		})
	})
})
