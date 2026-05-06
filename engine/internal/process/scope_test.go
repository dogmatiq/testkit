package process_test

import (
	"context"
	"testing"
	"time"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/enginekit/config"
	"github.com/dogmatiq/enginekit/config/runtimeconfig"
	. "github.com/dogmatiq/enginekit/enginetest/stubs"
	"github.com/dogmatiq/testkit/engine/internal/panicx"
	. "github.com/dogmatiq/testkit/engine/internal/process"
	"github.com/dogmatiq/testkit/envelope"
	"github.com/dogmatiq/testkit/fact"
	"github.com/dogmatiq/testkit/internal/x/xtesting"
)

func TestScope(t *testing.T) {
	t.Run("InstanceID", func(t *testing.T) {
		env := newProcessScopeTestEnv()
		called := false

		env.handler.HandleEventFunc = func(
			_ context.Context,
			_ *ProcessRootStub,
			s dogma.ProcessEventScope[*ProcessRootStub],
			_ dogma.Event,
		) error {
			called = true
			xtesting.Expect(t, "unexpected instance ID", s.InstanceID(), "<instance>")
			return nil
		}

		_, err := env.ctrl.Handle(
			context.Background(),
			fact.Ignore,
			time.Now(),
			env.event,
		)
		expectNoError(t, err)

		if !called {
			t.Fatal("expected HandleEvent() to be called")
		}
	})

	t.Run("End", func(t *testing.T) {
		t.Run("records a fact", func(t *testing.T) {
			env := newProcessScopeTestEnv()
			env.handler.HandleEventFunc = func(
				_ context.Context,
				_ *ProcessRootStub,
				s dogma.ProcessEventScope[*ProcessRootStub],
				_ dogma.Event,
			) error {
				s.End()
				return nil
			}

			buf := &fact.Buffer{}
			_, err := env.ctrl.Handle(
				context.Background(),
				buf,
				time.Now(),
				env.event,
			)
			expectNoError(t, err)

			expectFacts(
				t,
				buf.Facts(),
				[]fact.Fact{
					fact.ProcessInstanceNotFound{
						Handler:    env.cfg,
						InstanceID: "<instance>",
						Envelope:   env.event,
					},
					fact.ProcessInstanceBegun{
						Handler:    env.cfg,
						InstanceID: "<instance>",
						Root:       &ProcessRootStub{},
						Envelope:   env.event,
					},
					fact.ProcessInstanceEnded{
						Handler:    env.cfg,
						InstanceID: "<instance>",
						Root:       &ProcessRootStub{},
						Envelope:   env.event,
					},
				},
			)
		})

		t.Run("does nothing if the instance has already been ended", func(t *testing.T) {
			env := newProcessScopeTestEnv()
			env.handler.HandleEventFunc = func(
				_ context.Context,
				_ *ProcessRootStub,
				s dogma.ProcessEventScope[*ProcessRootStub],
				_ dogma.Event,
			) error {
				s.End()
				s.End()
				return nil
			}

			buf := &fact.Buffer{}
			_, err := env.ctrl.Handle(
				context.Background(),
				buf,
				time.Now(),
				env.event,
			)
			expectNoError(t, err)

			expectFacts(
				t,
				buf.Facts(),
				[]fact.Fact{
					fact.ProcessInstanceNotFound{
						Handler:    env.cfg,
						InstanceID: "<instance>",
						Envelope:   env.event,
					},
					fact.ProcessInstanceBegun{
						Handler:    env.cfg,
						InstanceID: "<instance>",
						Root:       &ProcessRootStub{},
						Envelope:   env.event,
					},
					fact.ProcessInstanceEnded{
						Handler:    env.cfg,
						InstanceID: "<instance>",
						Root:       &ProcessRootStub{},
						Envelope:   env.event,
					},
				},
			)
		})
	})

	t.Run("ExecuteCommand", func(t *testing.T) {
		t.Run("records a fact", func(t *testing.T) {
			env := newProcessScopeTestEnv()
			env.handler.HandleEventFunc = func(
				_ context.Context,
				_ *ProcessRootStub,
				s dogma.ProcessEventScope[*ProcessRootStub],
				_ dogma.Event,
			) error {
				s.ExecuteCommand(CommandA1)
				return nil
			}

			buf := &fact.Buffer{}
			now := time.Now()
			_, err := env.ctrl.Handle(
				context.Background(),
				buf,
				now,
				env.event,
			)
			expectNoError(t, err)

			expectFacts(
				t,
				buf.Facts(),
				[]fact.Fact{
					fact.ProcessInstanceNotFound{
						Handler:    env.cfg,
						InstanceID: "<instance>",
						Envelope:   env.event,
					},
					fact.ProcessInstanceBegun{
						Handler:    env.cfg,
						InstanceID: "<instance>",
						Root:       &ProcessRootStub{},
						Envelope:   env.event,
					},
					fact.CommandExecutedByProcess{
						Handler:    env.cfg,
						InstanceID: "<instance>",
						Root:       &ProcessRootStub{},
						Envelope:   env.event,
						CommandEnvelope: env.event.NewCommand(
							"1",
							CommandA1,
							now,
							envelope.Origin{
								Handler:     env.cfg,
								HandlerType: config.ProcessHandlerType,
								InstanceID:  "<instance>",
							},
						),
					},
				},
			)
		})

		t.Run("panics if the command type is not configured to be produced", func(t *testing.T) {
			env := newProcessScopeTestEnv()
			env.handler.HandleEventFunc = func(
				_ context.Context,
				_ *ProcessRootStub,
				s dogma.ProcessEventScope[*ProcessRootStub],
				_ dogma.Event,
			) error {
				s.ExecuteCommand(CommandX1)
				return nil
			}

			expectUnexpectedBehavior(
				t,
				func() {
					_, _ = env.ctrl.Handle(
						context.Background(),
						fact.Ignore,
						time.Now(),
						env.event,
					)
				},
				panicx.UnexpectedBehavior{
					Handler:        env.cfg,
					Interface:      "ProcessMessageHandler",
					Method:         "HandleEvent",
					Implementation: env.cfg.Implementation(),
					Message:        env.event.Message,
					Description:    "executed a command of type *stubs.CommandStub[TypeX], which is not produced by this handler",
				},
				"/engine/internal/process/scope_test.go",
			)
		})

		t.Run("panics if the command is invalid", func(t *testing.T) {
			env := newProcessScopeTestEnv()
			env.handler.HandleEventFunc = func(
				_ context.Context,
				_ *ProcessRootStub,
				s dogma.ProcessEventScope[*ProcessRootStub],
				_ dogma.Event,
			) error {
				s.ExecuteCommand(&CommandStub[TypeA]{
					ValidationError: "<invalid>",
				})
				return nil
			}

			expectUnexpectedBehavior(
				t,
				func() {
					_, _ = env.ctrl.Handle(
						context.Background(),
						fact.Ignore,
						time.Now(),
						env.event,
					)
				},
				panicx.UnexpectedBehavior{
					Handler:        env.cfg,
					Interface:      "ProcessMessageHandler",
					Method:         "HandleEvent",
					Implementation: env.cfg.Implementation(),
					Message:        env.event.Message,
					Description:    "executed an invalid *stubs.CommandStub[TypeA] command: <invalid>",
				},
				"/engine/internal/process/scope_test.go",
			)
		})

		t.Run("panics if the process has ended", func(t *testing.T) {
			env := newProcessScopeTestEnv()
			env.handler.HandleEventFunc = func(
				_ context.Context,
				_ *ProcessRootStub,
				s dogma.ProcessEventScope[*ProcessRootStub],
				_ dogma.Event,
			) error {
				s.End()
				s.ExecuteCommand(CommandA1)
				return nil
			}

			expectUnexpectedBehavior(
				t,
				func() {
					_, _ = env.ctrl.Handle(
						context.Background(),
						fact.Ignore,
						time.Now(),
						env.event,
					)
				},
				panicx.UnexpectedBehavior{
					Handler:        env.cfg,
					Interface:      "ProcessMessageHandler",
					Method:         "HandleEvent",
					Implementation: env.cfg.Implementation(),
					Message:        env.event.Message,
					Description:    "executed a command of type *stubs.CommandStub[TypeA] on an ended process",
				},
				"/engine/internal/process/scope_test.go",
			)
		})
	})

	t.Run("ScheduleDeadline", func(t *testing.T) {
		t.Run("records a fact", func(t *testing.T) {
			env := newProcessScopeTestEnv()
			scheduledFor := time.Now().Add(10 * time.Second)
			env.handler.HandleEventFunc = func(
				_ context.Context,
				_ *ProcessRootStub,
				s dogma.ProcessEventScope[*ProcessRootStub],
				_ dogma.Event,
			) error {
				s.ScheduleDeadline(DeadlineA1, scheduledFor)
				return nil
			}

			buf := &fact.Buffer{}
			now := time.Now()
			_, err := env.ctrl.Handle(
				context.Background(),
				buf,
				now,
				env.event,
			)
			expectNoError(t, err)

			expectFacts(
				t,
				buf.Facts(),
				[]fact.Fact{
					fact.ProcessInstanceNotFound{
						Handler:    env.cfg,
						InstanceID: "<instance>",
						Envelope:   env.event,
					},
					fact.ProcessInstanceBegun{
						Handler:    env.cfg,
						InstanceID: "<instance>",
						Root:       &ProcessRootStub{},
						Envelope:   env.event,
					},
					fact.DeadlineScheduledByProcess{
						Handler:    env.cfg,
						InstanceID: "<instance>",
						Root:       &ProcessRootStub{},
						Envelope:   env.event,
						DeadlineEnvelope: env.event.NewDeadline(
							"1",
							DeadlineA1,
							now,
							scheduledFor,
							envelope.Origin{
								Handler:     env.cfg,
								HandlerType: config.ProcessHandlerType,
								InstanceID:  "<instance>",
							},
						),
					},
				},
			)
		})

		t.Run("panics if the deadline type is not configured to be scheduled", func(t *testing.T) {
			env := newProcessScopeTestEnv()
			env.handler.HandleEventFunc = func(
				_ context.Context,
				_ *ProcessRootStub,
				s dogma.ProcessEventScope[*ProcessRootStub],
				_ dogma.Event,
			) error {
				s.ScheduleDeadline(DeadlineX1, time.Now())
				return nil
			}

			expectUnexpectedBehavior(
				t,
				func() {
					_, _ = env.ctrl.Handle(
						context.Background(),
						fact.Ignore,
						time.Now(),
						env.event,
					)
				},
				panicx.UnexpectedBehavior{
					Handler:        env.cfg,
					Interface:      "ProcessMessageHandler",
					Method:         "HandleEvent",
					Implementation: env.cfg.Implementation(),
					Message:        env.event.Message,
					Description:    "scheduled a deadline of type *stubs.DeadlineStub[TypeX], which is not produced by this handler",
				},
				"/engine/internal/process/scope_test.go",
			)
		})

		t.Run("panics if the deadline is invalid", func(t *testing.T) {
			env := newProcessScopeTestEnv()
			env.handler.HandleEventFunc = func(
				_ context.Context,
				_ *ProcessRootStub,
				s dogma.ProcessEventScope[*ProcessRootStub],
				_ dogma.Event,
			) error {
				s.ScheduleDeadline(
					&DeadlineStub[TypeA]{
						ValidationError: "<invalid>",
					},
					time.Now(),
				)
				return nil
			}

			expectUnexpectedBehavior(
				t,
				func() {
					_, _ = env.ctrl.Handle(
						context.Background(),
						fact.Ignore,
						time.Now(),
						env.event,
					)
				},
				panicx.UnexpectedBehavior{
					Handler:        env.cfg,
					Interface:      "ProcessMessageHandler",
					Method:         "HandleEvent",
					Implementation: env.cfg.Implementation(),
					Message:        env.event.Message,
					Description:    "scheduled an invalid *stubs.DeadlineStub[TypeA] deadline: <invalid>",
				},
				"/engine/internal/process/scope_test.go",
			)
		})

		t.Run("panics if the process has ended", func(t *testing.T) {
			env := newProcessScopeTestEnv()
			scheduledFor := time.Now().Add(10 * time.Second)
			env.handler.HandleEventFunc = func(
				_ context.Context,
				_ *ProcessRootStub,
				s dogma.ProcessEventScope[*ProcessRootStub],
				_ dogma.Event,
			) error {
				s.End()
				s.ScheduleDeadline(DeadlineA1, scheduledFor)
				return nil
			}

			expectUnexpectedBehavior(
				t,
				func() {
					_, _ = env.ctrl.Handle(
						context.Background(),
						fact.Ignore,
						time.Now(),
						env.event,
					)
				},
				panicx.UnexpectedBehavior{
					Handler:        env.cfg,
					Interface:      "ProcessMessageHandler",
					Method:         "HandleEvent",
					Implementation: env.cfg.Implementation(),
					Message:        env.event.Message,
					Description:    "scheduled a deadline of type *stubs.DeadlineStub[TypeA] on an ended process",
				},
				"/engine/internal/process/scope_test.go",
			)
		})
	})

	t.Run("ScheduledFor", func(t *testing.T) {
		env := newProcessScopeTestEnv()

		_, err := env.ctrl.Handle(
			context.Background(),
			fact.Ignore,
			time.Now(),
			env.event,
		)
		expectNoError(t, err)

		deadline := env.event.NewDeadline(
			"2000",
			DeadlineA1,
			time.Now(),
			time.Now().Add(10*time.Second),
			envelope.Origin{
				Handler:     env.cfg,
				HandlerType: config.ProcessHandlerType,
				InstanceID:  "<instance>",
			},
		)

		env.handler.HandleDeadlineFunc = func(
			_ context.Context,
			_ *ProcessRootStub,
			s dogma.ProcessDeadlineScope[*ProcessRootStub],
			_ dogma.Deadline,
		) error {
			if !s.ScheduledFor().Equal(deadline.ScheduledFor) {
				t.Fatalf(
					"unexpected scheduled time: got %s, want %s",
					s.ScheduledFor(),
					deadline.ScheduledFor,
				)
			}
			return nil
		}

		_, err = env.ctrl.Handle(
			context.Background(),
			fact.Ignore,
			time.Now(),
			deadline,
		)
		expectNoError(t, err)
	})

	t.Run("Mutate", func(t *testing.T) {
		t.Run("calls the function with the instance root", func(t *testing.T) {
			env := newProcessScopeTestEnv()
			called := false

			env.handler.HandleEventFunc = func(
				_ context.Context,
				_ *ProcessRootStub,
				s dogma.ProcessEventScope[*ProcessRootStub],
				_ dogma.Event,
			) error {
				s.Mutate(func(r *ProcessRootStub) {
					called = true
					r.Value = "<mutated>"
				})
				return nil
			}

			_, err := env.ctrl.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				env.event,
			)
			expectNoError(t, err)

			if !called {
				t.Fatal("expected Mutate() to call the function")
			}
		})

		t.Run("panics if the process has ended", func(t *testing.T) {
			env := newProcessScopeTestEnv()
			env.handler.HandleEventFunc = func(
				_ context.Context,
				_ *ProcessRootStub,
				s dogma.ProcessEventScope[*ProcessRootStub],
				_ dogma.Event,
			) error {
				s.End()
				s.Mutate(func(*ProcessRootStub) {})
				return nil
			}

			expectUnexpectedBehavior(
				t,
				func() {
					_, _ = env.ctrl.Handle(
						context.Background(),
						fact.Ignore,
						time.Now(),
						env.event,
					)
				},
				panicx.UnexpectedBehavior{
					Handler:        env.cfg,
					Interface:      "ProcessMessageHandler",
					Method:         "HandleEvent",
					Implementation: env.cfg.Implementation(),
					Message:        env.event.Message,
					Description:    "mutated an ended process instance",
				},
				"/engine/internal/process/scope_test.go",
			)
		})
	})

	t.Run("Log", func(t *testing.T) {
		env := newProcessScopeTestEnv()
		env.handler.HandleEventFunc = func(
			_ context.Context,
			_ *ProcessRootStub,
			s dogma.ProcessEventScope[*ProcessRootStub],
			_ dogma.Event,
		) error {
			s.Log("<format>", "<arg-1>", "<arg-2>")
			return nil
		}

		buf := &fact.Buffer{}
		_, err := env.ctrl.Handle(
			context.Background(),
			buf,
			time.Now(),
			env.event,
		)
		expectNoError(t, err)

		expectFacts(
			t,
			buf.Facts(),
			[]fact.Fact{
				fact.ProcessInstanceNotFound{
					Handler:    env.cfg,
					InstanceID: "<instance>",
					Envelope:   env.event,
				},
				fact.ProcessInstanceBegun{
					Handler:    env.cfg,
					InstanceID: "<instance>",
					Root:       &ProcessRootStub{},
					Envelope:   env.event,
				},
				fact.MessageLoggedByProcess{
					Handler:    env.cfg,
					InstanceID: "<instance>",
					Root:       &ProcessRootStub{},
					Envelope:   env.event,
					LogFormat:  "<format>",
					LogArguments: []any{
						"<arg-1>",
						"<arg-2>",
					},
				},
			},
		)
	})
}

