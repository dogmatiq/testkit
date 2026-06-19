package process_test

import (
	"context"
	"errors"
	"fmt"
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

func TestController(t *testing.T) {
	t.Run("HandlerConfig", func(t *testing.T) {
		f := newControllerTestFixture()

		if got := f.ctrl.HandlerConfig(); got != f.cfg {
			t.Fatalf("unexpected handler config: got %p, want %p", got, f.cfg)
		}
	})

	t.Run("Tick", func(t *testing.T) {
		setup := func(t *testing.T) (*controllerTestFixture, time.Time, time.Time, time.Time, time.Time) {
			t.Helper()

			f := newControllerTestFixture()
			createdTime := time.Now()
			t1Time := createdTime.Add(1 * time.Hour)
			t2Time := createdTime.Add(2 * time.Hour)
			t3Time := createdTime.Add(3 * time.Hour)

			f.handler.HandleEventFunc = func(
				_ context.Context,
				_ *ProcessRootStub,
				s dogma.ProcessEventScope[*ProcessRootStub],
				_ dogma.Event,
			) error {
				s.ScheduleDeadline(DeadlineA3, t3Time)
				s.ScheduleDeadline(DeadlineA2, t2Time)
				s.ScheduleDeadline(DeadlineA1, t1Time)
				return nil
			}

			_, err := f.ctrl.Handle(
				context.Background(),
				fact.Ignore,
				createdTime,
				f.event,
			)
			if err != nil {
				t.Fatal(err)
			}

			f.messageIDs.Reset()

			return f, createdTime, t1Time, t2Time, t3Time
		}

		t.Run("returns deadlines that are ready to be handled", func(t *testing.T) {
			f, createdTime, t1Time, t2Time, _ := setup(t)

			deadlines, err := f.ctrl.Tick(
				context.Background(),
				fact.Ignore,
				t2Time,
			)
			if err != nil {
				t.Fatal(err)
			}

			xtesting.ExpectSet(
				t,
				"unexpected envelopes",
				deadlines,
				[]*envelope.Envelope{
					f.event.NewDeadline(
						"3",
						DeadlineA1,
						createdTime,
						t1Time,
						envelope.Origin{
							Handler:     f.cfg,
							HandlerType: config.ProcessHandlerType,
							InstanceID:  "<instance-A1>",
						},
					),
					f.event.NewDeadline(
						"2",
						DeadlineA2,
						createdTime,
						t2Time,
						envelope.Origin{
							Handler:     f.cfg,
							HandlerType: config.ProcessHandlerType,
							InstanceID:  "<instance-A1>",
						},
					),
				},
				func(a, b *envelope.Envelope) bool { return a.MessageID < b.MessageID },
			)
		})

		t.Run("does not return the same deadlines multiple times", func(t *testing.T) {
			f, _, _, t2Time, _ := setup(t)

			deadlines, err := f.ctrl.Tick(
				context.Background(),
				fact.Ignore,
				t2Time,
			)
			if err != nil {
				t.Fatal(err)
			}

			if got, want := len(deadlines), 2; got != want {
				t.Fatalf("unexpected deadline count: got %d, want %d", got, want)
			}

			deadlines, err = f.ctrl.Tick(
				context.Background(),
				fact.Ignore,
				t2Time,
			)
			if err != nil {
				t.Fatal(err)
			}

			if len(deadlines) != 0 {
				t.Fatalf("unexpected deadlines: got %d, want 0", len(deadlines))
			}
		})

		t.Run("does not return deadlines for instances that have been ended", func(t *testing.T) {
			f, createdTime, t1Time, t2Time, _ := setup(t)

			secondInstanceEvent := envelope.NewEvent(
				"3000",
				EventA2,
				time.Now(),
			)

			_, err := f.ctrl.Handle(
				context.Background(),
				fact.Ignore,
				createdTime,
				secondInstanceEvent,
			)
			if err != nil {
				t.Fatal(err)
			}

			f.handler.HandleEventFunc = func(
				_ context.Context,
				_ *ProcessRootStub,
				s dogma.ProcessEventScope[*ProcessRootStub],
				_ dogma.Event,
			) error {
				s.End()
				return nil
			}

			_, err = f.ctrl.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				f.event,
			)
			if err != nil {
				t.Fatal(err)
			}

			deadlines, err := f.ctrl.Tick(
				context.Background(),
				fact.Ignore,
				t2Time,
			)
			if err != nil {
				t.Fatal(err)
			}

			xtesting.ExpectSet(
				t,
				"unexpected envelopes",
				deadlines,
				[]*envelope.Envelope{
					secondInstanceEvent.NewDeadline(
						"3",
						DeadlineA1,
						createdTime,
						t1Time,
						envelope.Origin{
							Handler:     f.cfg,
							HandlerType: config.ProcessHandlerType,
							InstanceID:  "<instance-A2>",
						},
					),
					secondInstanceEvent.NewDeadline(
						"2",
						DeadlineA2,
						createdTime,
						t2Time,
						envelope.Origin{
							Handler:     f.cfg,
							HandlerType: config.ProcessHandlerType,
							InstanceID:  "<instance-A2>",
						},
					),
				},
				func(a, b *envelope.Envelope) bool {
					return a.MessageID < b.MessageID
				},
			)
		})
	})

	t.Run("Handle", func(t *testing.T) {
		t.Run("handling an event", func(t *testing.T) {
			t.Run("forwards the message to the handler", func(t *testing.T) {
				f := newControllerTestFixture()
				called := false

				f.handler.HandleEventFunc = func(
					_ context.Context,
					_ *ProcessRootStub,
					_ dogma.ProcessEventScope[*ProcessRootStub],
					m dogma.Event,
				) error {
					called = true
					xtesting.Expect(t, "unexpected event", m, EventA1)
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

			t.Run("propagates handler errors", func(t *testing.T) {
				f := newControllerTestFixture()
				expected := errors.New("<error>")

				f.handler.HandleEventFunc = func(
					_ context.Context,
					_ *ProcessRootStub,
					_ dogma.ProcessEventScope[*ProcessRootStub],
					_ dogma.Event,
				) error {
					return expected
				}

				_, err := f.ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					f.event,
				)

				xtesting.Expect(t, "unexpected error", err, expected)
			})

			t.Run("returns both commands and deadlines", func(t *testing.T) {
				f := newControllerTestFixture()
				now := time.Now()

				f.handler.HandleEventFunc = func(
					_ context.Context,
					_ *ProcessRootStub,
					s dogma.ProcessEventScope[*ProcessRootStub],
					_ dogma.Event,
				) error {
					s.ExecuteCommand(CommandA1)
					s.ScheduleDeadline(DeadlineA1, now)
					return nil
				}

				envelopes, err := f.ctrl.Handle(
					context.Background(),
					fact.Ignore,
					now,
					f.event,
				)
				if err != nil {
					t.Fatal(err)
				}

				xtesting.ExpectSet(
					t,
					"unexpected envelopes",
					envelopes,
					[]*envelope.Envelope{
						f.event.NewCommand(
							"1",
							CommandA1,
							now,
							envelope.Origin{
								Handler:     f.cfg,
								HandlerType: config.ProcessHandlerType,
								InstanceID:  "<instance-A1>",
							},
						),
						f.event.NewDeadline(
							"2",
							DeadlineA1,
							now,
							now,
							envelope.Origin{
								Handler:     f.cfg,
								HandlerType: config.ProcessHandlerType,
								InstanceID:  "<instance-A1>",
							},
						),
					},
					func(a, b *envelope.Envelope) bool {
						return a.MessageID < b.MessageID
					},
				)
			})

			t.Run("returns deadlines scheduled in the past", func(t *testing.T) {
				f := newControllerTestFixture()
				now := time.Now()

				f.handler.HandleEventFunc = func(
					_ context.Context,
					_ *ProcessRootStub,
					s dogma.ProcessEventScope[*ProcessRootStub],
					_ dogma.Event,
				) error {
					s.ScheduleDeadline(DeadlineA1, now.Add(-1))
					return nil
				}

				envelopes, err := f.ctrl.Handle(
					context.Background(),
					fact.Ignore,
					now,
					f.event,
				)
				if err != nil {
					t.Fatal(err)
				}

				if got, want := len(envelopes), 1; got != want {
					t.Fatalf("unexpected envelope count: got %d, want %d", got, want)
				}
			})

			t.Run("does not return deadlines scheduled in the future", func(t *testing.T) {
				f := newControllerTestFixture()
				now := time.Now()

				f.handler.HandleEventFunc = func(
					_ context.Context,
					_ *ProcessRootStub,
					s dogma.ProcessEventScope[*ProcessRootStub],
					_ dogma.Event,
				) error {
					s.ScheduleDeadline(DeadlineA1, now.Add(1))
					return nil
				}

				envelopes, err := f.ctrl.Handle(
					context.Background(),
					fact.Ignore,
					now,
					f.event,
				)
				if err != nil {
					t.Fatal(err)
				}

				if len(envelopes) != 0 {
					t.Fatalf("unexpected envelopes: got %d, want 0", len(envelopes))
				}
			})

			t.Run("when the event is not routed to an instance", func(t *testing.T) {
				t.Run("does not forward the message to the handler", func(t *testing.T) {
					f := newControllerTestFixture()
					f.handler.RouteEventToInstanceFunc = func(
						context.Context,
						dogma.Event,
					) (string, bool, error) {
						return "", false, nil
					}

					f.handler.HandleEventFunc = func(
						context.Context,
						*ProcessRootStub,
						dogma.ProcessEventScope[*ProcessRootStub],
						dogma.Event,
					) error {
						t.Fatal("unexpected call to HandleEvent()")
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
				})

				t.Run("records a fact", func(t *testing.T) {
					f := newControllerTestFixture()
					f.handler.RouteEventToInstanceFunc = func(
						context.Context,
						dogma.Event,
					) (string, bool, error) {
						return "", false, nil
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
							fact.ProcessEventIgnored{
								Handler:  f.cfg,
								Envelope: f.event,
							},
						},
					)
				})
			})

			t.Run("when the instance has ended", func(t *testing.T) {
				setup := func(t *testing.T) *controllerTestFixture {
					t.Helper()

					f := newControllerTestFixture()
					f.handler.HandleEventFunc = func(
						_ context.Context,
						_ *ProcessRootStub,
						s dogma.ProcessEventScope[*ProcessRootStub],
						_ dogma.Event,
					) error {
						s.End()
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

					f.messageIDs.Reset()
					return f
				}

				t.Run("does not forward the message to the handler", func(t *testing.T) {
					f := setup(t)
					f.handler.HandleEventFunc = func(
						context.Context,
						*ProcessRootStub,
						dogma.ProcessEventScope[*ProcessRootStub],
						dogma.Event,
					) error {
						t.Fatal("unexpected call to HandleEvent()")
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
				})

				t.Run("records a fact", func(t *testing.T) {
					f := setup(t)
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
							fact.ProcessEventRoutedToEndedInstance{
								Handler:    f.cfg,
								InstanceID: "<instance-A1>",
								Envelope:   f.event,
							},
						},
					)
				})
			})
		})

		t.Run("handling a deadline", func(t *testing.T) {
			setup := func(t *testing.T) *controllerTestFixture {
				t.Helper()

				f := newControllerTestFixture()
				f.handler.HandleEventFunc = func(
					context.Context,
					*ProcessRootStub,
					dogma.ProcessEventScope[*ProcessRootStub],
					dogma.Event,
				) error {
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

				f.messageIDs.Reset()
				return f
			}

			t.Run("forwards the message to the handler", func(t *testing.T) {
				f := setup(t)
				called := false

				f.handler.HandleDeadlineFunc = func(
					_ context.Context,
					_ *ProcessRootStub,
					_ dogma.ProcessDeadlineScope[*ProcessRootStub],
					m dogma.Deadline,
				) error {
					called = true
					xtesting.Expect(t, "unexpected deadline", m, DeadlineA1)
					return nil
				}

				_, err := f.ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					f.deadline,
				)
				if err != nil {
					t.Fatal(err)
				}

				if !called {
					t.Fatal("expected HandleDeadline() to be called")
				}
			})

			t.Run("propagates handler errors", func(t *testing.T) {
				f := setup(t)
				expected := errors.New("<error>")

				f.handler.HandleDeadlineFunc = func(
					context.Context,
					*ProcessRootStub,
					dogma.ProcessDeadlineScope[*ProcessRootStub],
					dogma.Deadline,
				) error {
					return expected
				}

				_, err := f.ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					f.deadline,
				)

				xtesting.Expect(t, "unexpected error", err, expected)
			})

			t.Run("returns both commands and deadlines", func(t *testing.T) {
				f := setup(t)
				now := time.Now()

				f.handler.HandleDeadlineFunc = func(
					_ context.Context,
					_ *ProcessRootStub,
					s dogma.ProcessDeadlineScope[*ProcessRootStub],
					_ dogma.Deadline,
				) error {
					s.ExecuteCommand(CommandA1)
					s.ScheduleDeadline(DeadlineA1, now)
					return nil
				}

				envelopes, err := f.ctrl.Handle(
					context.Background(),
					fact.Ignore,
					now,
					f.deadline,
				)
				if err != nil {
					t.Fatal(err)
				}

				xtesting.ExpectSet(
					t,
					"unexpected envelopes",
					envelopes,
					[]*envelope.Envelope{
						f.deadline.NewCommand(
							"1",
							CommandA1,
							now,
							envelope.Origin{
								Handler:     f.cfg,
								HandlerType: config.ProcessHandlerType,
								InstanceID:  "<instance-A1>",
							},
						),
						f.deadline.NewDeadline(
							"2",
							DeadlineA1,
							now,
							now,
							envelope.Origin{
								Handler:     f.cfg,
								HandlerType: config.ProcessHandlerType,
								InstanceID:  "<instance-A1>",
							},
						),
					},
					func(a, b *envelope.Envelope) bool {
						return a.MessageID < b.MessageID
					},
				)
			})

			t.Run("returns deadlines scheduled in the past", func(t *testing.T) {
				f := setup(t)
				now := time.Now()

				f.handler.HandleDeadlineFunc = func(
					_ context.Context,
					_ *ProcessRootStub,
					s dogma.ProcessDeadlineScope[*ProcessRootStub],
					_ dogma.Deadline,
				) error {
					s.ScheduleDeadline(DeadlineA2, now.Add(-1))
					return nil
				}

				envelopes, err := f.ctrl.Handle(
					context.Background(),
					fact.Ignore,
					now,
					f.deadline,
				)
				if err != nil {
					t.Fatal(err)
				}

				if got, want := len(envelopes), 1; got != want {
					t.Fatalf("unexpected envelope count: got %d, want %d", got, want)
				}
			})

			t.Run("does not return deadlines scheduled in the future", func(t *testing.T) {
				f := setup(t)
				now := time.Now()

				f.handler.HandleDeadlineFunc = func(
					_ context.Context,
					_ *ProcessRootStub,
					s dogma.ProcessDeadlineScope[*ProcessRootStub],
					_ dogma.Deadline,
				) error {
					s.ScheduleDeadline(DeadlineA2, now.Add(1))
					return nil
				}

				envelopes, err := f.ctrl.Handle(
					context.Background(),
					fact.Ignore,
					now,
					f.deadline,
				)
				if err != nil {
					t.Fatal(err)
				}

				if len(envelopes) != 0 {
					t.Fatalf("unexpected envelopes: got %d, want 0", len(envelopes))
				}
			})

			t.Run("when the instance that created the deadline has ended", func(t *testing.T) {
				setupEnded := func(t *testing.T) *controllerTestFixture {
					t.Helper()

					f := setup(t)
					f.handler.HandleEventFunc = func(
						_ context.Context,
						_ *ProcessRootStub,
						s dogma.ProcessEventScope[*ProcessRootStub],
						_ dogma.Event,
					) error {
						s.End()
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

					f.messageIDs.Reset()
					return f
				}

				t.Run("does not forward the message to the handler", func(t *testing.T) {
					f := setupEnded(t)
					f.handler.HandleDeadlineFunc = func(
						context.Context,
						*ProcessRootStub,
						dogma.ProcessDeadlineScope[*ProcessRootStub],
						dogma.Deadline,
					) error {
						t.Fatal("unexpected call to HandleDeadline()")
						return nil
					}

					_, err := f.ctrl.Handle(
						context.Background(),
						fact.Ignore,
						time.Now(),
						f.deadline,
					)
					if err != nil {
						t.Fatal(err)
					}
				})

				t.Run("records a fact", func(t *testing.T) {
					f := setupEnded(t)
					buf := &fact.Buffer{}

					_, err := f.ctrl.Handle(
						context.Background(),
						buf,
						time.Now(),
						f.deadline,
					)
					if err != nil {
						t.Fatal(err)
					}

					xtesting.Expect(
						t,
						"unexpected facts",
						buf.Facts(),
						[]fact.Fact{
							fact.ProcessDeadlineRoutedToEndedInstance{
								Handler:    f.cfg,
								InstanceID: "<instance-A1>",
								Envelope:   f.deadline,
							},
						},
					)
				})
			})
		})

		t.Run("propagates routing errors", func(t *testing.T) {
			f := newControllerTestFixture()
			expected := errors.New("<error>")

			f.handler.RouteEventToInstanceFunc = func(
				context.Context,
				dogma.Event,
			) (string, bool, error) {
				return "<instance>", true, expected
			}

			_, err := f.ctrl.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				f.event,
			)

			xtesting.Expect(t, "unexpected error", err, expected)
		})

		t.Run("panics when the handler routes to an empty instance ID", func(t *testing.T) {
			f := newControllerTestFixture()

			f.handler.RouteEventToInstanceFunc = func(
				context.Context,
				dogma.Event,
			) (string, bool, error) {
				return "", true, nil
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
				xtesting.Expect(t, "unexpected method", x.Method, "RouteEventToInstance")
				xtesting.Expect(t, "unexpected implementation", x.Implementation, f.cfg.Implementation())
				xtesting.Expect(t, "unexpected message", x.Message, f.event.Message)
				xtesting.Expect(t, "unexpected description", x.Description, "routed an event of type *stubs.EventStub[TypeA] to an empty ID")
				xtesting.ExpectLocation(t, x.Location, "/stubs/process.go")
			})
		})

		t.Run("when the instance does not exist", func(t *testing.T) {
			t.Run("records facts", func(t *testing.T) {
				f := newControllerTestFixture()
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
							InstanceID: "<instance-A1>",
							Envelope:   f.event,
						},
						fact.ProcessInstanceBegun{
							Handler:    f.cfg,
							InstanceID: "<instance-A1>",
							Root:       &ProcessRootStub{},
							Envelope:   f.event,
						},
					},
				)
			})

			t.Run("panics if New returns nil", func(t *testing.T) {
				f := newControllerTestFixture()
				f.handler.NewFunc = func() *ProcessRootStub {
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
					xtesting.Expect(t, "unexpected method", x.Method, "New")
					xtesting.Expect(t, "unexpected implementation", x.Implementation, f.cfg.Implementation())
					xtesting.Expect(t, "unexpected message", x.Message, f.event.Message)
					xtesting.Expect(t, "unexpected description", x.Description, "returned a nil process root")
					xtesting.ExpectLocation(t, x.Location, "/stubs/process.go")
				})
			})
		})

		t.Run("when the instance exists", func(t *testing.T) {
			setup := func(t *testing.T) *controllerTestFixture {
				t.Helper()

				f := newControllerTestFixture()
				f.handler.HandleEventFunc = func(
					_ context.Context,
					_ *ProcessRootStub,
					_ dogma.ProcessEventScope[*ProcessRootStub],
					_ dogma.Event,
				) error {
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

				f.messageIDs.Reset()
				return f
			}

			t.Run("records a fact", func(t *testing.T) {
				f := setup(t)
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
						fact.ProcessInstanceLoaded{
							Handler:    f.cfg,
							InstanceID: "<instance-A1>",
							Root:       &ProcessRootStub{},
							Envelope:   f.event,
						},
					},
				)
			})

			t.Run("provides the root with state from the prior Handle() call", func(t *testing.T) {
				f := newControllerTestFixture()

				f.handler.HandleEventFunc = func(
					_ context.Context,
					r *ProcessRootStub,
					s dogma.ProcessEventScope[*ProcessRootStub],
					_ dogma.Event,
				) error {
					s.Mutate(func(r *ProcessRootStub) {
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

				var got string
				f.handler.HandleEventFunc = func(
					_ context.Context,
					r *ProcessRootStub,
					_ dogma.ProcessEventScope[*ProcessRootStub],
					_ dogma.Event,
				) error {
					got = r.Value.(string)
					return nil
				}

				_, err = f.ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					f.event,
				)
				if err != nil {
					t.Fatal(err)
				}

				if got != "<mutated>" {
					t.Fatalf("expected root state to persist, got %q", got)
				}
			})

			t.Run("panics if New() returns nil", func(t *testing.T) {
				f := setup(t)
				f.handler.NewFunc = func() *ProcessRootStub {
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
					xtesting.Expect(t, "unexpected method", x.Method, "New")
					xtesting.Expect(t, "unexpected implementation", x.Implementation, f.cfg.Implementation())
					xtesting.Expect(t, "unexpected message", x.Message, f.event.Message)
					xtesting.Expect(t, "unexpected description", x.Description, "returned a nil process root")
					xtesting.ExpectLocation(t, x.Location, "/stubs/process.go")
				})
			})
		})

		t.Run("provides more context to UnexpectedMessage panics from RouteEventToInstance", func(t *testing.T) {
			f := newControllerTestFixture()
			f.handler.RouteEventToInstanceFunc = func(
				context.Context,
				dogma.Event,
			) (string, bool, error) {
				panic(dogma.UnexpectedMessage)
			}

			xtesting.ExpectPanicMatching(t, func() {
				_, _ = f.ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					f.event,
				)
			}, func(x panicx.UnexpectedMessage) {
				xtesting.Expect(t, "unexpected handler", x.Handler, f.cfg)
				xtesting.Expect(t, "unexpected interface", x.Interface, "ProcessMessageHandler")
				xtesting.Expect(t, "unexpected method", x.Method, "RouteEventToInstance")
				xtesting.Expect(t, "unexpected message", x.Message, f.event.Message)
			})
		})

		t.Run("provides more context to UnexpectedMessage panics from HandleEvent", func(t *testing.T) {
			f := newControllerTestFixture()
			f.handler.HandleEventFunc = func(
				context.Context,
				*ProcessRootStub,
				dogma.ProcessEventScope[*ProcessRootStub],
				dogma.Event,
			) error {
				panic(dogma.UnexpectedMessage)
			}

			xtesting.ExpectPanicMatching(t, func() {
				_, _ = f.ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					f.event,
				)
			}, func(x panicx.UnexpectedMessage) {
				xtesting.Expect(t, "unexpected handler", x.Handler, f.cfg)
				xtesting.Expect(t, "unexpected interface", x.Interface, "ProcessMessageHandler")
				xtesting.Expect(t, "unexpected method", x.Method, "HandleEvent")
				xtesting.Expect(t, "unexpected message", x.Message, f.event.Message)
			})
		})

		t.Run("provides more context to UnexpectedMessage panics from HandleDeadline", func(t *testing.T) {
			f := newControllerTestFixture()
			f.handler.HandleEventFunc = func(
				context.Context,
				*ProcessRootStub,
				dogma.ProcessEventScope[*ProcessRootStub],
				dogma.Event,
			) error {
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

			f.handler.HandleDeadlineFunc = func(
				context.Context,
				*ProcessRootStub,
				dogma.ProcessDeadlineScope[*ProcessRootStub],
				dogma.Deadline,
			) error {
				panic(dogma.UnexpectedMessage)
			}

			xtesting.ExpectPanicMatching(t, func() {
				_, _ = f.ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					f.deadline,
				)
			}, func(x panicx.UnexpectedMessage) {
				xtesting.Expect(t, "unexpected handler", x.Handler, f.cfg)
				xtesting.Expect(t, "unexpected interface", x.Interface, "ProcessMessageHandler")
				xtesting.Expect(t, "unexpected method", x.Method, "HandleDeadline")
				xtesting.Expect(t, "unexpected message", x.Message, f.deadline.Message)
			})
		})

		t.Run("panics if MarshalBinary() fails", func(t *testing.T) {
			f := newControllerTestFixture()
			f.handler.HandleEventFunc = func(
				_ context.Context,
				_ *ProcessRootStub,
				s dogma.ProcessEventScope[*ProcessRootStub],
				_ dogma.Event,
			) error {
				s.Mutate(func(r *ProcessRootStub) {
					r.MarshalBinaryFunc = func() ([]byte, error) {
						return nil, errors.New("<marshal error>")
					}
				})
				return nil
			}

			xtesting.ExpectPanic(
				t,
				"the '<name>' process message handler behaved unexpectedly in *stubs.ProcessRootStub.MarshalBinary(): unable to marshal the process root: <marshal error>",
				func() {
					_, _ = f.ctrl.Handle(
						context.Background(),
						fact.Ignore,
						time.Now(),
						f.event,
					)
				},
			)
		})

		t.Run("panics if UnmarshalBinary fails", func(t *testing.T) {
			f := newControllerTestFixture()
			f.handler.HandleEventFunc = func(
				_ context.Context,
				_ *ProcessRootStub,
				s dogma.ProcessEventScope[*ProcessRootStub],
				_ dogma.Event,
			) error {
				s.Mutate(func(*ProcessRootStub) {})
				return nil
			}

			// First Handle: creates instance, mutates, marshal succeeds.
			_, err := f.ctrl.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				f.event,
			)
			if err != nil {
				t.Fatal(err)
			}

			// Second Handle: loads instance, unmarshal fails.
			f.handler.NewFunc = func() *ProcessRootStub {
				return &ProcessRootStub{
					UnmarshalBinaryFunc: func([]byte) error {
						return errors.New("<unmarshal error>")
					},
				}
			}

			xtesting.ExpectPanic(
				t,
				"the '<name>' process message handler behaved unexpectedly in *stubs.ProcessRootStub.UnmarshalBinary(): unable to unmarshal the process root: <unmarshal error>",
				func() {
					_, _ = f.ctrl.Handle(
						context.Background(),
						fact.Ignore,
						time.Now(),
						f.event,
					)
				},
			)
		})

		t.Run("calls UnmarshalBinary when MarshalBinary returns nil", func(t *testing.T) {
			f := newControllerTestFixture()
			f.handler.HandleEventFunc = func(
				_ context.Context,
				_ *ProcessRootStub,
				s dogma.ProcessEventScope[*ProcessRootStub],
				_ dogma.Event,
			) error {
				s.Mutate(func(r *ProcessRootStub) {
					r.MarshalBinaryFunc = func() ([]byte, error) {
						return nil, nil
					}
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

			called := false
			f.handler.NewFunc = func() *ProcessRootStub {
				return &ProcessRootStub{
					UnmarshalBinaryFunc: func([]byte) error {
						called = true
						return nil
					},
				}
			}

			_, err = f.ctrl.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				f.event,
			)
			if err != nil {
				t.Fatal(err)
			}

			xtesting.Expect(t, "expected UnmarshalBinary to be called", called, true)
		})

		t.Run("calls UnmarshalBinary when MarshalBinary returns an empty slice", func(t *testing.T) {
			f := newControllerTestFixture()
			f.handler.HandleEventFunc = func(
				_ context.Context,
				_ *ProcessRootStub,
				s dogma.ProcessEventScope[*ProcessRootStub],
				_ dogma.Event,
			) error {
				s.Mutate(func(r *ProcessRootStub) {
					r.MarshalBinaryFunc = func() ([]byte, error) {
						return []byte{}, nil
					}
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

			called := false
			f.handler.NewFunc = func() *ProcessRootStub {
				return &ProcessRootStub{
					UnmarshalBinaryFunc: func([]byte) error {
						called = true
						return nil
					},
				}
			}

			_, err = f.ctrl.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				f.event,
			)
			if err != nil {
				t.Fatal(err)
			}

			xtesting.Expect(t, "expected UnmarshalBinary to be called", called, true)
		})
	})

	t.Run("Reset", func(t *testing.T) {
		f := newControllerTestFixture()
		f.handler.HandleEventFunc = func(
			context.Context,
			*ProcessRootStub,
			dogma.ProcessEventScope[*ProcessRootStub],
			dogma.Event,
		) error {
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

		f.messageIDs.Reset()
		f.ctrl.Reset()

		buf := &fact.Buffer{}
		_, err = f.ctrl.Handle(
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
					InstanceID: "<instance-A1>",
					Envelope:   f.event,
				},
				fact.ProcessInstanceBegun{
					Handler:    f.cfg,
					InstanceID: "<instance-A1>",
					Root:       &ProcessRootStub{},
					Envelope:   f.event,
				},
			},
		)
	})
}

