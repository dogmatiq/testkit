package engine_test

import (
	"context"
	"errors"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/enginekit/config/runtimeconfig"
	. "github.com/dogmatiq/enginekit/enginetest/stubs"
	. "github.com/dogmatiq/testkit/engine"
	g "github.com/onsi/ginkgo/v2"
	gm "github.com/onsi/gomega"
)

var _ = g.Describe("type CommandExecutor", func() {
	var (
		aggregate *AggregateMessageHandlerStub
		app       *ApplicationStub
		engine    *Engine
		executor  *CommandExecutor
	)

	g.BeforeEach(func() {
		aggregate = &AggregateMessageHandlerStub{
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

		app = &ApplicationStub{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "d905114d-b026-4f1a-9bc6-3abd86058e2d")
				c.Routes(
					dogma.ViaAggregate(aggregate),
				)
			},
		}

		engine = MustNew(runtimeconfig.FromApplication(app))

		executor = &CommandExecutor{
			Engine: engine,
		}
	})

	g.Describe("func ExecuteCommand()", func() {
		g.It("dispatches to the engine", func() {
			called := false
			aggregate.HandleCommandFunc = func(
				_ dogma.AggregateRoot,
				_ dogma.AggregateCommandScope,
				m dogma.Command,
			) {
				called = true
				gm.Expect(m).To(gm.Equal(CommandA1))
			}

			err := executor.ExecuteCommand(context.Background(), CommandA1)
			gm.Expect(err).ShouldNot(gm.HaveOccurred())
			gm.Expect(called).To(gm.BeTrue())
		})

		g.It("panics if the message is unrecognized", func() {
			gm.Expect(func() {
				executor.ExecuteCommand(context.Background(), CommandX1)
			}).To(gm.PanicWith("the *stubs.CommandStub[TypeX] message type is not consumed by any handlers"))
		})

		g.It("deduplicates commands with the same idempotency key", func() {
			callCount := 0
			aggregate.HandleCommandFunc = func(_ dogma.AggregateRoot, _ dogma.AggregateCommandScope, _ dogma.Command) {
				callCount++
			}

			err := executor.ExecuteCommand(context.Background(), CommandA1, dogma.WithIdempotencyKey("my-key"))
			gm.Expect(err).ShouldNot(gm.HaveOccurred())
			gm.Expect(callCount).To(gm.Equal(1))

			err = executor.ExecuteCommand(context.Background(), CommandA1, dogma.WithIdempotencyKey("my-key"))
			gm.Expect(err).ShouldNot(gm.HaveOccurred())
			gm.Expect(callCount).To(gm.Equal(1))
		})

		g.It("does not deduplicate commands with different idempotency keys", func() {
			callCount := 0
			aggregate.HandleCommandFunc = func(_ dogma.AggregateRoot, _ dogma.AggregateCommandScope, _ dogma.Command) {
				callCount++
			}

			err := executor.ExecuteCommand(context.Background(), CommandA1, dogma.WithIdempotencyKey("key-1"))
			gm.Expect(err).ShouldNot(gm.HaveOccurred())
			err = executor.ExecuteCommand(context.Background(), CommandA1, dogma.WithIdempotencyKey("key-2"))
			gm.Expect(err).ShouldNot(gm.HaveOccurred())
			gm.Expect(callCount).To(gm.Equal(2))
		})

		g.It("supports WithEventObserver()", func() {
			aggregate.HandleCommandFunc = func(
				_ dogma.AggregateRoot,
				s dogma.AggregateCommandScope,
				_ dogma.Command,
			) {
				s.RecordEvent(EventA1)
			}

			called := false
			err := executor.ExecuteCommand(
				context.Background(),
				CommandA1,
				dogma.WithEventObserver(
					func(_ context.Context, e *EventStub[TypeA]) (bool, error) {
						called = true
						gm.Expect(e).To(gm.Equal(EventA1))
						return true, nil
					},
				),
			)

			gm.Expect(err).ShouldNot(gm.HaveOccurred())
			gm.Expect(called).To(gm.BeTrue())
		})

		g.It("returns the observer error", func() {
			aggregate.HandleCommandFunc = func(
				_ dogma.AggregateRoot,
				s dogma.AggregateCommandScope,
				_ dogma.Command,
			) {
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

			gm.Expect(err).To(gm.MatchError("<observer-error>"))
		})

		g.It("returns ErrEventObserverNotSatisfied when observer is not called", func() {
			aggregate.HandleCommandFunc = func(
				_ dogma.AggregateRoot,
				s dogma.AggregateCommandScope,
				_ dogma.Command,
			) {
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

			gm.Expect(err).To(gm.MatchError(dogma.ErrEventObserverNotSatisfied))
		})

		g.It("returns ErrEventObserverNotSatisfied when observer returns false", func() {
			aggregate.HandleCommandFunc = func(
				_ dogma.AggregateRoot,
				s dogma.AggregateCommandScope,
				_ dogma.Command,
			) {
				s.RecordEvent(EventA1)
			}

			called := false
			err := executor.ExecuteCommand(
				context.Background(),
				CommandA1,
				dogma.WithEventObserver(
					func(_ context.Context, e *EventStub[TypeA]) (bool, error) {
						called = true
						gm.Expect(e).To(gm.Equal(EventA1))
						return false, nil
					},
				),
			)

			gm.Expect(err).To(gm.MatchError(dogma.ErrEventObserverNotSatisfied))
			gm.Expect(called).To(gm.BeTrue())
		})

		g.It("supports multiple event observers", func() {
			aggregate.HandleCommandFunc = func(
				_ dogma.AggregateRoot,
				s dogma.AggregateCommandScope,
				_ dogma.Command,
			) {
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

			gm.Expect(err).ShouldNot(gm.HaveOccurred())
		})

		g.It("returns ErrEventObserverNotSatisfied when all observers return false", func() {
			aggregate.HandleCommandFunc = func(
				_ dogma.AggregateRoot,
				s dogma.AggregateCommandScope,
				_ dogma.Command,
			) {
				s.RecordEvent(EventA1)
			}

			called1 := false
			called2 := false
			err := executor.ExecuteCommand(
				context.Background(),
				CommandA1,
				dogma.WithEventObserver(
					func(_ context.Context, e *EventStub[TypeA]) (bool, error) {
						called1 = true
						return false, nil
					},
				),
				dogma.WithEventObserver(
					func(_ context.Context, e *EventStub[TypeA]) (bool, error) {
						called2 = true
						return false, nil
					},
				),
			)

			gm.Expect(err).To(gm.MatchError(dogma.ErrEventObserverNotSatisfied))
			gm.Expect(called1).To(gm.BeTrue())
			gm.Expect(called2).To(gm.BeTrue())
		})

		g.It("returns success when any observer is satisfied among multiple", func() {
			aggregate.HandleCommandFunc = func(
				_ dogma.AggregateRoot,
				s dogma.AggregateCommandScope,
				_ dogma.Command,
			) {
				s.RecordEvent(EventA1)
			}

			called1 := false
			called2 := false
			err := executor.ExecuteCommand(
				context.Background(),
				CommandA1,
				dogma.WithEventObserver(
					func(_ context.Context, e *EventStub[TypeA]) (bool, error) {
						called1 = true
						return false, nil
					},
				),
				dogma.WithEventObserver(
					func(_ context.Context, e *EventStub[TypeA]) (bool, error) {
						called2 = true
						return true, nil
					},
				),
			)

			gm.Expect(err).ShouldNot(gm.HaveOccurred())
			gm.Expect(called1).To(gm.BeTrue())
			gm.Expect(called2).To(gm.BeTrue())
		})
	})

	g.Context("with an integration handler", func() {
		var (
			integration *IntegrationMessageHandlerStub
		)

		g.BeforeEach(func() {
			integration = &IntegrationMessageHandlerStub{
				ConfigureFunc: func(c dogma.IntegrationConfigurer) {
					c.Identity("<integration>", "c8b8a8e0-8e0a-4c2a-9f8e-8e0a4c2a9f8e")
					c.Routes(
						dogma.HandlesCommand[*CommandStub[TypeB]](),
						dogma.RecordsEvent[*EventStub[TypeB]](),
					)
				},
				HandleCommandFunc: func(
					_ context.Context,
					s dogma.IntegrationCommandScope,
					_ dogma.Command,
				) error {
					s.RecordEvent(EventB1)
					return nil
				},
			}

			app = &ApplicationStub{
				ConfigureFunc: func(c dogma.ApplicationConfigurer) {
					c.Identity("<app>", "d905114d-b026-4f1a-9bc6-3abd86058e2d")
					c.Routes(
						dogma.ViaIntegration(integration),
					)
				},
			}

			engine = MustNew(runtimeconfig.FromApplication(app))

			executor = &CommandExecutor{
				Engine: engine,
			}
		})

		g.It("observer is called for integration-recorded events", func() {
			called := false
			err := executor.ExecuteCommand(
				context.Background(),
				CommandB1,
				dogma.WithEventObserver(
					func(_ context.Context, e *EventStub[TypeB]) (bool, error) {
						called = true
						gm.Expect(e).To(gm.Equal(EventB1))
						return true, nil
					},
				),
			)

			gm.Expect(err).ShouldNot(gm.HaveOccurred())
			gm.Expect(called).To(gm.BeTrue())
		})
	})
})