type processScopeTestEnv struct {
	messageIDs envelope.MessageIDGenerator
	handler    *ProcessMessageHandlerStub[*ProcessRootStub]
	cfg        *config.Process
	ctrl       *Controller
	event      *envelope.Envelope
}

func newProcessScopeTestEnv() *processScopeTestEnv {
	event := envelope.NewEvent(
		"1000",
		EventA1,
		time.Now(),
	)

	handler := &ProcessMessageHandlerStub[*ProcessRootStub]{
		ConfigureFunc: func(c dogma.ProcessConfigurer) {
			c.Identity("<name>", "6901c34c-6e4d-4184-9414-780cb21a791a")
			c.Routes(
				dogma.HandlesEvent[*EventStub[TypeA]](),
				dogma.ExecutesCommand[*CommandStub[TypeA]](),
				dogma.SchedulesDeadline[*DeadlineStub[TypeA]](),
			)
		},
		RouteEventToInstanceFunc: func(
			_ context.Context,
			m dogma.Event,
		) (string, bool, error) {
			switch m.(type) {
			case *EventStub[TypeA]:
				return "<instance>", true, nil
			default:
				panic(dogma.UnexpectedMessage)
			}
		},
	}

	cfg := runtimeconfig.FromProcess(handler)
	env := &processScopeTestEnv{
		handler: handler,
		cfg:     cfg,
		event:   event,
	}

	env.ctrl = &Controller{
		Config:     cfg,
		MessageIDs: &env.messageIDs,
	}

	env.messageIDs.Reset()

	return env
}
