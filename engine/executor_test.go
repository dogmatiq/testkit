package engine_test

import (
	"context"
	"errors"
	"testing"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/enginekit/config/runtimeconfig"
	. "github.com/dogmatiq/enginekit/enginetest/stubs"
	. "github.com/dogmatiq/testkit/engine"
	"github.com/dogmatiq/testkit/x/xtesting"
)

func newCommandExecutorFixture() (*AggregateMessageHandlerStub, *CommandExecutor) {
	aggregate := &AggregateMessageHandlerStub{
		ConfigureFunc: func(c dogma.AggregateConfigurer) {
			c.Identity("<aggregate>", "4acf3050-8d02-4052-a9af-abb9e67add78")
			c.Routes(
				dogma.HandlesCommand[*CommandStub[TypeA]](),
				dogma.RecordsEvent[*EventStub[TypeA]](),
			)
		},
		RouteCommandToInstanceFunc: func(dogma.Command) string {
			return "<instance>"
		},
	}

	app := &ApplicationStub{
		ConfigureFunc: func(c dogma.ApplicationConfigurer) {
			c.Identity("<app>", "d905114d-b026-4f1a-9bc6-3abd86058e2d")
			c.Routes(
				dogma.ViaAggregate(aggregate),
			)
		},
	}

	executor := &CommandExecutor{
		Engine: MustNew(runtimeconfig.FromApplication(app)),
	}

	return aggregate, executor
}

