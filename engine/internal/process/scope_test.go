package process_test

import (
	"context"
	"strings"
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
		f := newProcessTestFixture()
		called := false

		f.handler.HandleEventFunc = func(
			_ context.Context,
			_ *ProcessRootStub,
			s dogma.ProcessEventScope[*ProcessRootStub],
			_ dogma.Event,
		) error {
			called = true
			xtesting.Expect(t, "unexpected instance ID", s.InstanceID(), "<instance>")
			return nil
		}

		_, err := f.ctrl.Handle(
			context.Background(),
			fact.Ignore,
			time.Now(),
			f.event,
		)
		if err != nil {
			t.Fatal(err)
		}

		if !called {
			t.Fatal("expected HandleEvent() to be called")
		}
	})

	t.Run("End", func(t *testing.T) {
		t.Run("records a fact", func(t *testing.T) {
			f := newProcessTestFixture()
			f.handler.HandleEventFunc = func(
				_ context.Context,
				_ *ProcessRootStub,
				s dogma.ProcessEventScope[*ProcessRootStub],
				_ dogma.Event,
			) error {
				s.End()
				return nil
			}

			buf := &fact.Buffer{}
			_, err := f.ctrl.Handle(
				context.Background(),
				buf,
				time.Now(),
				f.event,
			)
			if err != nil {
				t.Fatal(err)
			}

			xtesting.Expect(
				t,
				"unexpected facts",
				buf.Facts(),
				[]fact.Fact{
					fact.ProcessInstanceNotFound{
						Handler:    f.cfg,
						InstanceID: "<instance>",
						Envelope:   f.event,
					},
					fact.ProcessInstanceBegun{
						Handler:    f.cfg,
						InstanceID: "<instance>",
						Root:       &ProcessRootStub{},
						Envelope:   f.event,
					},
					fact.ProcessInstanceEnded{
						Handler:    f.cfg,
						InstanceID: "<instance>",
						Root:       &ProcessRootStub{},
						Envelope:   f.event,
					},
				},
			)
		})

		t.Run("does nothing if the instance has already been ended", func(t *testing.T) {
			f := newProcessTestFixture()
			f.handler.HandleEventFunc = func(
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
			_, err := f.ctrl.Handle(
				context.Background(),
				buf,
				time.Now(),
				f.event,
			)
			if err != nil {
				t.Fatal(err)
			}

			xtesting.Expect(
				t,
				"unexpected facts",
				buf.Facts(),
				[]fact.Fact{
					fact.ProcessInstanceNotFound{
						Handler:    f.cfg,
						InstanceID: "<instance>",
						Envelope:   f.event,
					},
					fact.ProcessInstanceBegun{
						Handler:    f.cfg,
						InstanceID: "<instance>",
						Root:       &ProcessRootStub{},
						Envelope:   f.event,
					},
					fact.ProcessInstanceEnded{
						Handler:    f.cfg,
						InstanceID: "<instance>",
						Root:       &ProcessRootStub{},
						Envelope:   f.event,
					},
				},
			)
		})
	})

	t.Run("ExecuteCommand", func(t *testing.T) {
		t.Run("records a fact", func(t *testing.T) {
			f := newProcessTestFixture()
			f.handler.HandleEventFunc = func(
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
			_, err := f.ctrl.Handle(
				context.Background(),
				buf,
				now,
				f.event,
			)
			if err != nil {
				t.Fatal(err)
			}

			xtesting.Expect(
				t,
				"unexpected facts",
				buf.Facts(),
				[]fact.Fact{
					fact.ProcessInstanceNotFound{
						Handler:    f.cfg,
						InstanceID: "<instance>",
						Envelope:   f.event,
					},
					fact.ProcessInstanceBegun{
						Handler:    f.cfg,
						InstanceID: "<instance>",
						Root:       &ProcessRootStub{},
						Envelope:   f.event,
					},
					fact.CommandExecutedByProcess{
						Handler:    f.cfg,
						InstanceID: "<instance>",
						Root:       &ProcessRootStub{},
						Envelope:   f.event,
						CommandEnvelope: f.event.NewCommand(
							"1",
							CommandA1,
							now,
							envelope.Origin{
								Handler:     f.cfg,
								HandlerType: config.ProcessHandlerType,
								InstanceID:  "<instance>",
							},
						),
					},
				},
			)
		})

		t.Run("panics if the command type is not configured to be produced", func(t *testing.T) {
			f := newProcessTestFixture()
			f.handler.HandleEventFunc = func(
				_ context.Context,
				_ *ProcessRootStub,
				s dogma.ProcessEventScope[*ProcessRootStub],
				_ dogma.Event,
			) error {
				s.ExecuteCommand(CommandX1)
				return nil
			}

			xtesting.ExpectPanicMatching(t, func() {
				_, _ = f.ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					f.event,
				)
			}, func(x panicx.UnexpectedBehavior) {
				xtesting.Expect(t, "unexpected handler", x.Handler, f.cfg)
				xtesting.Expect(t, "unexpected interface", x.Interface, "ProcessMessageHandler")
				xtesting.Expect(t, "unexpected method", x.Method, "HandleEvent")
				xtesting.Expect(t, "unexpected implementation", x.Implementation, f.cfg.Implementation())
				xtesting.Expect(t, "unexpected message", x.Message, f.event.Message)
				xtesting.Expect(t, "unexpected description", x.Description, "executed a command of type *stubs.CommandStub[TypeX], which is not produced by this handler")
				xtesting.ExpectLocation(t, x.Location, "/engine/internal/process/scope_test.go")
			})
		})

		t.Run("panics if the command is invalid", func(t *testing.T) {
			f := newProcessTestFixture()
			f.handler.HandleEventFunc = func(
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

			xtesting.ExpectPanicMatching(t, func() {
				_, _ = f.ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					f.event,
				)
			}, func(x panicx.UnexpectedBehavior) {
				xtesting.Expect(t, "unexpected handler", x.Handler, f.cfg)
				xtesting.Expect(t, "unexpected interface", x.Interface, "ProcessMessageHandler")
				xtesting.Expect(t, "unexpected method", x.Method, "HandleEvent")
				xtesting.Expect(t, "unexpected implementation", x.Implementation, f.cfg.Implementation())
				xtesting.Expect(t, "unexpected message", x.Message, f.event.Message)
				xtesting.Expect(t, "unexpected description", x.Description, "executed an invalid *stubs.CommandStub[TypeA] command: <invalid>")
				xtesting.ExpectLocation(t, x.Location, "/engine/internal/process/scope_test.go")
			})
		})

		t.Run("panics if the process has ended", func(t *testing.T) {
			f := newProcessTestFixture()
			f.handler.HandleEventFunc = func(
				_ context.Context,
				_ *ProcessRootStub,
				s dogma.ProcessEventScope[*ProcessRootStub],
				_ dogma.Event,
			) error {
				s.End()
				s.ExecuteCommand(CommandA1)
				return nil
			}

			xtesting.ExpectPanicMatching(t, func() {
				_, _ = f.ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					f.event,
				)
			}, func(x panicx.UnexpectedBehavior) {
				xtesting.Expect(t, "unexpected handler", x.Handler, f.cfg)
				xtesting.Expect(t, "unexpected interface", x.Interface, "ProcessMessageHandler")
				xtesting.Expect(t, "unexpected method", x.Method, "HandleEvent")
				xtesting.Expect(t, "unexpected implementation", x.Implementation, f.cfg.Implementation())
				xtesting.Expect(t, "unexpected message", x.Message, f.event.Message)
				xtesting.Expect(t, "unexpected description", x.Description, "executed a command of type *stubs.CommandStub[TypeA] on an ended process")
				xtesting.ExpectLocation(t, x.Location, "/engine/internal/process/scope_test.go")
			})
		})
	})

	t.Run("ScheduleDeadline", func(t *testing.T) {
		t.Run("records a fact", func(t *testing.T) {
			f := newProcessTestFixture()
			scheduledFor := time.Now().Add(10 * time.Second)
			f.handler.HandleEventFunc = func(
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
			_, err := f.ctrl.Handle(
				context.Background(),
				buf,
				now,
				f.event,
			)
			if err != nil {
				t.Fatal(err)
			}

			xtesting.Expect(
				t,
				"unexpected facts",
				buf.Facts(),
				[]fact.Fact{
					fact.ProcessInstanceNotFound{
						Handler:    f.cfg,
						InstanceID: "<instance>",
						Envelope:   f.event,
					},
					fact.ProcessInstanceBegun{
						Handler:    f.cfg,
						InstanceID: "<instance>",
						Root:       &ProcessRootStub{},
						Envelope:   f.event,
					},
					fact.DeadlineScheduledByProcess{
						Handler:    f.cfg,
						InstanceID: "<instance>",
						Root:       &ProcessRootStub{},
						Envelope:   f.event,
						DeadlineEnvelope: f.event.NewDeadline(
							"1",
							DeadlineA1,
							now,
							scheduledFor,
							envelope.Origin{
								Handler:     f.cfg,
								HandlerType: config.ProcessHandlerType,
								InstanceID:  "<instance>",
							},
						),
					},
				},
			)
		})

		t.Run("panics if the deadline type is not configured to be scheduled", func(t *testing.T) {
			f := newProcessTestFixture()
			f.handler.HandleEventFunc = func(
				_ context.Context,
				_ *ProcessRootStub,
				s dogma.ProcessEventScope[*ProcessRootStub],
				_ dogma.Event,
			) error {
				s.ScheduleDeadline(DeadlineX1, time.Now())
				return nil
			}

			xtesting.ExpectPanicMatching(t, func() {
				_, _ = f.ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					f.event,
				)
			}, func(x panicx.UnexpectedBehavior) {
				xtesting.Expect(t, "unexpected handler", x.Handler, f.cfg)
				xtesting.Expect(t, "unexpected interface", x.Interface, "ProcessMessageHandler")
				xtesting.Expect(t, "unexpected method", x.Method, "HandleEvent")
				xtesting.Expect(t, "unexpected implementation", x.Implementation, f.cfg.Implementation())
				xtesting.Expect(t, "unexpected message", x.Message, f.event.Message)
				xtesting.Expect(t, "unexpected description", x.Description, "scheduled a deadline of type *stubs.DeadlineStub[TypeX], which is not produced by this handler")
				xtesting.ExpectLocation(t, x.Location, "/engine/internal/process/scope_test.go")
			})
		})

		t.Run("panics if the deadline is invalid", func(t *testing.T) {
			f := newProcessTestFixture()
			f.handler.HandleEventFunc = func(
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

			xtesting.ExpectPanicMatching(t, func() {
				_, _ = f.ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					f.event,
				)
			}, func(x panicx.UnexpectedBehavior) {
				xtesting.Expect(t, "unexpected handler", x.Handler, f.cfg)
				xtesting.Expect(t, "unexpected interface", x.Interface, "ProcessMessageHandler")
				xtesting.Expect(t, "unexpected method", x.Method, "HandleEvent")
				xtesting.Expect(t, "unexpected implementation", x.Implementation, f.cfg.Implementation())
				xtesting.Expect(t, "unexpected message", x.Message, f.event.Message)
				xtesting.Expect(t, "unexpected description", x.Description, "scheduled an invalid *stubs.DeadlineStub[TypeA] deadline: <invalid>")
				xtesting.ExpectLocation(t, x.Location, "/engine/internal/process/scope_test.go")
			})
		})

		t.Run("panics if the process has ended", func(t *testing.T) {
			f := newProcessTestFixture()
			scheduledFor := time.Now().Add(10 * time.Second)
			f.handler.HandleEventFunc = func(
				_ context.Context,
				_ *ProcessRootStub,
				s dogma.ProcessEventScope[*ProcessRootStub],
				_ dogma.Event,
			) error {
				s.End()
				s.ScheduleDeadline(DeadlineA1, scheduledFor)
				return nil
			}

			xtesting.ExpectPanicMatching(t, func() {
				_, _ = f.ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					f.event,
				)
			}, func(x panicx.UnexpectedBehavior) {
				xtesting.Expect(t, "unexpected handler", x.Handler, f.cfg)
				xtesting.Expect(t, "unexpected interface", x.Interface, "ProcessMessageHandler")
				xtesting.Expect(t, "unexpected method", x.Method, "HandleEvent")
				xtesting.Expect(t, "unexpected implementation", x.Implementation, f.cfg.Implementation())
				xtesting.Expect(t, "unexpected message", x.Message, f.event.Message)
				xtesting.Expect(t, "unexpected description", x.Description, "scheduled a deadline of type *stubs.DeadlineStub[TypeA] on an ended process")
				xtesting.ExpectLocation(t, x.Location, "/engine/internal/process/scope_test.go")
			})
		})
	})

	t.Run("ScheduledFor", func(t *testing.T) {
		f := newProcessTestFixture()

		_, err := f.ctrl.Handle(
			context.Background(),
			fact.Ignore,
			time.Now(),
			f.event,
		)
		if err != nil {
			t.Fatal(err)
		}

		deadline := f.event.NewDeadline(
			"2000",
			DeadlineA1,
			time.Now(),
			time.Now().Add(10*time.Second),
			envelope.Origin{
				Handler:     f.cfg,
				HandlerType: config.ProcessHandlerType,
				InstanceID:  "<instance>",
			},
		)

		f.handler.HandleDeadlineFunc = func(
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

		_, err = f.ctrl.Handle(
			context.Background(),
			fact.Ignore,
			time.Now(),
			deadline,
		)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Mutate", func(t *testing.T) {
		t.Run("calls the function with the instance root", func(t *testing.T) {
			f := newProcessTestFixture()
			called := false

			f.handler.HandleEventFunc = func(
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

			_, err := f.ctrl.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				f.event,
			)
			if err != nil {
				t.Fatal(err)
			}

			if !called {
				t.Fatal("expected Mutate() to call the function")
			}
		})

		t.Run("panics if the process has ended", func(t *testing.T) {
			f := newProcessTestFixture()
			f.handler.HandleEventFunc = func(
				_ context.Context,
				_ *ProcessRootStub,
				s dogma.ProcessEventScope[*ProcessRootStub],
				_ dogma.Event,
			) error {
				s.End()
				s.Mutate(func(*ProcessRootStub) {})
				return nil
			}

			xtesting.ExpectPanicMatching(t, func() {
				_, _ = f.ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					f.event,
				)
			}, func(x panicx.UnexpectedBehavior) {
				xtesting.Expect(t, "unexpected handler", x.Handler, f.cfg)
				xtesting.Expect(t, "unexpected interface", x.Interface, "ProcessMessageHandler")
				xtesting.Expect(t, "unexpected method", x.Method, "HandleEvent")
				xtesting.Expect(t, "unexpected implementation", x.Implementation, f.cfg.Implementation())
				xtesting.Expect(t, "unexpected message", x.Message, f.event.Message)
				xtesting.Expect(t, "unexpected description", x.Description, "mutated an ended process instance")
				xtesting.ExpectLocation(t, x.Location, "/engine/internal/process/scope_test.go")
			})
		})
	})

	t.Run("Log", func(t *testing.T) {
		f := newProcessTestFixture()
		f.handler.HandleEventFunc = func(
			_ context.Context,
			_ *ProcessRootStub,
			s dogma.ProcessEventScope[*ProcessRootStub],
			_ dogma.Event,
		) error {
			s.Log("<format>", "<arg-1>", "<arg-2>")
			return nil
		}

		buf := &fact.Buffer{}
		_, err := f.ctrl.Handle(
			context.Background(),
			buf,
			time.Now(),
			f.event,
		)
		if err != nil {
			t.Fatal(err)
		}

		xtesting.Expect(
			t,
			"unexpected facts",
			buf.Facts(),
			[]fact.Fact{
				fact.ProcessInstanceNotFound{
					Handler:    f.cfg,
					InstanceID: "<instance>",
					Envelope:   f.event,
				},
				fact.ProcessInstanceBegun{
					Handler:    f.cfg,
					InstanceID: "<instance>",
					Root:       &ProcessRootStub{},
					Envelope:   f.event,
				},
				fact.MessageLoggedByProcess{
					Handler:    f.cfg,
					InstanceID: "<instance>",
					Root:       &ProcessRootStub{},
					Envelope:   f.event,
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

func TestMutationDetection(t *testing.T) {
	t.Run("panics if the handler modifies the root before calling ExecuteCommand", func(t *testing.T) {
		f := newProcessTestFixture()
		f.handler.HandleEventFunc = func(
			_ context.Context,
			r *ProcessRootStub,
			s dogma.ProcessEventScope[*ProcessRootStub],
			_ dogma.Event,
		) error {
			r.Value = "<mutated>"
			s.ExecuteCommand(CommandA1)
			return nil
		}

		xtesting.ExpectPanicMatching(t, func() {
			_, _ = f.ctrl.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				f.event,
			)
		}, func(x panicx.UnexpectedBehavior) {
			wantPrefix := "modified the process root without using Mutate(), before call to ExecuteCommand() at"
			if !strings.HasPrefix(x.Description, wantPrefix) {
				t.Fatalf("unexpected panic description: %s", x.Description)
			}
		})
	})

	t.Run("panics if the handler modifies the root before calling InstanceID", func(t *testing.T) {
		f := newProcessTestFixture()
		f.handler.HandleEventFunc = func(
			_ context.Context,
			r *ProcessRootStub,
			s dogma.ProcessEventScope[*ProcessRootStub],
			_ dogma.Event,
		) error {
			r.Value = "<mutated>"
			s.InstanceID()
			return nil
		}

		xtesting.ExpectPanicMatching(t, func() {
			_, _ = f.ctrl.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				f.event,
			)
		}, func(x panicx.UnexpectedBehavior) {
			wantPrefix := "modified the process root without using Mutate(), before call to InstanceID() at"
			if !strings.HasPrefix(x.Description, wantPrefix) {
				t.Fatalf("unexpected panic description: %s", x.Description)
			}
		})
	})

	t.Run("panics if the handler modifies the root before calling Now", func(t *testing.T) {
		f := newProcessTestFixture()
		f.handler.HandleEventFunc = func(
			_ context.Context,
			r *ProcessRootStub,
			s dogma.ProcessEventScope[*ProcessRootStub],
			_ dogma.Event,
		) error {
			r.Value = "<mutated>"
			s.Now()
			return nil
		}

		xtesting.ExpectPanicMatching(t, func() {
			_, _ = f.ctrl.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				f.event,
			)
		}, func(x panicx.UnexpectedBehavior) {
			wantPrefix := "modified the process root without using Mutate(), before call to Now() at"
			if !strings.HasPrefix(x.Description, wantPrefix) {
				t.Fatalf("unexpected panic description: %s", x.Description)
			}
		})
	})

	t.Run("panics if the handler modifies the root before calling Log", func(t *testing.T) {
		f := newProcessTestFixture()
		f.handler.HandleEventFunc = func(
			_ context.Context,
			r *ProcessRootStub,
			s dogma.ProcessEventScope[*ProcessRootStub],
			_ dogma.Event,
		) error {
			r.Value = "<mutated>"
			s.Log("hello")
			return nil
		}

		xtesting.ExpectPanicMatching(t, func() {
			_, _ = f.ctrl.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				f.event,
			)
		}, func(x panicx.UnexpectedBehavior) {
			wantPrefix := "modified the process root without using Mutate(), before call to Log() at"
			if !strings.HasPrefix(x.Description, wantPrefix) {
				t.Fatalf("unexpected panic description: %s", x.Description)
			}
		})
	})

	t.Run("panics if the handler modifies the root before calling End", func(t *testing.T) {
		f := newProcessTestFixture()
		f.handler.HandleEventFunc = func(
			_ context.Context,
			r *ProcessRootStub,
			s dogma.ProcessEventScope[*ProcessRootStub],
			_ dogma.Event,
		) error {
			r.Value = "<mutated>"
			s.End()
			return nil
		}

		xtesting.ExpectPanicMatching(t, func() {
			_, _ = f.ctrl.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				f.event,
			)
		}, func(x panicx.UnexpectedBehavior) {
			wantPrefix := "modified the process root without using Mutate(), before call to End() at"
			if !strings.HasPrefix(x.Description, wantPrefix) {
				t.Fatalf("unexpected panic description: %s", x.Description)
			}
		})
	})

	t.Run("panics if the handler modifies the root before calling Mutate", func(t *testing.T) {
		f := newProcessTestFixture()
		f.handler.HandleEventFunc = func(
			_ context.Context,
			r *ProcessRootStub,
			s dogma.ProcessEventScope[*ProcessRootStub],
			_ dogma.Event,
		) error {
			r.Value = "<mutated>"
			s.Mutate(func(*ProcessRootStub) {})
			return nil
		}

		xtesting.ExpectPanicMatching(t, func() {
			_, _ = f.ctrl.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				f.event,
			)
		}, func(x panicx.UnexpectedBehavior) {
			wantPrefix := "modified the process root without using Mutate(), before call to Mutate() at"
			if !strings.HasPrefix(x.Description, wantPrefix) {
				t.Fatalf("unexpected panic description: %s", x.Description)
			}
		})
	})

	t.Run("panics if the handler modifies the root before calling ScheduleDeadline", func(t *testing.T) {
		f := newProcessTestFixture()
		f.handler.HandleEventFunc = func(
			_ context.Context,
			r *ProcessRootStub,
			s dogma.ProcessEventScope[*ProcessRootStub],
			_ dogma.Event,
		) error {
			r.Value = "<mutated>"
			s.ScheduleDeadline(DeadlineA1, time.Now())
			return nil
		}

		xtesting.ExpectPanicMatching(t, func() {
			_, _ = f.ctrl.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				f.event,
			)
		}, func(x panicx.UnexpectedBehavior) {
			wantPrefix := "modified the process root without using Mutate(), before call to ScheduleDeadline() at"
			if !strings.HasPrefix(x.Description, wantPrefix) {
				t.Fatalf("unexpected panic description: %s", x.Description)
			}
		})
	})

	t.Run("panics if the handler modifies the root between two scope calls", func(t *testing.T) {
		f := newProcessTestFixture()
		f.handler.HandleEventFunc = func(
			_ context.Context,
			r *ProcessRootStub,
			s dogma.ProcessEventScope[*ProcessRootStub],
			_ dogma.Event,
		) error {
			s.InstanceID()
			r.Value = "<mutated>"
			s.ExecuteCommand(CommandA1)
			return nil
		}

		xtesting.ExpectPanicMatching(t, func() {
			_, _ = f.ctrl.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				f.event,
			)
		}, func(x panicx.UnexpectedBehavior) {
			wantPrefix := "modified the process root without using Mutate(), between call to InstanceID() at"
			if !strings.HasPrefix(x.Description, wantPrefix) {
				t.Fatalf("unexpected panic description: %s", x.Description)
			}
		})
	})

	t.Run("panics at end of handler if the root was modified without a scope call", func(t *testing.T) {
		f := newProcessTestFixture()
		f.handler.HandleEventFunc = func(
			_ context.Context,
			r *ProcessRootStub,
			_ dogma.ProcessEventScope[*ProcessRootStub],
			_ dogma.Event,
		) error {
			r.Value = "<mutated>"
			return nil
		}

		xtesting.ExpectPanicMatching(t, func() {
			_, _ = f.ctrl.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				f.event,
			)
		}, func(x panicx.UnexpectedBehavior) {
			xtesting.Expect(t, "unexpected description", x.Description, "modified the process root without using Mutate()")
		})
	})
}

