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
		env := newProcessControllerTestEnv()

		if got := env.ctrl.HandlerConfig(); got != env.cfg {
			t.Fatalf("unexpected handler config: got %p, want %p", got, env.cfg)
		}
	})

	t.Run("Tick", func(t *testing.T) {
		setup := func(t *testing.T) (*processControllerTestEnv, time.Time, time.Time, time.Time, time.Time) {
			t.Helper()

			env := newProcessControllerTestEnv()
			createdTime := time.Now()
			t1Time := createdTime.Add(1 * time.Hour)
			t2Time := createdTime.Add(2 * time.Hour)
			t3Time := createdTime.Add(3 * time.Hour)

			env.handler.HandleEventFunc = func(
				_ context.Context,
				_ dogma.ProcessRoot,
				s dogma.ProcessEventScope,
				_ dogma.Event,
			) error {
				s.ScheduleTimeout(TimeoutA3, t3Time)
				s.ScheduleTimeout(TimeoutA2, t2Time)
				s.ScheduleTimeout(TimeoutA1, t1Time)
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

		t.Run("returns timeouts that are ready to be handled", func(t *testing.T) {
			env, createdTime, t1Time, t2Time, _ := setup(t)

			timeouts, err := env.ctrl.Tick(
				context.Background(),
				fact.Ignore,
				t2Time,
			)
			expectNoError(t, err)

			expectEnvelopeSet(
				t,
				timeouts,
				[]*envelope.Envelope{
					env.event.NewTimeout(
						"3",
						TimeoutA1,
						createdTime,
						t1Time,
						envelope.Origin{
							Handler:     env.cfg,
							HandlerType: config.ProcessHandlerType,
							InstanceID:  "<instance-A1>",
						},
					),
					env.event.NewTimeout(
						"2",
						TimeoutA2,
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

		t.Run("does not return the same timeouts multiple times", func(t *testing.T) {
			env, _, _, t2Time, _ := setup(t)

			timeouts, err := env.ctrl.Tick(
				context.Background(),
				fact.Ignore,
				t2Time,
			)
			expectNoError(t, err)

			if got, want := len(timeouts), 2; got != want {
				t.Fatalf("unexpected timeout count: got %d, want %d", got, want)
			}

			timeouts, err = env.ctrl.Tick(
				context.Background(),
				fact.Ignore,
				t2Time,
			)
			expectNoError(t, err)

			if len(timeouts) != 0 {
				t.Fatalf("unexpected timeouts: got %d, want 0", len(timeouts))
			}
		})

		t.Run("does not return timeouts for instances that have been ended", func(t *testing.T) {
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
				_ dogma.ProcessRoot,
				s dogma.ProcessEventScope,
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

			timeouts, err := env.ctrl.Tick(
				context.Background(),
				fact.Ignore,
				t2Time,
			)
			expectNoError(t, err)

			expectEnvelopeSet(
				t,
				timeouts,
				[]*envelope.Envelope{
					secondInstanceEvent.NewTimeout(
						"3",
						TimeoutA1,
						createdTime,
						t1Time,
						envelope.Origin{
							Handler:     env.cfg,
							HandlerType: config.ProcessHandlerType,
							InstanceID:  "<instance-A2>",
						},
					),
					secondInstanceEvent.NewTimeout(
						"2",
						TimeoutA2,
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
				env := newProcessControllerTestEnv()
				called := false

				env.handler.HandleEventFunc = func(
					_ context.Context,
					_ dogma.ProcessRoot,
					_ dogma.ProcessEventScope,
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
				env := newProcessControllerTestEnv()
				expected := errors.New("<error>")

				env.handler.HandleEventFunc = func(
					_ context.Context,
					_ dogma.ProcessRoot,
					_ dogma.ProcessEventScope,
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

			t.Run("returns both commands and timeouts", func(t *testing.T) {
				env := newProcessControllerTestEnv()
				now := time.Now()

				env.handler.HandleEventFunc = func(
					_ context.Context,
					_ dogma.ProcessRoot,
					s dogma.ProcessEventScope,
					_ dogma.Event,
				) error {
					s.ExecuteCommand(CommandA1)
					s.ScheduleTimeout(TimeoutA1, now)
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
						env.event.NewTimeout(
							"2",
							TimeoutA1,
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

			t.Run("returns timeouts scheduled in the past", func(t *testing.T) {
				env := newProcessControllerTestEnv()
				now := time.Now()

				env.handler.HandleEventFunc = func(
					_ context.Context,
					_ dogma.ProcessRoot,
					s dogma.ProcessEventScope,
					_ dogma.Event,
				) error {
					s.ScheduleTimeout(TimeoutA1, now.Add(-1))
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

			t.Run("does not return timeouts scheduled in the future", func(t *testing.T) {
				env := newProcessControllerTestEnv()
				now := time.Now()

				env.handler.HandleEventFunc = func(
					_ context.Context,
					_ dogma.ProcessRoot,
					s dogma.ProcessEventScope,
					_ dogma.Event,
				) error {
					s.ScheduleTimeout(TimeoutA1, now.Add(1))
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
					env := newProcessControllerTestEnv()
					env.handler.RouteEventToInstanceFunc = func(
						context.Context,
						dogma.Event,
					) (string, bool, error) {
						return "", false, nil
					}

					env.handler.HandleEventFunc = func(
						context.Context,
						dogma.ProcessRoot,
						dogma.ProcessEventScope,
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
					env := newProcessControllerTestEnv()
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
				setup := func(t *testing.T) *processControllerTestEnv {
					t.Helper()

					env := newProcessControllerTestEnv()
					env.handler.HandleEventFunc = func(
						_ context.Context,
						_ dogma.ProcessRoot,
						s dogma.ProcessEventScope,
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
						dogma.ProcessRoot,
						dogma.ProcessEventScope,
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

		t.Run("handling a timeout", func(t *testing.T) {
			setup := func(t *testing.T) *processControllerTestEnv {
				t.Helper()

				env := newProcessControllerTestEnv()
				env.handler.HandleEventFunc = func(
					context.Context,
					dogma.ProcessRoot,
					dogma.ProcessEventScope,
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

				env.handler.HandleTimeoutFunc = func(
					_ context.Context,
					_ dogma.ProcessRoot,
					_ dogma.ProcessTimeoutScope,
					m dogma.Timeout,
				) error {
					called = true
					xtesting.Expect(t, "unexpected timeout", m, TimeoutA1)
					return nil
				}

				_, err := env.ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					env.timeout,
				)
				expectNoError(t, err)

				if !called {
					t.Fatal("expected HandleTimeout() to be called")
				}
			})

			t.Run("propagates handler errors", func(t *testing.T) {
				env := setup(t)
				expected := errors.New("<error>")

				env.handler.HandleTimeoutFunc = func(
					context.Context,
					dogma.ProcessRoot,
					dogma.ProcessTimeoutScope,
					dogma.Timeout,
				) error {
					return expected
				}

				_, err := env.ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					env.timeout,
				)

				xtesting.Expect(t, "unexpected error", err, expected)
			})

			t.Run("returns both commands and timeouts", func(t *testing.T) {
				env := setup(t)
				now := time.Now()

				env.handler.HandleTimeoutFunc = func(
					_ context.Context,
					_ dogma.ProcessRoot,
					s dogma.ProcessTimeoutScope,
					_ dogma.Timeout,
				) error {
					s.ExecuteCommand(CommandA1)
					s.ScheduleTimeout(TimeoutA1, now)
					return nil
				}

				envelopes, err := env.ctrl.Handle(
					context.Background(),
					fact.Ignore,
					now,
					env.timeout,
				)
				expectNoError(t, err)

				expectEnvelopeSet(
					t,
					envelopes,
					[]*envelope.Envelope{
						env.timeout.NewCommand(
							"1",
							CommandA1,
							now,
							envelope.Origin{
								Handler:     env.cfg,
								HandlerType: config.ProcessHandlerType,
								InstanceID:  "<instance-A1>",
							},
						),
						env.timeout.NewTimeout(
							"2",
							TimeoutA1,
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

			t.Run("returns timeouts scheduled in the past", func(t *testing.T) {
				env := setup(t)
				now := time.Now()

				env.handler.HandleTimeoutFunc = func(
					_ context.Context,
					_ dogma.ProcessRoot,
					s dogma.ProcessTimeoutScope,
					_ dogma.Timeout,
				) error {
					s.ScheduleTimeout(TimeoutA2, now.Add(-1))
					return nil
				}

				envelopes, err := env.ctrl.Handle(
					context.Background(),
					fact.Ignore,
					now,
					env.timeout,
				)
				expectNoError(t, err)

				if got, want := len(envelopes), 1; got != want {
					t.Fatalf("unexpected envelope count: got %d, want %d", got, want)
				}
			})

			t.Run("does not return timeouts scheduled in the future", func(t *testing.T) {
				env := setup(t)
				now := time.Now()

				env.handler.HandleTimeoutFunc = func(
					_ context.Context,
					_ dogma.ProcessRoot,
					s dogma.ProcessTimeoutScope,
					_ dogma.Timeout,
				) error {
					s.ScheduleTimeout(TimeoutA2, now.Add(1))
					return nil
				}

				envelopes, err := env.ctrl.Handle(
					context.Background(),
					fact.Ignore,
					now,
					env.timeout,
				)
				expectNoError(t, err)

				if len(envelopes) != 0 {
					t.Fatalf("unexpected envelopes: got %d, want 0", len(envelopes))
				}
			})

			t.Run("when the instance that created the timeout has ended", func(t *testing.T) {
				setupEnded := func(t *testing.T) *processControllerTestEnv {
					t.Helper()

					env := setup(t)
					env.handler.HandleEventFunc = func(
						_ context.Context,
						_ dogma.ProcessRoot,
						s dogma.ProcessEventScope,
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
					env.handler.HandleTimeoutFunc = func(
						context.Context,
						dogma.ProcessRoot,
						dogma.ProcessTimeoutScope,
						dogma.Timeout,
					) error {
						t.Fatal("unexpected call to HandleTimeout()")
						return nil
					}

					_, err := env.ctrl.Handle(
						context.Background(),
						fact.Ignore,
						time.Now(),
						env.timeout,
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
						env.timeout,
					)
					expectNoError(t, err)

					expectFacts(
						t,
						buf.Facts(),
						[]fact.Fact{
							fact.ProcessTimeoutRoutedToEndedInstance{
								Handler:    env.cfg,
								InstanceID: "<instance-A1>",
								Envelope:   env.timeout,
							},
						},
					)
				})
			})
		})

		t.Run("propagates routing errors", func(t *testing.T) {
			env := newProcessControllerTestEnv()
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
			env := newProcessControllerTestEnv()

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
					Implementation: env.cfg.Source.Get(),
					Message:        env.event.Message,
					Description:    "routed an event of type *stubs.EventStub[TypeA] to an empty ID",
				},
				"/stubs/process.go",
			)
		})

		t.Run("when the instance does not exist", func(t *testing.T) {
			t.Run("records facts", func(t *testing.T) {
				env := newProcessControllerTestEnv()
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
				env := newProcessControllerTestEnv()
				env.handler.NewFunc = func() dogma.ProcessRoot {
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
						Implementation: env.cfg.Source.Get(),
						Message:        env.event.Message,
						Description:    "returned a nil ProcessRoot",
					},
					"/stubs/process.go",
				)
			})
		})

		t.Run("when the instance exists", func(t *testing.T) {
			setup := func(t *testing.T) *processControllerTestEnv {
				t.Helper()

				env := newProcessControllerTestEnv()
				env.handler.HandleEventFunc = func(
					_ context.Context,
					_ dogma.ProcessRoot,
					_ dogma.ProcessEventScope,
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

			t.Run("does not call New", func(t *testing.T) {
				env := setup(t)
				env.handler.NewFunc = func() dogma.ProcessRoot {
					t.Fatal("unexpected call to New()")
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
		})

		t.Run("provides more context to UnexpectedMessage panics from RouteEventToInstance", func(t *testing.T) {
			env := newProcessControllerTestEnv()
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
			env := newProcessControllerTestEnv()
			env.handler.HandleEventFunc = func(
				context.Context,
				dogma.ProcessRoot,
				dogma.ProcessEventScope,
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

		t.Run("provides more context to UnexpectedMessage panics from HandleTimeout", func(t *testing.T) {
			env := newProcessControllerTestEnv()
			env.handler.HandleEventFunc = func(
				context.Context,
				dogma.ProcessRoot,
				dogma.ProcessEventScope,
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

			env.handler.HandleTimeoutFunc = func(
				context.Context,
				dogma.ProcessRoot,
				dogma.ProcessTimeoutScope,
				dogma.Timeout,
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
						env.timeout,
					)
				},
				env.cfg,
				"HandleTimeout",
				env.timeout.Message,
			)
		})
	})

	t.Run("Reset", func(t *testing.T) {
		env := newProcessControllerTestEnv()
		env.handler.HandleEventFunc = func(
			context.Context,
			dogma.ProcessRoot,
			dogma.ProcessEventScope,
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

type processControllerTestEnv struct {
	messageIDs envelope.MessageIDGenerator
	handler    *ProcessMessageHandlerStub
	cfg        *config.Process
	ctrl       *Controller
	event      *envelope.Envelope
	timeout    *envelope.Envelope
}

func newProcessControllerTestEnv() *processControllerTestEnv {
	event := envelope.NewEvent(
		"1000",
		EventA1,
		time.Now(),
	)

	handler := &ProcessMessageHandlerStub{
		ConfigureFunc: func(c dogma.ProcessConfigurer) {
			c.Identity("<name>", "7db72921-b805-4db5-8287-0af94a768643")
			c.Routes(
				dogma.HandlesEvent[*EventStub[TypeA]](),
				dogma.ExecutesCommand[*CommandStub[TypeA]](),
				dogma.SchedulesTimeout[*TimeoutStub[TypeA]](),
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
	timeout := event.NewTimeout(
		"2000",
		TimeoutA1,
		time.Now(),
		time.Now().Add(10*time.Second),
		envelope.Origin{
			Handler:     cfg,
			HandlerType: config.ProcessHandlerType,
			InstanceID:  "<instance-A1>",
		},
	)

	env := &processControllerTestEnv{
		handler: handler,
		cfg:     cfg,
		event:   event,
		timeout: timeout,
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
		xtesting.Expect(t, "unexpected implementation", x.Implementation, handler.Source.Get())
		xtesting.Expect(t, "unexpected message", x.Message, message)
	}()

	fn()
}