func TestCommandExecutor_ExecuteCommand(t *testing.T) {
	t.Run("it dispatches to the engine", func(t *testing.T) {
		aggregate, executor := newCommandExecutorFixture()
		called := false
		aggregate.HandleCommandFunc = func(_ dogma.AggregateRoot, _ dogma.AggregateCommandScope, m dogma.Command) {
			called = true
			xtesting.Expect(t, "unexpected command", m, CommandA1)
		}

		err := executor.ExecuteCommand(context.Background(), CommandA1)
		if err != nil {
			t.Fatal(err)
		}
		if !called {
			t.Fatal("expected HandleCommand to be called")
		}
	})

	t.Run("it panics if the message is unrecognized", func(t *testing.T) {
		_, executor := newCommandExecutorFixture()
		xtesting.ExpectPanic(
			t,
			"the *stubs.CommandStub[TypeX] message type is not consumed by any handlers",
			func() {
				executor.ExecuteCommand(context.Background(), CommandX1) //nolint:errcheck
			},
		)
	})

	t.Run("it deduplicates commands with the same idempotency key", func(t *testing.T) {
		aggregate, executor := newCommandExecutorFixture()
		callCount := 0
		aggregate.HandleCommandFunc = func(_ dogma.AggregateRoot, _ dogma.AggregateCommandScope, _ dogma.Command) {
			callCount++
		}

		err := executor.ExecuteCommand(context.Background(), CommandA1, dogma.WithIdempotencyKey("my-key"))
		if err != nil {
			t.Fatal(err)
		}
		xtesting.Expect(t, "unexpected call count after first dispatch", callCount, 1)

		err = executor.ExecuteCommand(context.Background(), CommandA1, dogma.WithIdempotencyKey("my-key"))
		if err != nil {
			t.Fatal(err)
		}
		xtesting.Expect(t, "unexpected call count after second dispatch", callCount, 1)
	})

	t.Run("it does not deduplicate commands with different idempotency keys", func(t *testing.T) {
		aggregate, executor := newCommandExecutorFixture()
		callCount := 0
		aggregate.HandleCommandFunc = func(_ dogma.AggregateRoot, _ dogma.AggregateCommandScope, _ dogma.Command) {
			callCount++
		}

		err := executor.ExecuteCommand(context.Background(), CommandA1, dogma.WithIdempotencyKey("key-1"))
		if err != nil {
			t.Fatal(err)
		}
		err = executor.ExecuteCommand(context.Background(), CommandA1, dogma.WithIdempotencyKey("key-2"))
		if err != nil {
			t.Fatal(err)
		}
		xtesting.Expect(t, "unexpected call count", callCount, 2)
	})

	t.Run("it supports WithEventObserver()", func(t *testing.T) {
		aggregate, executor := newCommandExecutorFixture()
		aggregate.HandleCommandFunc = func(_ dogma.AggregateRoot, s dogma.AggregateCommandScope, _ dogma.Command) {
			s.RecordEvent(EventA1)
		}

		called := false
		err := executor.ExecuteCommand(
			context.Background(),
			CommandA1,
			dogma.WithEventObserver(
				func(_ context.Context, e *EventStub[TypeA]) (bool, error) {
					called = true
					xtesting.Expect(t, "unexpected event", e, EventA1)
					return true, nil
				},
			),
		)
		if err != nil {
			t.Fatal(err)
		}
		if !called {
			t.Fatal("expected observer to be called")
		}
	})

	t.Run("it returns the observer error", func(t *testing.T) {
		aggregate, executor := newCommandExecutorFixture()
		aggregate.HandleCommandFunc = func(_ dogma.AggregateRoot, s dogma.AggregateCommandScope, _ dogma.Command) {
			s.RecordEvent(EventA1)
		}

		err := executor.ExecuteCommand(
			context.Background(),
			CommandA1,
			dogma.WithEventObserver(
				func(context.Context, *EventStub[TypeA]) (bool, error) {
					return false, errors.New("<observer-error>")
				},
			),
		)
		if err == nil || err.Error() != "<observer-error>" {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("it returns ErrEventObserverNotSatisfied when observer is not called", func(t *testing.T) {
		aggregate, executor := newCommandExecutorFixture()
		aggregate.HandleCommandFunc = func(_ dogma.AggregateRoot, s dogma.AggregateCommandScope, _ dogma.Command) {
			s.RecordEvent(EventA1)
		}

		err := executor.ExecuteCommand(
			context.Background(),
			CommandA1,
			dogma.WithEventObserver(
				func(context.Context, *EventStub[TypeB]) (bool, error) {
					return true, nil
				},
			),
		)
		xtesting.Expect(t, "unexpected error", err, dogma.ErrEventObserverNotSatisfied)
	})

	t.Run("it returns ErrEventObserverNotSatisfied when observer returns false", func(t *testing.T) {
		aggregate, executor := newCommandExecutorFixture()
		aggregate.HandleCommandFunc = func(_ dogma.AggregateRoot, s dogma.AggregateCommandScope, _ dogma.Command) {
			s.RecordEvent(EventA1)
		}

		called := false
		err := executor.ExecuteCommand(
			context.Background(),
			CommandA1,
			dogma.WithEventObserver(
				func(_ context.Context, e *EventStub[TypeA]) (bool, error) {
					called = true
					xtesting.Expect(t, "unexpected event", e, EventA1)
					return false, nil
				},
			),
		)
		xtesting.Expect(t, "unexpected error", err, dogma.ErrEventObserverNotSatisfied)
		if !called {
			t.Fatal("expected observer to be called")
		}
	})

	t.Run("it supports multiple event observers", func(t *testing.T) {
		aggregate, executor := newCommandExecutorFixture()
		aggregate.HandleCommandFunc = func(_ dogma.AggregateRoot, s dogma.AggregateCommandScope, _ dogma.Command) {
			s.RecordEvent(EventA1)
		}

		err := executor.ExecuteCommand(
			context.Background(),
			CommandA1,
			dogma.WithEventObserver(
				func(context.Context, *EventStub[TypeB]) (bool, error) {
					return true, nil
				},
			),
			dogma.WithEventObserver(
				func(context.Context, *EventStub[TypeA]) (bool, error) {
					return true, nil
				},
			),
		)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("it returns ErrEventObserverNotSatisfied when all observers return false", func(t *testing.T) {
		aggregate, executor := newCommandExecutorFixture()
		aggregate.HandleCommandFunc = func(_ dogma.AggregateRoot, s dogma.AggregateCommandScope, _ dogma.Command) {
			s.RecordEvent(EventA1)
		}

		called1 := false
		called2 := false
		err := executor.ExecuteCommand(
			context.Background(),
			CommandA1,
			dogma.WithEventObserver(
				func(_ context.Context, _ *EventStub[TypeA]) (bool, error) {
					called1 = true
					return false, nil
				},
			),
			dogma.WithEventObserver(
				func(_ context.Context, _ *EventStub[TypeA]) (bool, error) {
					called2 = true
					return false, nil
				},
			),
		)
		xtesting.Expect(t, "unexpected error", err, dogma.ErrEventObserverNotSatisfied)
		if !called1 {
			t.Fatal("expected first observer to be called")
		}
		if !called2 {
			t.Fatal("expected second observer to be called")
		}
	})

	t.Run("it returns success when any observer is satisfied among multiple", func(t *testing.T) {
		aggregate, executor := newCommandExecutorFixture()
		aggregate.HandleCommandFunc = func(_ dogma.AggregateRoot, s dogma.AggregateCommandScope, _ dogma.Command) {
			s.RecordEvent(EventA1)
		}

		called1 := false
		called2 := false
		err := executor.ExecuteCommand(
			context.Background(),
			CommandA1,
			dogma.WithEventObserver(
				func(_ context.Context, _ *EventStub[TypeA]) (bool, error) {
					called1 = true
					return false, nil
				},
			),
			dogma.WithEventObserver(
				func(_ context.Context, _ *EventStub[TypeA]) (bool, error) {
					called2 = true
					return true, nil
				},
			),
		)
		if err != nil {
			t.Fatal(err)
		}
		if !called1 {
			t.Fatal("expected first observer to be called")
		}
		if !called2 {
			t.Fatal("expected second observer to be called")
		}
	})

	t.Run("with an integration handler", func(t *testing.T) {
		integration := &IntegrationMessageHandlerStub{
			ConfigureFunc: func(c dogma.IntegrationConfigurer) {
				c.Identity("<integration>", "c8b8a8e0-8e0a-4c2a-9f8e-8e0a4c2a9f8e")
				c.Routes(
					dogma.HandlesCommand[*CommandStub[TypeB]](),
					dogma.RecordsEvent[*EventStub[TypeB]](),
				)
			},
			HandleCommandFunc: func(_ context.Context, s dogma.IntegrationCommandScope, _ dogma.Command) error {
				s.RecordEvent(EventB1)
				return nil
			},
		}

		app := &ApplicationStub{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "d905114d-b026-4f1a-9bc6-3abd86058e2d")
				c.Routes(
					dogma.ViaIntegration(integration),
				)
			},
		}

		executor := &CommandExecutor{
			Engine: MustNew(runtimeconfig.FromApplication(app)),
		}

		t.Run("it observer is called for integration-recorded events", func(t *testing.T) {
			called := false
			err := executor.ExecuteCommand(
				context.Background(),
				CommandB1,
				dogma.WithEventObserver(
					func(_ context.Context, e *EventStub[TypeB]) (bool, error) {
						called = true
						xtesting.Expect(t, "unexpected event", e, EventB1)
						return true, nil
					},
				),
			)
			if err != nil {
				t.Fatal(err)
			}
			if !called {
				t.Fatal("expected observer to be called")
			}
		})
	})
}
