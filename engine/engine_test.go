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
	"github.com/dogmatiq/testkit/internal/x/xtesting"
)

type (
	engineAggregateCommand          = CommandStub[TypeA]
	engineIntegrationCommand        = CommandStub[TypeB]
	engineAggregateEvent            = EventStub[TypeA]
	engineForeignEventForProcess    = EventStub[TypeC]
	engineForeignEventForProjection = EventStub[TypeD]
)

type engineFixture struct {
	aggregate   *AggregateMessageHandlerStub
	process     *ProcessMessageHandlerStub
	integration *IntegrationMessageHandlerStub
	projection  *ProjectionMessageHandlerStub
	disabled    *ProjectionMessageHandlerStub
	cfg         *config.Application
	engine      *Engine
}

func newEngineFixture() *engineFixture {
	fx := &engineFixture{}

	fx.aggregate = &AggregateMessageHandlerStub{
		ConfigureFunc: func(c dogma.AggregateConfigurer) {
			c.Identity("<aggregate>", "c72c106b-771e-42f8-b3e6-05452d4002ed")
			c.Routes(
				dogma.HandlesCommand[*engineAggregateCommand](),
				dogma.RecordsEvent[*engineAggregateEvent](),
			)
		},
		RouteCommandToInstanceFunc: func(dogma.Command) string {
			return "<instance>"
		},
	}

	fx.process = &ProcessMessageHandlerStub{
		ConfigureFunc: func(c dogma.ProcessConfigurer) {
			c.Identity("<process>", "4721492d-7fa3-4cfa-9f0f-a3cb1f95933e")
			c.Routes(
				dogma.HandlesEvent[*engineForeignEventForProcess](),
				dogma.HandlesEvent[*engineAggregateEvent](),
				dogma.ExecutesCommand[*engineIntegrationCommand](),
			)
		},
		RouteEventToInstanceFunc: func(context.Context, dogma.Event) (string, bool, error) {
			return "<instance>", true, nil
		},
	}

	fx.integration = &IntegrationMessageHandlerStub{
		ConfigureFunc: func(c dogma.IntegrationConfigurer) {
			c.Identity("<integration>", "8b840c55-0b04-4107-bd4c-c69052c9fca3")
			c.Routes(
				dogma.HandlesCommand[*engineIntegrationCommand](),
				dogma.RecordsEvent[*EventStub[TypeB]](),
			)
		},
	}

	fx.projection = &ProjectionMessageHandlerStub{
		ConfigureFunc: func(c dogma.ProjectionConfigurer) {
			c.Identity("<projection>", "f2b324d6-74f1-409e-8b28-8e44454037a9")
			c.Routes(
				dogma.HandlesEvent[*engineForeignEventForProjection](),
				dogma.HandlesEvent[*engineAggregateEvent](),
			)
		},
	}

	fx.disabled = &ProjectionMessageHandlerStub{
		ConfigureFunc: func(c dogma.ProjectionConfigurer) {
			c.Identity("<disabled-projection>", "06426c1f-788d-4852-9a3f-c77580dafaed")
			c.Routes(
				dogma.HandlesEvent[*engineForeignEventForProjection](),
			)
			c.Disable()
		},
	}

	app := &ApplicationStub{
		ConfigureFunc: func(c dogma.ApplicationConfigurer) {
			c.Identity("<app>", "9bc07eeb-5821-4649-941a-d931c8c88cb9")
			c.Routes(
				dogma.ViaAggregate(fx.aggregate),
				dogma.ViaProcess(fx.process),
				dogma.ViaIntegration(fx.integration),
				dogma.ViaProjection(fx.projection),
				dogma.ViaProjection(fx.disabled),
			)
		},
	}

	fx.cfg = runtimeconfig.FromApplication(app)
	fx.engine = MustNew(fx.cfg)

	return fx
}

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