type controllerTestFixture struct {
	messageIDs envelope.MessageIDGenerator
	handler    *ProcessMessageHandlerStub[*ProcessRootStub]
	cfg        *config.Process
	ctrl       *Controller
	event      *envelope.Envelope
	deadline   *envelope.Envelope
}

func newControllerTestFixture() *controllerTestFixture {
	event := envelope.NewEvent(
		"1000",
		EventA1,
		time.Now(),
	)

	handler := &ProcessMessageHandlerStub[*ProcessRootStub]{
		ConfigureFunc: func(c dogma.ProcessConfigurer) {
			c.Identity("<name>", "7db72921-b805-4db5-8287-0af94a768643")
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
			switch x := m.(type) {
			case *EventStub[TypeA]:
				return fmt.Sprintf("<instance-%s>", x.Content), true, nil
			default:
				panic(dogma.UnexpectedMessage)
			}
		},
	}

	cfg := runtimeconfig.FromProcess(handler)
	deadline := event.NewDeadline(
		"2000",
		DeadlineA1,
		time.Now(),
		time.Now().Add(10*time.Second),
		envelope.Origin{
			Handler:     cfg,
			HandlerType: config.ProcessHandlerType,
			InstanceID:  "<instance-A1>",
		},
	)

	f := &controllerTestFixture{
		handler:  handler,
		cfg:      cfg,
		event:    event,
		deadline: deadline,
	}

	f.ctrl = &Controller{
		Config:     cfg,
		MessageIDs: &f.messageIDs,
	}

	f.messageIDs.Reset()

	return f
}
