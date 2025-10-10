package process_test

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/enginekit/config"
	"github.com/dogmatiq/enginekit/config/runtimeconfig"
	. "github.com/dogmatiq/enginekit/enginetest/stubs"
	. "github.com/dogmatiq/testkit/engine/internal/process"
	"github.com/dogmatiq/testkit/envelope"
	"github.com/dogmatiq/testkit/fact"
	g "github.com/onsi/ginkgo/v2"
	gm "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = g.Describe("type Controller", func() {
	var (
		messageIDs envelope.MessageIDGenerator
		handler    *ProcessMessageHandlerStub
		cfg        *config.Process
		ctrl       *Controller
		event      *envelope.Envelope
		timeout    *envelope.Envelope
	)

	g.BeforeEach(func() {
		event = envelope.NewEvent(
			"1000",
			EventA1,
			time.Now(),
		)

		timeout = event.NewTimeout(
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

		handler = &ProcessMessageHandlerStub{
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
					id := fmt.Sprintf(
						"<instance-%s>",
						x.Content,
					)
					return id, true, nil
				default:
					panic(dogma.UnexpectedMessage)
				}
			},
		}

		cfg = runtimeconfig.FromProcess(handler)

		ctrl = &Controller{
			Config:     cfg,
			MessageIDs: &messageIDs,
		}

		messageIDs.Reset() // reset after setup for a predictable ID.
	})

	g.Describe("func HandlerConfig()", func() {
		g.It("returns the handler config", func() {
			gm.Expect(ctrl.HandlerConfig()).To(gm.BeIdenticalTo(cfg))
		})
	})

	g.Describe("func Tick()", func() {
		var (
			createdTime time.Time
			t1Time      time.Time
			t2Time      time.Time
			t3Time      time.Time
		)

		g.BeforeEach(func() {
			createdTime = time.Now()

			t1Time = createdTime.Add(1 * time.Hour)
			t2Time = createdTime.Add(2 * time.Hour)
			t3Time = createdTime.Add(3 * time.Hour)

			handler.HandleEventFunc = func(
				_ context.Context,
				_ dogma.ProcessRoot,
				s dogma.ProcessEventScope,
				_ dogma.Event,
			) error {
				// note, calls to ScheduleTimeout are NOT in chronological order
				s.ScheduleTimeout(TimeoutA3, t3Time)
				s.ScheduleTimeout(TimeoutA2, t2Time)
				s.ScheduleTimeout(TimeoutA1, t1Time)

				return nil
			}

			_, err := ctrl.Handle(
				context.Background(),
				fact.Ignore,
				createdTime,
				event,
			)
			gm.Expect(err).ShouldNot(gm.HaveOccurred())

			messageIDs.Reset() // reset after setup for a predictable ID.
		})

		g.It("returns timeouts that are ready to be handled", func() {
			timeouts, err := ctrl.Tick(
				context.Background(),
				fact.Ignore,
				t2Time, // advance time
			)

			gm.Expect(err).ShouldNot(gm.HaveOccurred())
			gm.Expect(timeouts).To(gm.ConsistOf(
				event.NewTimeout(
					"3",
					TimeoutA1,
					createdTime,
					t1Time,
					envelope.Origin{
						Handler:     cfg,
						HandlerType: config.ProcessHandlerType,
						InstanceID:  "<instance-A1>",
					},
				),
				event.NewTimeout(
					"2",
					TimeoutA2,
					createdTime,
					t2Time,
					envelope.Origin{
						Handler:     cfg,
						HandlerType: config.ProcessHandlerType,
						InstanceID:  "<instance-A1>",
					},
				),
			))
		})

		g.It("does not return the same timeouts multiple times", func() {
			timeouts, err := ctrl.Tick(
				context.Background(),
				fact.Ignore,
				t2Time, // advance time
			)

			gm.Expect(err).ShouldNot(gm.HaveOccurred())
			gm.Expect(timeouts).To(gm.HaveLen(2))

			timeouts, err = ctrl.Tick(
				context.Background(),
				fact.Ignore,
				t2Time, // advance time
			)

			gm.Expect(err).ShouldNot(gm.HaveOccurred())
			gm.Expect(timeouts).To(gm.BeEmpty())
		})

		g.It("does not return timeouts for instances that have been ended", func() {
			secondInstanceEvent := envelope.NewEvent(
				"3000",
				EventA2, // different message value = different instance
				time.Now(),
			)

			_, err := ctrl.Handle(
				context.Background(),
				fact.Ignore,
				createdTime,
				secondInstanceEvent,
			)
			gm.Expect(err).ShouldNot(gm.HaveOccurred())

			// end our original instance
			handler.HandleEventFunc = func(
				_ context.Context,
				_ dogma.ProcessRoot,
				s dogma.ProcessEventScope,
				_ dogma.Event,
			) error {
				s.End()
				return nil
			}

			_, err = ctrl.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				event,
			)
			gm.Expect(err).ShouldNot(gm.HaveOccurred())

			// expect only the timeout from the E2 instance.
			timeouts, err := ctrl.Tick(
				context.Background(),
				fact.Ignore,
				t2Time,
			)

			gm.Expect(err).ShouldNot(gm.HaveOccurred())
			gm.Expect(timeouts).To(gm.ConsistOf(
				secondInstanceEvent.NewTimeout(
					"3",
					TimeoutA1,
					createdTime,
					t1Time,
					envelope.Origin{
						Handler:     cfg,
						HandlerType: config.ProcessHandlerType,
						InstanceID:  "<instance-A2>", // A2, not A1!
					},
				),
				secondInstanceEvent.NewTimeout(
					"2",
					TimeoutA2,
					createdTime,
					t2Time,
					envelope.Origin{
						Handler:     cfg,
						HandlerType: config.ProcessHandlerType,
						InstanceID:  "<instance-A2>", // A2, not A1!
					},
				),
			))
		})
	})

	g.Describe("func Handle()", func() {
		g.When("handling an event", func() {
			g.It("forwards the message to the handler", func() {
				called := false
				handler.HandleEventFunc = func(
					_ context.Context,
					_ dogma.ProcessRoot,
					_ dogma.ProcessEventScope,
					m dogma.Event,
				) error {
					called = true
					gm.Expect(m).To(gm.Equal(EventA1))
					return nil
				}

				_, err := ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					event,
				)

				gm.Expect(err).ShouldNot(gm.HaveOccurred())
				gm.Expect(called).To(gm.BeTrue())
			})

			g.It("propagates handler errors", func() {
				expected := errors.New("<error>")

				handler.HandleEventFunc = func(
					_ context.Context,
					_ dogma.ProcessRoot,
					_ dogma.ProcessEventScope,
					_ dogma.Event,
				) error {
					return expected
				}

				_, err := ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					event,
				)

				gm.Expect(err).To(gm.Equal(expected))
			})

			g.It("returns both commands and timeouts", func() {
				now := time.Now()

				handler.HandleEventFunc = func(
					_ context.Context,
					_ dogma.ProcessRoot,
					s dogma.ProcessEventScope,
					_ dogma.Event,
				) error {
					s.ExecuteCommand(CommandA1)
					s.ScheduleTimeout(TimeoutA1, now) // timeouts at current time are "ready"
					return nil
				}

				envelopes, err := ctrl.Handle(
					context.Background(),
					fact.Ignore,
					now,
					event,
				)

				gm.Expect(err).ShouldNot(gm.HaveOccurred())
				gm.Expect(envelopes).To(gm.ConsistOf(
					event.NewCommand(
						"1",
						CommandA1,
						now,
						envelope.Origin{
							Handler:     cfg,
							HandlerType: config.ProcessHandlerType,
							InstanceID:  "<instance-A1>",
						},
					),
					event.NewTimeout(
						"2",
						TimeoutA1,
						now,
						now,
						envelope.Origin{
							Handler:     cfg,
							HandlerType: config.ProcessHandlerType,
							InstanceID:  "<instance-A1>",
						},
					),
				))
			})

			g.It("returns timeouts scheduled in the past", func() {
				now := time.Now()

				handler.HandleEventFunc = func(
					_ context.Context,
					_ dogma.ProcessRoot,
					s dogma.ProcessEventScope,
					_ dogma.Event,
				) error {
					s.ScheduleTimeout(TimeoutA1, now.Add(-1))
					return nil
				}

				envelopes, err := ctrl.Handle(
					context.Background(),
					fact.Ignore,
					now,
					event,
				)

				gm.Expect(err).ShouldNot(gm.HaveOccurred())
				gm.Expect(envelopes).To(gm.HaveLen(1))
			})

			g.It("does not return timeouts scheduled in the future", func() {
				now := time.Now()

				handler.HandleEventFunc = func(
					_ context.Context,
					_ dogma.ProcessRoot,
					s dogma.ProcessEventScope,
					_ dogma.Event,
				) error {
					s.ScheduleTimeout(TimeoutA1, now.Add(1))
					return nil
				}

				envelopes, err := ctrl.Handle(
					context.Background(),
					fact.Ignore,
					now,
					event,
				)

				gm.Expect(err).ShouldNot(gm.HaveOccurred())
				gm.Expect(envelopes).To(gm.BeEmpty())
			})

			g.When("the event is not routed to an instance", func() {
				g.BeforeEach(func() {
					handler.RouteEventToInstanceFunc = func(
						_ context.Context,
						_ dogma.Event,
					) (string, bool, error) {
						return "", false, nil
					}
				})

				g.It("does not forward the message to the handler", func() {
					handler.HandleEventFunc = func(
						context.Context,
						dogma.ProcessRoot,
						dogma.ProcessEventScope,
						dogma.Event,
					) error {
						g.Fail("unexpected call to HandleEvent()")
						return nil
					}

					_, err := ctrl.Handle(
						context.Background(),
						fact.Ignore,
						time.Now(),
						event,
					)

					gm.Expect(err).ShouldNot(gm.HaveOccurred())
				})

				g.It("records a fact", func() {
					buf := &fact.Buffer{}
					_, err := ctrl.Handle(
						context.Background(),
						buf,
						time.Now(),
						event,
					)

					gm.Expect(err).ShouldNot(gm.HaveOccurred())
					gm.Expect(buf.Facts()).To(gm.ContainElement(
						fact.ProcessEventIgnored{
							Handler:  cfg,
							Envelope: event,
						},
					))
				})
			})

			g.When("the instance has ended", func() {
				g.BeforeEach(func() {
					handler.HandleEventFunc = func(
						_ context.Context,
						_ dogma.ProcessRoot,
						s dogma.ProcessEventScope,
						_ dogma.Event,
					) error {
						s.End()
						return nil
					}

					_, err := ctrl.Handle(
						context.Background(),
						fact.Ignore,
						time.Now(),
						event,
					)
					gm.Expect(err).ShouldNot(gm.HaveOccurred())

					messageIDs.Reset() // reset after setup for a predictable ID.
				})

				g.It("does not forward the message to the handler", func() {
					handler.HandleEventFunc = func(
						context.Context,
						dogma.ProcessRoot,
						dogma.ProcessEventScope,
						dogma.Event,
					) error {
						g.Fail("unexpected call to HandleEvent()")
						return nil
					}

					_, err := ctrl.Handle(
						context.Background(),
						fact.Ignore,
						time.Now(),
						event,
					)

					gm.Expect(err).ShouldNot(gm.HaveOccurred())
				})

				g.It("records a fact", func() {
					buf := &fact.Buffer{}
					_, err := ctrl.Handle(
						context.Background(),
						buf,
						time.Now(),
						event,
					)

					gm.Expect(err).ShouldNot(gm.HaveOccurred())
					gm.Expect(buf.Facts()).To(gm.ContainElement(
						fact.ProcessEventRoutedToEndedInstance{
							Handler:    cfg,
							InstanceID: "<instance-A1>",
							Envelope:   event,
						},
					))
				})
			})
		})

		g.When("handling a timeout", func() {
			g.BeforeEach(func() {
				handler.HandleEventFunc = func(
					context.Context,
					dogma.ProcessRoot,
					dogma.ProcessEventScope,
					dogma.Event,
				) error {
					return nil
				}

				_, err := ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					event,
				)
				gm.Expect(err).ShouldNot(gm.HaveOccurred())

				messageIDs.Reset() // reset after setup for a predictable ID.
			})

			g.It("forwards the message to the handler", func() {
				_, err := ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					timeout,
				)

				gm.Expect(err).ShouldNot(gm.HaveOccurred())
			})

			g.It("propagates handler errors", func() {
				expected := errors.New("<error>")

				handler.HandleTimeoutFunc = func(
					context.Context,
					dogma.ProcessRoot,
					dogma.ProcessTimeoutScope,
					dogma.Timeout,
				) error {
					return expected
				}

				_, err := ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					timeout,
				)

				gm.Expect(err).To(gm.Equal(expected))
			})

			g.It("returns both commands and timeouts", func() {
				now := time.Now()

				handler.HandleTimeoutFunc = func(
					_ context.Context,
					_ dogma.ProcessRoot,
					s dogma.ProcessTimeoutScope,
					_ dogma.Timeout,
				) error {
					s.ExecuteCommand(CommandA1)
					s.ScheduleTimeout(TimeoutA1, now) // timeouts at current time are "ready"
					return nil
				}

				envelopes, err := ctrl.Handle(
					context.Background(),
					fact.Ignore,
					now,
					timeout,
				)

				gm.Expect(err).ShouldNot(gm.HaveOccurred())
				gm.Expect(envelopes).To(gm.ConsistOf(
					timeout.NewCommand(
						"1",
						CommandA1,
						now,
						envelope.Origin{
							Handler:     cfg,
							HandlerType: config.ProcessHandlerType,
							InstanceID:  "<instance-A1>",
						},
					),
					timeout.NewTimeout(
						"2",
						TimeoutA1,
						now,
						now,
						envelope.Origin{
							Handler:     cfg,
							HandlerType: config.ProcessHandlerType,
							InstanceID:  "<instance-A1>",
						},
					),
				))
			})

			g.It("returns timeouts scheduled in the past", func() {
				now := time.Now()

				handler.HandleTimeoutFunc = func(
					_ context.Context,
					_ dogma.ProcessRoot,
					s dogma.ProcessTimeoutScope,
					_ dogma.Timeout,
				) error {
					s.ScheduleTimeout(TimeoutA2, now.Add(-1))
					return nil
				}

				envelopes, err := ctrl.Handle(
					context.Background(),
					fact.Ignore,
					now,
					timeout,
				)

				gm.Expect(err).ShouldNot(gm.HaveOccurred())
				gm.Expect(envelopes).To(gm.HaveLen(1))
			})

			g.It("does not return timeouts scheduled in the future", func() {
				now := time.Now()

				handler.HandleTimeoutFunc = func(
					_ context.Context,
					_ dogma.ProcessRoot,
					s dogma.ProcessTimeoutScope,
					_ dogma.Timeout,
				) error {
					s.ScheduleTimeout(TimeoutA2, now.Add(1))
					return nil
				}

				envelopes, err := ctrl.Handle(
					context.Background(),
					fact.Ignore,
					now,
					event,
				)

				gm.Expect(err).ShouldNot(gm.HaveOccurred())
				gm.Expect(envelopes).To(gm.BeEmpty())
			})

			g.When("the instance that created the timeout has ended", func() {
				g.BeforeEach(func() {
					handler.HandleEventFunc = func(
						_ context.Context,
						_ dogma.ProcessRoot,
						s dogma.ProcessEventScope,
						_ dogma.Event,
					) error {
						s.End()
						return nil
					}

					_, err := ctrl.Handle(
						context.Background(),
						fact.Ignore,
						time.Now(),
						event,
					)
					gm.Expect(err).ShouldNot(gm.HaveOccurred())

					messageIDs.Reset() // reset after setup for a predictable ID.
				})

				g.It("does not forward the message to the handler", func() {
					handler.HandleTimeoutFunc = func(
						context.Context,
						dogma.ProcessRoot,
						dogma.ProcessTimeoutScope,
						dogma.Timeout,
					) error {
						g.Fail("unexpected call to HandleTimeout()")
						return nil
					}

					_, err := ctrl.Handle(
						context.Background(),
						fact.Ignore,
						time.Now(),
						timeout,
					)

					gm.Expect(err).ShouldNot(gm.HaveOccurred())
				})

				g.It("records a fact", func() {
					buf := &fact.Buffer{}
					_, err := ctrl.Handle(
						context.Background(),
						buf,
						time.Now(),
						timeout,
					)

					gm.Expect(err).ShouldNot(gm.HaveOccurred())
					gm.Expect(buf.Facts()).To(gm.ContainElement(
						fact.ProcessTimeoutRoutedToEndedInstance{
							Handler:    cfg,
							InstanceID: "<instance-A1>",
							Envelope:   timeout,
						},
					))
				})
			})
		})

		g.It("propagates routing errors", func() {
			expected := errors.New("<error>")

			handler.RouteEventToInstanceFunc = func(
				_ context.Context,
				_ dogma.Event,
			) (string, bool, error) {
				// note, we return a valid id and true here to verify that the error is
				// checked first.
				return "<instance>", true, expected
			}

			_, err := ctrl.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				event,
			)

			gm.Expect(err).To(gm.Equal(expected))
		})

		g.It("panics when the handler routes to an empty instance ID", func() {
			handler.RouteEventToInstanceFunc = func(
				context.Context,
				dogma.Event,
			) (string, bool, error) {
				return "", true, nil
			}

			gm.Expect(func() {
				ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					event,
				)
			}).To(gm.PanicWith(
				MatchAllFields(
					Fields{
						"Handler":        gm.Equal(cfg),
						"Interface":      gm.Equal("ProcessMessageHandler"),
						"Method":         gm.Equal("RouteEventToInstance"),
						"Implementation": gm.Equal(cfg.Source.Get()),
						"Message":        gm.Equal(event.Message),
						"Description":    gm.Equal("routed an event of type *stubs.EventStub[TypeA] to an empty ID"),
						"Location": MatchAllFields(
							Fields{
								"Func": gm.Not(gm.BeEmpty()),
								"File": gm.HaveSuffix("/stubs/process.go"), // from dogmatiq/enginekit module
								"Line": gm.Not(gm.BeZero()),
							},
						),
					},
				),
			))
		})

		g.When("the instance does not exist", func() {
			g.It("records facts", func() {
				buf := &fact.Buffer{}
				_, err := ctrl.Handle(
					context.Background(),
					buf,
					time.Now(),
					event,
				)

				gm.Expect(err).ShouldNot(gm.HaveOccurred())
				gm.Expect(buf.Facts()).To(gm.ContainElement(
					fact.ProcessInstanceNotFound{
						Handler:    cfg,
						InstanceID: "<instance-A1>",
						Envelope:   event,
					},
				))
				gm.Expect(buf.Facts()).To(gm.ContainElement(
					fact.ProcessInstanceBegun{
						Handler:    cfg,
						InstanceID: "<instance-A1>",
						Root:       &ProcessRootStub{},
						Envelope:   event,
					},
				))
			})

			g.It("panics if New() returns nil", func() {
				handler.NewFunc = func() dogma.ProcessRoot {
					return nil
				}

				gm.Expect(func() {
					ctrl.Handle(
						context.Background(),
						fact.Ignore,
						time.Now(),
						event,
					)
				}).To(gm.PanicWith(
					MatchAllFields(
						Fields{
							"Handler":        gm.Equal(cfg),
							"Interface":      gm.Equal("ProcessMessageHandler"),
							"Method":         gm.Equal("New"),
							"Implementation": gm.Equal(cfg.Source.Get()),
							"Message":        gm.Equal(event.Message),
							"Description":    gm.Equal("returned a nil ProcessRoot"),
							"Location": MatchAllFields(
								Fields{
									"Func": gm.Not(gm.BeEmpty()),
									"File": gm.HaveSuffix("/stubs/process.go"), // from dogmatiq/enginekit module
									"Line": gm.Not(gm.BeZero()),
								},
							),
						},
					),
				))
			})
		})

		g.When("the instance exists", func() {
			g.BeforeEach(func() {
				handler.HandleEventFunc = func(
					_ context.Context,
					_ dogma.ProcessRoot,
					s dogma.ProcessEventScope,
					_ dogma.Event,
				) error {
					return nil
				}

				_, err := ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					event,
				)

				gm.Expect(err).ShouldNot(gm.HaveOccurred())

				messageIDs.Reset() // reset after setup for a predictable ID.
			})

			g.It("records a fact", func() {
				buf := &fact.Buffer{}
				_, err := ctrl.Handle(
					context.Background(),
					buf,
					time.Now(),
					event,
				)

				gm.Expect(err).ShouldNot(gm.HaveOccurred())
				gm.Expect(buf.Facts()).To(gm.ContainElement(
					fact.ProcessInstanceLoaded{
						Handler:    cfg,
						InstanceID: "<instance-A1>",
						Root:       &ProcessRootStub{},
						Envelope:   event,
					},
				))
			})

			g.It("does not call New()", func() {
				handler.NewFunc = func() dogma.ProcessRoot {
					g.Fail("unexpected call to New()")
					return nil
				}

				ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					event,
				)
			})
		})

		g.It("provides more context to UnexpectedMessage panics from RouteEventToInstance()", func() {
			handler.RouteEventToInstanceFunc = func(
				context.Context,
				dogma.Event,
			) (string, bool, error) {
				panic(dogma.UnexpectedMessage)
			}

			gm.Expect(func() {
				ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					event,
				)
			}).To(gm.PanicWith(
				MatchFields(
					IgnoreExtras,
					Fields{
						"Handler":   gm.Equal(cfg),
						"Interface": gm.Equal("ProcessMessageHandler"),
						"Method":    gm.Equal("RouteEventToInstance"),
						"Message":   gm.Equal(event.Message),
					},
				),
			))
		})

		g.It("provides more context to UnexpectedMessage panics from HandleEvent()", func() {
			handler.HandleEventFunc = func(
				context.Context,
				dogma.ProcessRoot,
				dogma.ProcessEventScope,
				dogma.Event,
			) error {
				panic(dogma.UnexpectedMessage)
			}

			gm.Expect(func() {
				ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					event,
				)
			}).To(gm.PanicWith(
				MatchFields(
					IgnoreExtras,
					Fields{
						"Handler":   gm.Equal(cfg),
						"Interface": gm.Equal("ProcessMessageHandler"),
						"Method":    gm.Equal("HandleEvent"),
						"Message":   gm.Equal(event.Message),
					},
				),
			))
		})

		g.It("provides more context to UnexpectedMessage panics from HandleTimeout()", func() {
			handler.HandleEventFunc = func(
				context.Context,
				dogma.ProcessRoot,
				dogma.ProcessEventScope,
				dogma.Event,
			) error {
				return nil
			}

			ctrl.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				event,
			)

			handler.HandleTimeoutFunc = func(
				context.Context,
				dogma.ProcessRoot,
				dogma.ProcessTimeoutScope,
				dogma.Timeout,
			) error {
				panic(dogma.UnexpectedMessage)
			}

			gm.Expect(func() {
				ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					timeout,
				)
			}).To(gm.PanicWith(
				MatchFields(
					IgnoreExtras,
					Fields{
						"Handler":   gm.Equal(cfg),
						"Interface": gm.Equal("ProcessMessageHandler"),
						"Method":    gm.Equal("HandleTimeout"),
						"Message":   gm.Equal(timeout.Message),
					},
				),
			))
		})
	})

	g.Describe("func Reset()", func() {
		g.BeforeEach(func() {
			handler.HandleEventFunc = func(
				context.Context,
				dogma.ProcessRoot,
				dogma.ProcessEventScope,
				dogma.Event,
			) error {
				return nil
			}

			_, err := ctrl.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				event,
			)

			gm.Expect(err).ShouldNot(gm.HaveOccurred())

			messageIDs.Reset() // reset after setup for a predictable ID.
		})

		g.It("removes all instances", func() {
			ctrl.Reset()

			buf := &fact.Buffer{}
			_, err := ctrl.Handle(
				context.Background(),
				buf,
				time.Now(),
				event,
			)

			gm.Expect(err).ShouldNot(gm.HaveOccurred())
			gm.Expect(buf.Facts()).NotTo(gm.ContainElement(
				gm.BeAssignableToTypeOf(fact.ProcessInstanceLoaded{}),
			))
		})
	})
})