func TestEngine_Dispatch(t *testing.T) {
	t.Run("it allows dispatching commands", func(t *testing.T) {
		fx := newEngineFixture()
		called := false
		fx.aggregate.HandleCommandFunc = func(dogma.AggregateRoot, dogma.AggregateCommandScope, dogma.Command) {
			called = true
		}

		err := fx.engine.Dispatch(context.Background(), &engineAggregateCommand{})
		if err != nil {
			t.Fatal(err)
		}
		if !called {
			t.Fatal("expected HandleCommand to be called")
		}
	})

	t.Run("it allows dispatching events", func(t *testing.T) {
		fx := newEngineFixture()
		called := false
		fx.projection.HandleEventFunc = func(_ context.Context, s dogma.ProjectionEventScope, _ dogma.Event) (uint64, error) {
			called = true
			return s.Offset() + 1, nil
		}

		err := fx.engine.Dispatch(context.Background(), &engineForeignEventForProjection{})
		if err != nil {
			t.Fatal(err)
		}
		if !called {
			t.Fatal("expected HandleEvent to be called")
		}
	})

	t.Run("it skips handlers that are disabled by type", func(t *testing.T) {
		fx := newEngineFixture()
		fx.aggregate.HandleCommandFunc = func(dogma.AggregateRoot, dogma.AggregateCommandScope, dogma.Command) {
			t.Fatal("unexpected call")
		}

		now := time.Now()
		buf := &fact.Buffer{}
		err := fx.engine.Dispatch(
			context.Background(),
			&engineAggregateCommand{},
			WithCurrentTime(now),
			WithObserver(buf),
			EnableAggregates(false),
		)
		if err != nil {
			t.Fatal(err)
		}

		h, _ := fx.cfg.HandlerByName("<aggregate>")
		xtesting.ExpectContains[fact.Fact](
			t,

			"expected HandlingSkipped fact",
			buf.Facts(),
			fact.HandlingSkipped{
				Handler: h,
				Envelope: &envelope.Envelope{
					MessageID:     "1",
					CausationID:   "1",
					CorrelationID: "1",
					Message:       &engineAggregateCommand{},
					CreatedAt:     now,
				},
				Reason: fact.HandlerTypeDisabled,
			},
		)
	})

	t.Run("it skips handlers that are disabled by name", func(t *testing.T) {
		fx := newEngineFixture()
		fx.aggregate.HandleCommandFunc = func(dogma.AggregateRoot, dogma.AggregateCommandScope, dogma.Command) {
			t.Fatal("unexpected call")
		}

		now := time.Now()
		buf := &fact.Buffer{}
		err := fx.engine.Dispatch(
			context.Background(),
			&engineAggregateCommand{},
			WithCurrentTime(now),
			WithObserver(buf),
			EnableHandler("<aggregate>", false),
		)
		if err != nil {
			t.Fatal(err)
		}

		h, _ := fx.cfg.HandlerByName("<aggregate>")
		xtesting.ExpectContains[fact.Fact](
			t,

			"expected HandlingSkipped fact",
			buf.Facts(),
			fact.HandlingSkipped{
				Handler: h,
				Envelope: &envelope.Envelope{
					MessageID:     "1",
					CausationID:   "1",
					CorrelationID: "1",
					Message:       &engineAggregateCommand{},
					CreatedAt:     now,
				},
				Reason: fact.IndividualHandlerDisabled,
			},
		)
	})

	t.Run("it does not skip handlers that are enabled by name", func(t *testing.T) {
		fx := newEngineFixture()
		called := false
		fx.aggregate.HandleCommandFunc = func(dogma.AggregateRoot, dogma.AggregateCommandScope, dogma.Command) {
			called = true
		}

		err := fx.engine.Dispatch(
			context.Background(),
			&engineAggregateCommand{},
			EnableHandler("<aggregate>", false),
			EnableHandler("<aggregate>", true),
		)
		if err != nil {
			t.Fatal(err)
		}
		if !called {
			t.Fatal("expected HandleCommand to be called")
		}
	})

	t.Run("it always returns context errors even if other errors occur", func(t *testing.T) {
		fx := newEngineFixture()
		ctx, cancel := context.WithCancel(context.Background())

		fx.integration.HandleCommandFunc = func(context.Context, dogma.IntegrationCommandScope, dogma.Command) error {
			cancel()
			return errors.New("<error>")
		}

		err := fx.engine.Dispatch(ctx, &engineIntegrationCommand{})
		xtesting.Expect(t, "unexpected error", err, context.Canceled)
	})

	t.Run("it adds handler details to controller errors", func(t *testing.T) {
		fx := newEngineFixture()
		fx.integration.HandleCommandFunc = func(context.Context, dogma.IntegrationCommandScope, dogma.Command) error {
			return errors.New("<error>")
		}

		err := fx.engine.Dispatch(context.Background(), &engineIntegrationCommand{})
		if err == nil || err.Error() != "<integration> integration: <error>" {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("it panics if the message is invalid", func(t *testing.T) {
		fx := newEngineFixture()
		xtesting.ExpectPanic(
			t,
			"cannot dispatch invalid *stubs.CommandStub[TypeA] message: <invalid>",
			func() {
				fx.engine.Dispatch( //nolint:errcheck
					context.Background(),
					&engineAggregateCommand{ValidationError: "<invalid>"},
				)
			},
		)
	})

	t.Run("it panics if the message type is unrecognized", func(t *testing.T) {
		fx := newEngineFixture()
		xtesting.ExpectPanic(
			t,
			"the *stubs.CommandStub[TypeX] message type is not consumed by any handlers",
			func() {
				fx.engine.Dispatch(context.Background(), CommandX1) //nolint:errcheck
			},
		)
	})
}

func TestEngine_Tick(t *testing.T) {
	t.Run("it skips handlers that are disabled by type", func(t *testing.T) {
		fx := newEngineFixture()
		buf := &fact.Buffer{}
		err := fx.engine.Tick(
			context.Background(),
			WithObserver(buf),
			EnableAggregates(false),
		)
		if err != nil {
			t.Fatal(err)
		}

		h, _ := fx.cfg.HandlerByName("<aggregate>")
		xtesting.ExpectContains[fact.Fact](
			t,

			"expected TickSkipped fact",
			buf.Facts(),
			fact.TickSkipped{
				Handler: h,
				Reason:  fact.HandlerTypeDisabled,
			},
		)
	})

	t.Run("it skips handlers that are disabled by name", func(t *testing.T) {
		fx := newEngineFixture()
		buf := &fact.Buffer{}
		err := fx.engine.Tick(
			context.Background(),
			WithObserver(buf),
			EnableHandler("<aggregate>", false),
		)
		if err != nil {
			t.Fatal(err)
		}

		h, _ := fx.cfg.HandlerByName("<aggregate>")
		xtesting.ExpectContains[fact.Fact](
			t,

			"expected TickSkipped fact",
			buf.Facts(),
			fact.TickSkipped{
				Handler: h,
				Reason:  fact.IndividualHandlerDisabled,
			},
		)
	})

	t.Run("it does not skip handlers that are enabled by name", func(t *testing.T) {
		fx := newEngineFixture()
		buf := &fact.Buffer{}
		err := fx.engine.Tick(
			context.Background(),
			WithObserver(buf),
			EnableHandler("<aggregate>", false),
			EnableHandler("<aggregate>", true),
		)
		if err != nil {
			t.Fatal(err)
		}

		h, _ := fx.cfg.HandlerByName("<aggregate>")
		xtesting.ExpectContains[fact.Fact](
			t,

			"expected TickBegun fact",
			buf.Facts(),
			fact.TickBegun{
				Handler: h,
			},
		)
	})

	t.Run("it skips handlers that are disabled by their configuration", func(t *testing.T) {
		fx := newEngineFixture()
		buf := &fact.Buffer{}
		err := fx.engine.Tick(
			context.Background(),
			WithObserver(buf),
		)
		if err != nil {
			t.Fatal(err)
		}

		h, _ := fx.cfg.HandlerByName("<disabled-projection>")
		xtesting.ExpectContains[fact.Fact](
			t,

			"expected TickSkipped fact",
			buf.Facts(),
			fact.TickSkipped{
				Handler: h,
				Reason:  fact.IndividualHandlerDisabledByConfiguration,
			},
		)
	})

	t.Run("it skips handlers that are disabled by their configuration, even if they are explicitly enabled by type", func(t *testing.T) {
		fx := newEngineFixture()
		buf := &fact.Buffer{}
		err := fx.engine.Tick(
			context.Background(),
			WithObserver(buf),
			EnableProjections(true),
		)
		if err != nil {
			t.Fatal(err)
		}

		h, _ := fx.cfg.HandlerByName("<disabled-projection>")
		xtesting.ExpectContains[fact.Fact](
			t,

			"expected TickSkipped fact",
			buf.Facts(),
			fact.TickSkipped{
				Handler: h,
				Reason:  fact.IndividualHandlerDisabledByConfiguration,
			},
		)
	})

	t.Run("it does not skip handlers that are enabled by name, even if they are disabled by their configuration", func(t *testing.T) {
		fx := newEngineFixture()
		buf := &fact.Buffer{}
		err := fx.engine.Tick(
			context.Background(),
			WithObserver(buf),
			EnableHandler("<disabled-projection>", true),
		)
		if err != nil {
			t.Fatal(err)
		}

		h, _ := fx.cfg.HandlerByName("<disabled-projection>")
		xtesting.ExpectContains[fact.Fact](
			t,

			"expected TickBegun fact",
			buf.Facts(),
			fact.TickBegun{
				Handler: h,
			},
		)
	})

	t.Run("it always returns context errors even if other errors occur", func(t *testing.T) {
		fx := newEngineFixture()
		ctx, cancel := context.WithCancel(context.Background())

		fx.projection.CompactFunc = func(context.Context, dogma.ProjectionCompactScope) error {
			cancel()
			return errors.New("<error>")
		}

		err := fx.engine.Tick(ctx)
		xtesting.Expect(t, "unexpected error", err, context.Canceled)
	})

	t.Run("it adds handler details to controller errors", func(t *testing.T) {
		fx := newEngineFixture()
		fx.projection.CompactFunc = func(context.Context, dogma.ProjectionCompactScope) error {
			return errors.New("<error>")
		}

		err := fx.engine.Tick(context.Background())
		if err == nil || err.Error() != "<projection> projection: <error>" {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