func TestNonDeterministicMutate(t *testing.T) {
	t.Run("panics if Mutate callback produces different state on each call", func(t *testing.T) {
		f := newProcessTestFixture()

		callCount := 0
		f.handler.HandleEventFunc = func(
			_ context.Context,
			_ *ProcessRootStub,
			s dogma.ProcessEventScope[*ProcessRootStub],
			_ dogma.Event,
		) error {
			s.Mutate(func(r *ProcessRootStub) {
				callCount++
				if callCount == 1 {
					r.Value = "<first>"
				} else {
					r.Value = "<second>"
				}
			})
			return nil
		}

		xtesting.ExpectPanicMatching(t, func() {
			_, _ = f.ctrl.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				f.event,
			)
		}, func(x panicx.UnexpectedBehavior) {
			xtesting.Expect(t, "unexpected description", x.Description, "non-deterministic implementation of Mutate() callback detected")
		})
	})
}

type processTestFixture struct {
	messageIDs envelope.MessageIDGenerator
	handler    *ProcessMessageHandlerStub[*ProcessRootStub]
	cfg        *config.Process
	ctrl       *Controller
	event      *envelope.Envelope
}

func newProcessTestFixture() *processTestFixture {
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
	f := &processTestFixture{
		handler: handler,
		cfg:     cfg,
		event:   event,
	}

	f.ctrl = &Controller{
		Config:     cfg,
		MessageIDs: &f.messageIDs,
	}

	f.messageIDs.Reset()

	return f
}
