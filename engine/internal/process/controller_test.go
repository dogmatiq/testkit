package process_test

import (
	"context"
	"errors"
	"fmt"
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
	"github.com/dogmatiq/testkit/location"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestController(t *testing.T) {
	t.Run("HandlerConfig", func(t *testing.T) {
		env := newControllerTestFixture()

		if got := env.ctrl.HandlerConfig(); got != env.cfg {
			t.Fatalf("unexpected handler config: got %p, want %p", got, env.cfg)
		}
	})

	t.Run("Tick", func(t *testing.T) {
		setup := func(t *testing.T) (*controllerTestFixture, time.Time, time.Time, time.Time, time.Time) {
			t.Helper()

			env := newControllerTestFixture()
			createdTime := time.Now()
			t1Time := createdTime.Add(1 * time.Hour)
			t2Time := createdTime.Add(2 * time.Hour)
			t3Time := createdTime.Add(3 * time.Hour)

			env.handler.HandleEventFunc = func(
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

			_, err := env.ctrl.Handle(
				context.Background(),
				fact.Ignore,
				createdTime,
				env.event,
			)
			expectNoError(t, err)

			env.messageIDs.Reset()

			return env, createdTime, t1Time, t2Time, t3Time
		}

		t.Run("returns deadlines that are ready to be handled", func(t *testing.T) {
			env, createdTime, t1Time, t2Time, _ := setup(t)

			deadlines, err := env.ctrl.Tick(
				context.Background(),
				fact.Ignore,
				t2Time,
			)
			expectNoError(t, err)

			expectEnvelopeSet(
				t,
				deadlines,
				[]*envelope.Envelope{
					env.event.NewDeadline(
						"3",
						DeadlineA1,
						createdTime,
						t1Time,
						envelope.Origin{
							Handler:     env.cfg,
							HandlerType: config.ProcessHandlerType,
							InstanceID:  "<instance-A1>",
						},
					),
					env.event.NewDeadline(
						"2",
						DeadlineA2,
						createdTime,
						t2Time,
						envelope.Origin{
							Handler:     env.cfg,
							HandlerType: config.ProcessHandlerType,
							InstanceID:  "<instance-A1>",
						},
					),
				},
			)
		})

		t.Run("does not return the same deadlines multiple times", func(t *testing.T) {
			env, _, _, t2Time, _ := setup(t)

			deadlines, err := env.ctrl.Tick(
				context.Background(),
				fact.Ignore,
				t2Time,
			)
			expectNoError(t, err)

			if got, want := len(deadlines), 2; got != want {
				t.Fatalf("unexpected deadline count: got %d, want %d", got, want)
			}

			deadlines, err = env.ctrl.Tick(
				context.Background(),
				fact.Ignore,
				t2Time,
			)
			expectNoError(t, err)

			if len(deadlines) != 0 {
				t.Fatalf("unexpected deadlines: got %d, want 0", len(deadlines))
			}
		})

		t.Run("does not return deadlines for instances that have been ended", func(t *testing.T) {
			env, createdTime, t1Time, t2Time, _ := setup(t)

			secondInstanceEvent := envelope.NewEvent(
				"3000",
				EventA2,
				time.Now(),
			)

			_, err := env.ctrl.Handle(
				context.Background(),
				fact.Ignore,
				createdTime,
				secondInstanceEvent,
			)
			expectNoError(t, err)

			env.handler.HandleEventFunc = func(
				_ context.Context,
				_ *ProcessRootStub,
				s dogma.ProcessEventScope[*ProcessRootStub],
				_ dogma.Event,
			) error {
				s.End()
				return nil
			}

			_, err = env.ctrl.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				env.event,
			)
			expectNoError(t, err)

			deadlines, err := env.ctrl.Tick(
				context.Background(),
				fact.Ignore,
				t2Time,
			)
			expectNoError(t, err)

			expectEnvelopeSet(
				t,
				deadlines,
				[]*envelope.Envelope{
					secondInstanceEvent.NewDeadline(
						"3",
						DeadlineA1,
						createdTime,
						t1Time,
						envelope.Origin{
							Handler:     env.cfg,
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
							Handler:     env.cfg,
							HandlerType: config.ProcessHandlerType,
							InstanceID:  "<instance-A2>",
						},
					),
				},
			)
		})
	})

	t.Run("Handle", func(t *testing.T) {
		t.Run("handling an event", func(t *testing.T) {
			t.Run("forwards the message to the handler", func(t *testing.T) {
				env := newControllerTestFixture()
				called := false

				env.handler.HandleEventFunc = func(
					_ context.Context,
					_ *ProcessRootStub,
					_ dogma.ProcessEventScope[*ProcessRootStub],
					m dogma.Event,
				) error {
					called = true
					xtesting.Expect(t, "unexpected event", m, EventA1)
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

			t.Run("propagates handler errors", func(t *testing.T) {
				env := newControllerTestFixture()
				expected := errors.New("<error>")

				env.handler.HandleEventFunc = func(
					_ context.Context,
					_ *ProcessRootStub,
					_ dogma.ProcessEventScope[*ProcessRootStub],
					_ dogma.Event,
				) error {
					return expected
				}

				_, err := env.ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					env.event,
				)

				xtesting.Expect(t, "unexpected error", err, expected)
			})

			t.Run("returns both commands and deadlines", func(t *testing.T) {
				env := newControllerTestFixture()
				now := time.Now()

				env.handler.HandleEventFunc = func(
					_ context.Context,
					_ *ProcessRootStub,
					s dogma.ProcessEventScope[*ProcessRootStub],
					_ dogma.Event,
				) error {
					s.ExecuteCommand(CommandA1)
					s.ScheduleDeadline(DeadlineA1, now)
					return nil
				}

				envelopes, err := env.ctrl.Handle(
					context.Background(),
					fact.Ignore,
					now,
					env.event,
				)
				expectNoError(t, err)

				expectEnvelopeSet(
					t,
					envelopes,
					[]*envelope.Envelope{
						env.event.NewCommand(
							"1",
							CommandA1,
							now,
							envelope.Origin{
								Handler:     env.cfg,
								HandlerType: config.ProcessHandlerType,
								InstanceID:  "<instance-A1>",
							},
						),
						env.event.NewDeadline(
							"2",
							DeadlineA1,
							now,
							now,
							envelope.Origin{
								Handler:     env.cfg,
								HandlerType: config.ProcessHandlerType,
								InstanceID:  "<instance-A1>",
							},
						),
					},
				)
			})

			t.Run("returns deadlines scheduled in the past", func(t *testing.T) {
				env := newControllerTestFixture()
				now := time.Now()

				env.handler.HandleEventFunc = func(
					_ context.Context,
					_ *ProcessRootStub,
					s dogma.ProcessEventScope[*ProcessRootStub],
					_ dogma.Event,
				) error {
					s.ScheduleDeadline(DeadlineA1, now.Add(-1))
					return nil
				}

				envelopes, err := env.ctrl.Handle(
					context.Background(),
					fact.Ignore,
					now,
					env.event,
				)
				expectNoError(t, err)

				if got, want := len(envelopes), 1; got != want {
					t.Fatalf("unexpected envelope count: got %d, want %d", got, want)
				}
			})

			t.Run("does not return deadlines scheduled in the future", func(t *testing.T) {
				env := newControllerTestFixture()
				now := time.Now()

				env.handler.HandleEventFunc = func(
					_ context.Context,
					_ *ProcessRootStub,
					s dogma.ProcessEventScope[*ProcessRootStub],
					_ dogma.Event,
				) error {
					s.ScheduleDeadline(DeadlineA1, now.Add(1))
					return nil
				}

				envelopes, err := env.ctrl.Handle(
					context.Background(),
					fact.Ignore,
					now,
					env.event,
				)
				expectNoError(t, err)

				if len(envelopes) != 0 {
					t.Fatalf("unexpected envelopes: got %d, want 0", len(envelopes))
				}
			})

			t.Run("when the event is not routed to an instance", func(t *testing.T) {
				t.Run("does not forward the message to the handler", func(t *testing.T) {
					env := newControllerTestFixture()
					env.handler.RouteEventToInstanceFunc = func(
						context.Context,
						dogma.Event,
					) (string, bool, error) {
						return "", false, nil
					}

					env.handler.HandleEventFunc = func(
						context.Context,
						*ProcessRootStub,
						dogma.ProcessEventScope[*ProcessRootStub],
						dogma.Event,
					) error {
						t.Fatal("unexpected call to HandleEvent()")
						return nil
					}

					_, err := env.ctrl.Handle(
						context.Background(),
						fact.Ignore,
						time.Now(),
						env.event,
					)
					expectNoError(t, err)
				})

				t.Run("records a fact", func(t *testing.T) {
					env := newControllerTestFixture()
					env.handler.RouteEventToInstanceFunc = func(
						context.Context,
						dogma.Event,
					) (string, bool, error) {
						return "", false, nil
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
							fact.ProcessEventIgnored{
								Handler:  env.cfg,
								Envelope: env.event,
							},
						},
					)
				})
			})

			t.Run("when the instance has ended", func(t *testing.T) {
				setup := func(t *testing.T) *controllerTestFixture {
					t.Helper()

					env := newControllerTestFixture()
					env.handler.HandleEventFunc = func(
						_ context.Context,
						_ *ProcessRootStub,
						s dogma.ProcessEventScope[*ProcessRootStub],
						_ dogma.Event,
					) error {
						s.End()
						return nil
					}

					_, err := env.ctrl.Handle(
						context.Background(),
						fact.Ignore,
						time.Now(),
						env.event,
					)
					expectNoError(t, err)

					env.messageIDs.Reset()
					return env
				}

				t.Run("does not forward the message to the handler", func(t *testing.T) {
					env := setup(t)
					env.handler.HandleEventFunc = func(
						context.Context,
						*ProcessRootStub,
						dogma.ProcessEventScope[*ProcessRootStub],
						dogma.Event,
					) error {
						t.Fatal("unexpected call to HandleEvent()")
						return nil
					}

					_, err := env.ctrl.Handle(
						context.Background(),
						fact.Ignore,
						time.Now(),
						env.event,
					)
					expectNoError(t, err)
				})

				t.Run("records a fact", func(t *testing.T) {
					env := setup(t)
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
							fact.ProcessEventRoutedToEndedInstance{
								Handler:    env.cfg,
								InstanceID: "<instance-A1>",
								Envelope:   env.event,
							},
						},
					)
				})
			})
		})

		t.Run("handling a deadline", func(t *testing.T) {
			setup := func(t *testing.T) *controllerTestFixture {
				t.Helper()

				env := newControllerTestFixture()
				env.handler.HandleEventFunc = func(
					context.Context,
					*ProcessRootStub,
					dogma.ProcessEventScope[*ProcessRootStub],
					dogma.Event,
				) error {
					return nil
				}

				_, err := env.ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					env.event,
				)
				expectNoError(t, err)

				env.messageIDs.Reset()
				return env
			}

			t.Run("forwards the message to the handler", func(t *testing.T) {
				env := setup(t)
				called := false

				env.handler.HandleDeadlineFunc = func(
					_ context.Context,
					_ *ProcessRootStub,
					_ dogma.ProcessDeadlineScope[*ProcessRootStub],
					m dogma.Deadline,
				) error {
					called = true
					xtesting.Expect(t, "unexpected deadline", m, DeadlineA1)
					return nil
				}

				_, err := env.ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					env.deadline,
				)
				expectNoError(t, err)

				if !called {
					t.Fatal("expected HandleDeadline() to be called")
				}
			})

			t.Run("propagates handler errors", func(t *testing.T) {
				env := setup(t)
				expected := errors.New("<error>")

				env.handler.HandleDeadlineFunc = func(
					context.Context,
					*ProcessRootStub,
					dogma.ProcessDeadlineScope[*ProcessRootStub],
					dogma.Deadline,
				) error {
					return expected
				}

				_, err := env.ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					env.deadline,
				)

				xtesting.Expect(t, "unexpected error", err, expected)
			})

			t.Run("returns both commands and deadlines", func(t *testing.T) {
				env := setup(t)
				now := time.Now()

				env.handler.HandleDeadlineFunc = func(
					_ context.Context,
					_ *ProcessRootStub,
					s dogma.ProcessDeadlineScope[*ProcessRootStub],
					_ dogma.Deadline,
				) error {
					s.ExecuteCommand(CommandA1)
					s.ScheduleDeadline(DeadlineA1, now)
					return nil
				}

				envelopes, err := env.ctrl.Handle(
					context.Background(),
					fact.Ignore,
					now,
					env.deadline,
				)
				expectNoError(t, err)

				expectEnvelopeSet(
					t,
					envelopes,
					[]*envelope.Envelope{
						env.deadline.NewCommand(
							"1",
							CommandA1,
							now,
							envelope.Origin{
								Handler:     env.cfg,
								HandlerType: config.ProcessHandlerType,
								InstanceID:  "<instance-A1>",
							},
						),
						env.deadline.NewDeadline(
							"2",
							DeadlineA1,
							now,
							now,
							envelope.Origin{
								Handler:     env.cfg,
								HandlerType: config.ProcessHandlerType,
								InstanceID:  "<instance-A1>",
							},
						),
					},
				)
			})

			t.Run("returns deadlines scheduled in the past", func(t *testing.T) {
				env := setup(t)
				now := time.Now()

				env.handler.HandleDeadlineFunc = func(
					_ context.Context,
					_ *ProcessRootStub,
					s dogma.ProcessDeadlineScope[*ProcessRootStub],
					_ dogma.Deadline,
				) error {
					s.ScheduleDeadline(DeadlineA2, now.Add(-1))
					return nil
				}

				envelopes, err := env.ctrl.Handle(
					context.Background(),
					fact.Ignore,
					now,
					env.deadline,
				)
				expectNoError(t, err)

				if got, want := len(envelopes), 1; got != want {
					t.Fatalf("unexpected envelope count: got %d, want %d", got, want)
				}
			})

			t.Run("does not return deadlines scheduled in the future", func(t *testing.T) {
				env := setup(t)
				now := time.Now()

				env.handler.HandleDeadlineFunc = func(
					_ context.Context,
					_ *ProcessRootStub,
					s dogma.ProcessDeadlineScope[*ProcessRootStub],
					_ dogma.Deadline,
				) error {
					s.ScheduleDeadline(DeadlineA2, now.Add(1))
					return nil
				}

				envelopes, err := env.ctrl.Handle(
					context.Background(),
					fact.Ignore,
					now,
					env.deadline,
				)
				expectNoError(t, err)

				if len(envelopes) != 0 {
					t.Fatalf("unexpected envelopes: got %d, want 0", len(envelopes))
				}
			})

			t.Run("when the instance that created the deadline has ended", func(t *testing.T) {
				setupEnded := func(t *testing.T) *controllerTestFixture {
					t.Helper()

					env := setup(t)
					env.handler.HandleEventFunc = func(
						_ context.Context,
						_ *ProcessRootStub,
						s dogma.ProcessEventScope[*ProcessRootStub],
						_ dogma.Event,
					) error {
						s.End()
						return nil
					}

					_, err := env.ctrl.Handle(
						context.Background(),
						fact.Ignore,
						time.Now(),
						env.event,
					)
					expectNoError(t, err)

					env.messageIDs.Reset()
					return env
				}

				t.Run("does not forward the message to the handler", func(t *testing.T) {
					env := setupEnded(t)
					env.handler.HandleDeadlineFunc = func(
						context.Context,
						*ProcessRootStub,
						dogma.ProcessDeadlineScope[*ProcessRootStub],
						dogma.Deadline,
					) error {
						t.Fatal("unexpected call to HandleDeadline()")
						return nil
					}

					_, err := env.ctrl.Handle(
						context.Background(),
						fact.Ignore,
						time.Now(),
						env.deadline,
					)
					expectNoError(t, err)
				})

				t.Run("records a fact", func(t *testing.T) {
					env := setupEnded(t)
					buf := &fact.Buffer{}

					_, err := env.ctrl.Handle(
						context.Background(),
						buf,
						time.Now(),
						env.deadline,
					)
					expectNoError(t, err)

					expectFacts(
						t,
						buf.Facts(),
						[]fact.Fact{
							fact.ProcessDeadlineRoutedToEndedInstance{
								Handler:    env.cfg,
								InstanceID: "<instance-A1>",
								Envelope:   env.deadline,
							},
						},
					)
				})
			})
		})

		t.Run("propagates routing errors", func(t *testing.T) {
			env := newControllerTestFixture()
			expected := errors.New("<error>")

			env.handler.RouteEventToInstanceFunc = func(
				context.Context,
				dogma.Event,
			) (string, bool, error) {
				return "<instance>", true, expected
			}

			_, err := env.ctrl.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				env.event,
			)

			xtesting.Expect(t, "unexpected error", err, expected)
		})

		t.Run("panics when the handler routes to an empty instance ID", func(t *testing.T) {
			env := newControllerTestFixture()

			env.handler.RouteEventToInstanceFunc = func(
				context.Context,
				dogma.Event,
			) (string, bool, error) {
				return "", true, nil
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
					Method:         "RouteEventToInstance",
					Implementation: env.cfg.Implementation(),
					Message:        env.event.Message,
					Description:    "routed an event of type *stubs.EventStub[TypeA] to an empty ID",
				},
				"/stubs/process.go",
			)
		})

		t.Run("when the instance does not exist", func(t *testing.T) {
			t.Run("records facts", func(t *testing.T) {
				env := newControllerTestFixture()
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
							InstanceID: "<instance-A1>",
							Envelope:   env.event,
						},
						fact.ProcessInstanceBegun{
							Handler:    env.cfg,
							InstanceID: "<instance-A1>",
							Root:       &ProcessRootStub{},
							Envelope:   env.event,
						},
					},
				)
			})

			t.Run("panics if New returns nil", func(t *testing.T) {
				env := newControllerTestFixture()
				env.handler.NewFunc = func() *ProcessRootStub {
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
						Method:         "New",
						Implementation: env.cfg.Implementation(),
						Message:        env.event.Message,
						Description:    "returned a nil process root",
					},
					"/stubs/process.go",
				)
			})
		})

		t.Run("when the instance exists", func(t *testing.T) {
			setup := func(t *testing.T) *controllerTestFixture {
				t.Helper()

				env := newControllerTestFixture()
				env.handler.HandleEventFunc = func(
					_ context.Context,
					_ *ProcessRootStub,
					_ dogma.ProcessEventScope[*ProcessRootStub],
					_ dogma.Event,
				) error {
					return nil
				}

				_, err := env.ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					env.event,
				)
				expectNoError(t, err)

				env.messageIDs.Reset()
				return env
			}

			t.Run("records a fact", func(t *testing.T) {
				env := setup(t)
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
						fact.ProcessInstanceLoaded{
							Handler:    env.cfg,
							InstanceID: "<instance-A1>",
							Root:       &ProcessRootStub{},
							Envelope:   env.event,
						},
					},
				)
			})

			t.Run("provides the root with state from the prior Handle() call", func(t *testing.T) {
				env := newControllerTestFixture()

				env.handler.HandleEventFunc = func(
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

				_, err := env.ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					env.event,
				)
				expectNoError(t, err)

				var got string
				env.handler.HandleEventFunc = func(
					_ context.Context,
					r *ProcessRootStub,
					_ dogma.ProcessEventScope[*ProcessRootStub],
					_ dogma.Event,
				) error {
					got = r.Value.(string)
					return nil
				}

				_, err = env.ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					env.event,
				)
				expectNoError(t, err)

				if got != "<mutated>" {
					t.Fatalf("expected root state to persist, got %q", got)
				}
			})

			t.Run("panics if New() returns nil", func(t *testing.T) {
				env := setup(t)
				env.handler.NewFunc = func() *ProcessRootStub {
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
						Method:         "New",
						Implementation: env.cfg.Implementation(),
						Message:        env.event.Message,
						Description:    "returned a nil process root",
					},
					"/stubs/process.go",
				)
			})
		})

		t.Run("provides more context to UnexpectedMessage panics from RouteEventToInstance", func(t *testing.T) {
			env := newControllerTestFixture()
			env.handler.RouteEventToInstanceFunc = func(
				context.Context,
				dogma.Event,
			) (string, bool, error) {
				panic(dogma.UnexpectedMessage)
			}

			expectUnexpectedMessage(
				t,
				func() {
					_, _ = env.ctrl.Handle(
						context.Background(),
						fact.Ignore,
						time.Now(),
						env.event,
					)
				},
				env.cfg,
				"RouteEventToInstance",
				env.event.Message,
			)
		})

		t.Run("provides more context to UnexpectedMessage panics from HandleEvent", func(t *testing.T) {
			env := newControllerTestFixture()
			env.handler.HandleEventFunc = func(
				context.Context,
				*ProcessRootStub,
				dogma.ProcessEventScope[*ProcessRootStub],
				dogma.Event,
			) error {
				panic(dogma.UnexpectedMessage)
			}

			expectUnexpectedMessage(
				t,
				func() {
					_, _ = env.ctrl.Handle(
						context.Background(),
						fact.Ignore,
						time.Now(),
						env.event,
					)
				},
				env.cfg,
				"HandleEvent",
				env.event.Message,
			)
		})

		t.Run("provides more context to UnexpectedMessage panics from HandleDeadline", func(t *testing.T) {
			env := newControllerTestFixture()
			env.handler.HandleEventFunc = func(
				context.Context,
				*ProcessRootStub,
				dogma.ProcessEventScope[*ProcessRootStub],
				dogma.Event,
			) error {
				return nil
			}

			_, err := env.ctrl.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				env.event,
			)
			expectNoError(t, err)

			env.handler.HandleDeadlineFunc = func(
				context.Context,
				*ProcessRootStub,
				dogma.ProcessDeadlineScope[*ProcessRootStub],
				dogma.Deadline,
			) error {
				panic(dogma.UnexpectedMessage)
			}

			expectUnexpectedMessage(
				t,
				func() {
					_, _ = env.ctrl.Handle(
						context.Background(),
						fact.Ignore,
						time.Now(),
						env.deadline,
					)
				},
				env.cfg,
				"HandleDeadline",
				env.deadline.Message,
			)
		})

		t.Run("panics if MarshalBinary() fails", func(t *testing.T) {
			env := newControllerTestFixture()
			env.handler.HandleEventFunc = func(
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
					_, _ = env.ctrl.Handle(
						context.Background(),
						fact.Ignore,
						time.Now(),
						env.event,
					)
				},
			)
		})

		t.Run("panics if UnmarshalBinary fails", func(t *testing.T) {
			env := newControllerTestFixture()
			env.handler.HandleEventFunc = func(
				_ context.Context,
				_ *ProcessRootStub,
				s dogma.ProcessEventScope[*ProcessRootStub],
				_ dogma.Event,
			) error {
				s.Mutate(func(*ProcessRootStub) {})
				return nil
			}

			// First Handle: creates instance, mutates, marshal succeeds.
			_, err := env.ctrl.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				env.event,
			)
			expectNoError(t, err)

			// Second Handle: loads instance, unmarshal fails.
			env.handler.NewFunc = func() *ProcessRootStub {
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
					_, _ = env.ctrl.Handle(
						context.Background(),
						fact.Ignore,
						time.Now(),
						env.event,
					)
				},
			)
		})

		t.Run("calls UnmarshalBinary when MarshalBinary returns nil", func(t *testing.T) {
			env := newControllerTestFixture()
			env.handler.HandleEventFunc = func(
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

			_, err := env.ctrl.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				env.event,
			)
			expectNoError(t, err)

			called := false
			env.handler.NewFunc = func() *ProcessRootStub {
				return &ProcessRootStub{
					UnmarshalBinaryFunc: func([]byte) error {
						called = true
						return nil
					},
				}
			}

			_, err = env.ctrl.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				env.event,
			)
			expectNoError(t, err)

			xtesting.Expect(t, "expected UnmarshalBinary to be called", called, true)
		})

		t.Run("calls UnmarshalBinary when MarshalBinary returns an empty slice", func(t *testing.T) {
			env := newControllerTestFixture()
			env.handler.HandleEventFunc = func(
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

			_, err := env.ctrl.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				env.event,
			)
			expectNoError(t, err)

			called := false
			env.handler.NewFunc = func() *ProcessRootStub {
				return &ProcessRootStub{
					UnmarshalBinaryFunc: func([]byte) error {
						called = true
						return nil
					},
				}
			}

			_, err = env.ctrl.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				env.event,
			)
			expectNoError(t, err)

			xtesting.Expect(t, "expected UnmarshalBinary to be called", called, true)
		})
	})

	t.Run("Reset", func(t *testing.T) {
		env := newControllerTestFixture()
		env.handler.HandleEventFunc = func(
			context.Context,
			*ProcessRootStub,
			dogma.ProcessEventScope[*ProcessRootStub],
			dogma.Event,
		) error {
			return nil
		}

		_, err := env.ctrl.Handle(
			context.Background(),
			fact.Ignore,
			time.Now(),
			env.event,
		)
		expectNoError(t, err)

		env.messageIDs.Reset()
		env.ctrl.Reset()

		buf := &fact.Buffer{}
		_, err = env.ctrl.Handle(
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
					InstanceID: "<instance-A1>",
					Envelope:   env.event,
				},
				fact.ProcessInstanceBegun{
					Handler:    env.cfg,
					InstanceID: "<instance-A1>",
					Root:       &ProcessRootStub{},
					Envelope:   env.event,
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

	env := &controllerTestFixture{
		handler:  handler,
		cfg:      cfg,
		event:    event,
		deadline: deadline,
	}

	env.ctrl = &Controller{
		Config:     cfg,
		MessageIDs: &env.messageIDs,
	}

	env.messageIDs.Reset()

	return env
}

func expectNoError(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func expectEnvelopeSet(t *testing.T, got, want []*envelope.Envelope) {
	t.Helper()

	xtesting.Expect(
		t,
		"unexpected envelopes",
		got,
		want,
		cmpopts.SortSlices(
			func(a, b *envelope.Envelope) bool {
				return a.MessageID < b.MessageID
			},
		),
	)
}

func expectFacts(t *testing.T, got, want []fact.Fact) {
	t.Helper()
	xtesting.Expect(t, "unexpected facts", got, want)
}

func expectUnexpectedBehavior(
	t *testing.T,
	fn func(),
	want panicx.UnexpectedBehavior,
	wantFileSuffix string,
) {
	t.Helper()

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panicx.UnexpectedBehavior panic, got nil")
		}

		x, ok := r.(panicx.UnexpectedBehavior)
		if !ok {
			t.Fatalf("expected panicx.UnexpectedBehavior panic, got %T", r)
		}

		loc := x.Location
		x.Location = location.Location{}
		want.Location = location.Location{}

		xtesting.Expect(t, "unexpected panic", x, want)

		if loc.Func == "" {
			t.Fatal("unexpected empty panic location func")
		}

		if !strings.HasSuffix(loc.File, wantFileSuffix) {
			t.Fatalf("unexpected panic location file: %s", loc.File)
		}

		if loc.Line == 0 {
			t.Fatal("unexpected zero panic location line")
		}
	}()

	fn()
}

func expectUnexpectedMessage(
	t *testing.T,
	fn func(),
	handler *config.Process,
	method string,
	message dogma.Message,
) {
	t.Helper()

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panicx.UnexpectedMessage panic, got nil")
		}

		x, ok := r.(panicx.UnexpectedMessage)
		if !ok {
			t.Fatalf("expected panicx.UnexpectedMessage panic, got %T", r)
		}

		xtesting.Expect(t, "unexpected handler", x.Handler, handler)
		xtesting.Expect(t, "unexpected interface", x.Interface, "ProcessMessageHandler")
		xtesting.Expect(t, "unexpected method", x.Method, method)
		xtesting.Expect(t, "unexpected implementation", x.Implementation, handler.Implementation())
		xtesting.Expect(t, "unexpected message", x.Message, message)
	}()

	fn()
}
