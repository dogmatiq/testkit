package engine_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/enginekit/config"
	"github.com/dogmatiq/enginekit/config/runtimeconfig"
	"github.com/dogmatiq/enginekit/enginetest"
	. "github.com/dogmatiq/enginekit/enginetest/stubs"
	. "github.com/dogmatiq/testkit/engine"
	"github.com/dogmatiq/testkit/envelope"
	"github.com/dogmatiq/testkit/fact"
	g "github.com/onsi/ginkgo/v2"
	gm "github.com/onsi/gomega"
)

func TestEngine(t *testing.T) {
	enginetest.RunTests(
		t,
		func(p enginetest.SetupParams) enginetest.SetupResult {
			e, err := New(runtimeconfig.FromApplication(p.App))
			if err != nil {
				t.Fatal(err)
			}

			return enginetest.SetupResult{
				RunEngine: func(ctx context.Context) error {
					return Run(ctx, e, 0)
				},
				Executor: &CommandExecutor{
					Engine: e,
					Options: []OperationOption{
						WithObserver(
							fact.NewLogger(
								func(s string) {
									t.Log(s)
								},
							),
						),
					},
				},
			}
		},
	)
}

var _ = g.Describe("type Engine", func() {
	var (
		aggregate            *AggregateMessageHandlerStub
		process              *ProcessMessageHandlerStub
		integration          *IntegrationMessageHandlerStub
		projection, disabled *ProjectionMessageHandlerStub
		app                  *ApplicationStub
		cfg                  *config.Application
		engine               *Engine
	)

	type (
		AggregateCommand   = CommandStub[TypeA]
		IntegrationCommand = CommandStub[TypeB]

		AggregateEvent   = EventStub[TypeA]
		IntegrationEvent = EventStub[TypeB]

		ForeignEventForProcess    = EventStub[TypeC]
		ForeignEventForProjection = EventStub[TypeD]
	)

	g.BeforeEach(func() {
		aggregate = &AggregateMessageHandlerStub{
			ConfigureFunc: func(c dogma.AggregateConfigurer) {
				c.Identity("<aggregate>", "c72c106b-771e-42f8-b3e6-05452d4002ed")
				c.Routes(
					dogma.HandlesCommand[AggregateCommand](),
					dogma.RecordsEvent[AggregateEvent](),
				)
			},
			RouteCommandToInstanceFunc: func(dogma.Command) string {
				return "<instance>"
			},
		}

		process = &ProcessMessageHandlerStub{
			ConfigureFunc: func(c dogma.ProcessConfigurer) {
				c.Identity("<process>", "4721492d-7fa3-4cfa-9f0f-a3cb1f95933e")
				c.Routes(
					dogma.HandlesEvent[ForeignEventForProcess](),
					dogma.HandlesEvent[AggregateEvent](), // shared with <projection>
					dogma.ExecutesCommand[IntegrationCommand](),
				)
			},
			RouteEventToInstanceFunc: func(context.Context, dogma.Event) (string, bool, error) {
				return "<instance>", true, nil
			},
		}

		integration = &IntegrationMessageHandlerStub{
			ConfigureFunc: func(c dogma.IntegrationConfigurer) {
				c.Identity("<integration>", "8b840c55-0b04-4107-bd4c-c69052c9fca3")
				c.Routes(
					dogma.HandlesCommand[IntegrationCommand](),
					dogma.RecordsEvent[IntegrationEvent](),
				)
			},
		}

		projection = &ProjectionMessageHandlerStub{
			ConfigureFunc: func(c dogma.ProjectionConfigurer) {
				c.Identity("<projection>", "f2b324d6-74f1-409e-8b28-8e44454037a9")
				c.Routes(
					dogma.HandlesEvent[ForeignEventForProjection](),
					dogma.HandlesEvent[AggregateEvent](), // shared with <process>
				)
			},
		}

		disabled = &ProjectionMessageHandlerStub{
			ConfigureFunc: func(c dogma.ProjectionConfigurer) {
				c.Identity("<disabled-projection>", "06426c1f-788d-4852-9a3f-c77580dafaed")
				c.Routes(
					dogma.HandlesEvent[ForeignEventForProjection](),
				)
				c.Disable()
			},
		}

		app = &ApplicationStub{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "9bc07eeb-5821-4649-941a-d931c8c88cb9")
				c.Routes(
					dogma.ViaAggregate(aggregate),
					dogma.ViaProcess(process),
					dogma.ViaIntegration(integration),
					dogma.ViaProjection(projection),
					dogma.ViaProjection(disabled),
				)
			},
		}

		cfg = runtimeconfig.FromApplication(app)
		engine = MustNew(cfg)
	})

	g.Describe("func Dispatch()", func() {
		g.It("allows dispatching commands", func() {
			called := false
			aggregate.HandleCommandFunc = func(
				dogma.AggregateRoot,
				dogma.AggregateCommandScope,
				dogma.Command,
			) {
				called = true
			}

			err := engine.Dispatch(
				context.Background(),
				AggregateCommand{},
			)
			gm.Expect(err).ShouldNot(gm.HaveOccurred())
			gm.Expect(called).To(gm.BeTrue())
		})

		g.It("allows dispatching events", func() {
			called := false
			projection.HandleEventFunc = func(
				context.Context,
				[]byte,
				[]byte,
				[]byte,
				dogma.ProjectionEventScope,
				dogma.Event,
			) (bool, error) {
				called = true
				return true, nil
			}

			err := engine.Dispatch(
				context.Background(),
				ForeignEventForProjection{},
			)
			gm.Expect(err).ShouldNot(gm.HaveOccurred())
			gm.Expect(called).To(gm.BeTrue())
		})

		g.It("skips handlers that are disabled by type", func() {
			aggregate.HandleCommandFunc = func(
				dogma.AggregateRoot,
				dogma.AggregateCommandScope,
				dogma.Command,
			) {
				g.Fail("unexpected call")
			}

			now := time.Now()
			buf := &fact.Buffer{}
			err := engine.Dispatch(
				context.Background(),
				AggregateCommand{},
				WithCurrentTime(now),
				WithObserver(buf),
				EnableAggregates(false),
			)
			gm.Expect(err).ShouldNot(gm.HaveOccurred())

			h, _ := cfg.HandlerByName("<aggregate>")
			gm.Expect(buf.Facts()).To(gm.ContainElement(
				fact.HandlingSkipped{
					Handler: h,
					Envelope: &envelope.Envelope{
						MessageID:     "1",
						CausationID:   "1",
						CorrelationID: "1",
						Message:       AggregateCommand{},
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
				dogma.Command,
			) {
				g.Fail("unexpected call")
			}

			now := time.Now()
			buf := &fact.Buffer{}
			err := engine.Dispatch(
				context.Background(),
				AggregateCommand{},
				WithCurrentTime(now),
				WithObserver(buf),
				EnableHandler("<aggregate>", false),
			)
			gm.Expect(err).ShouldNot(gm.HaveOccurred())

			h, _ := cfg.HandlerByName("<aggregate>")
			gm.Expect(buf.Facts()).To(gm.ContainElement(
				fact.HandlingSkipped{
					Handler: h,
					Envelope: &envelope.Envelope{
						MessageID:     "1",
						CausationID:   "1",
						CorrelationID: "1",
						Message:       AggregateCommand{},
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
				dogma.Command,
			) {
				called = true
			}

			err := engine.Dispatch(
				context.Background(),
				AggregateCommand{},
				EnableHandler("<aggregate>", false),
				EnableHandler("<aggregate>", true),
			)
			gm.Expect(err).ShouldNot(gm.HaveOccurred())
			gm.Expect(called).To(gm.BeTrue())
		})

		g.It("always returns context errors even if other errors occur", func() {
			ctx, cancel := context.WithCancel(context.Background())

			integration.HandleCommandFunc = func(
				context.Context,
				dogma.IntegrationCommandScope,
				dogma.Command,
			) error {
				cancel()
				return errors.New("<error>")
			}

			err := engine.Dispatch(ctx, IntegrationCommand{})
			gm.Expect(err).To(gm.Equal(context.Canceled))
		})

		g.It("adds handler details to controller errors", func() {
			integration.HandleCommandFunc = func(
				context.Context,
				dogma.IntegrationCommandScope,
				dogma.Command,
			) error {
				return errors.New("<error>")
			}

			err := engine.Dispatch(context.Background(), IntegrationCommand{})
			gm.Expect(err).To(gm.MatchError("<integration> integration: <error>"))
		})

		g.It("panics if the message is invalid", func() {
			gm.Expect(func() {
				engine.Dispatch(
					context.Background(),
					AggregateCommand{
						ValidationError: "<invalid>",
					},
				)
			}).To(gm.PanicWith("cannot dispatch invalid stubs.CommandStub[TypeA] message: <invalid>"))
		})

		g.It("panics if the message type is unrecognized", func() {
			gm.Expect(func() {
				engine.Dispatch(context.Background(), CommandX1)
			}).To(gm.PanicWith("the stubs.CommandStub[TypeX] message type is not consumed by any handlers"))
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
			gm.Expect(err).ShouldNot(gm.HaveOccurred())

			h, _ := cfg.HandlerByName("<aggregate>")
			gm.Expect(buf.Facts()).To(gm.ContainElement(
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
			gm.Expect(err).ShouldNot(gm.HaveOccurred())

			h, _ := cfg.HandlerByName("<aggregate>")
			gm.Expect(buf.Facts()).To(gm.ContainElement(
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
			gm.Expect(err).ShouldNot(gm.HaveOccurred())

			h, _ := cfg.HandlerByName("<aggregate>")
			gm.Expect(buf.Facts()).To(gm.ContainElement(
				fact.TickBegun{
					Handler: h,
				},
			))
		})

		g.It("skips handlers that are disabled by their configuration", func() {
			buf := &fact.Buffer{}
			err := engine.Tick(
				context.Background(),
				WithObserver(buf),
			)
			gm.Expect(err).ShouldNot(gm.HaveOccurred())

			h, _ := cfg.HandlerByName("<disabled-projection>")
			gm.Expect(buf.Facts()).To(gm.ContainElement(
				fact.TickSkipped{
					Handler: h,
					Reason:  fact.IndividualHandlerDisabledByConfiguration,
				},
			))
		})

		g.It("skips handlers that are disabled by their configuration, even if they are explicitly enabled by type", func() {
			buf := &fact.Buffer{}
			err := engine.Tick(
				context.Background(),
				WithObserver(buf),
				EnableProjections(true),
			)
			gm.Expect(err).ShouldNot(gm.HaveOccurred())

			h, _ := cfg.HandlerByName("<disabled-projection>")
			gm.Expect(buf.Facts()).To(gm.ContainElement(
				fact.TickSkipped{
					Handler: h,
					Reason:  fact.IndividualHandlerDisabledByConfiguration,
				},
			))
		})

		g.It("does not skip handlers that are enabled by name, even if they are disabled by their configuration", func() {
			buf := &fact.Buffer{}
			err := engine.Tick(
				context.Background(),
				WithObserver(buf),
				EnableHandler("<disabled-projection>", true),
			)
			gm.Expect(err).ShouldNot(gm.HaveOccurred())

			h, _ := cfg.HandlerByName("<disabled-projection>")
			gm.Expect(buf.Facts()).To(gm.ContainElement(
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
			gm.Expect(err).To(gm.Equal(context.Canceled))
		})

		g.It("adds handler details to controller errors", func() {
			projection.CompactFunc = func(
				context.Context,
				dogma.ProjectionCompactScope,
			) error {
				return errors.New("<error>")
			}

			err := engine.Tick(context.Background())
			gm.Expect(err).To(gm.MatchError("<projection> projection: <error>"))
		})
	})
})
