package process_test

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogma/fixtures"
	. "github.com/dogmatiq/testkit/engine/internal/process"
	"github.com/dogmatiq/testkit/envelope"
	"github.com/dogmatiq/testkit/fact"
	g "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = g.Describe("type Controller", func() {
	var (
		messageIDs envelope.MessageIDGenerator
		handler    *ProcessMessageHandler
		config     configkit.RichProcess
		ctrl       *Controller
		event      *envelope.Envelope
		timeout    *envelope.Envelope
	)

	g.BeforeEach(func() {
		event = envelope.NewEvent(
			"1000",
			MessageE1,
			time.Now(),
		)

		timeout = event.NewTimeout(
			"2000",
			MessageT1,
			time.Now(),
			time.Now().Add(10*time.Second),
			envelope.Origin{
				Handler:     config,
				HandlerType: configkit.ProcessHandlerType,
				InstanceID:  "<instance-E1>",
			},
		)

		handler = &ProcessMessageHandler{
			ConfigureFunc: func(c dogma.ProcessConfigurer) {
				c.Identity("<name>", "7db72921-b805-4db5-8287-0af94a768643")
				c.ConsumesEventType(MessageE{})
				c.ProducesCommandType(MessageC{})
				c.SchedulesTimeoutType(MessageT{})
			},
			// setup routes for "E" (event) messages to an instance ID based on the
			// message's content
			RouteEventToInstanceFunc: func(
				_ context.Context,
				m dogma.Message,
			) (string, bool, error) {
				switch x := m.(type) {
				case MessageE:
					id := fmt.Sprintf(
						"<instance-%s>",
						x.Value.(string),
					)
					return id, true, nil
				default:
					panic(dogma.UnexpectedMessage)
				}
			},
		}

		config = configkit.FromProcess(handler)

		ctrl = &Controller{
			Config:     config,
			MessageIDs: &messageIDs,
		}

		messageIDs.Reset() // reset after setup for a predictable ID.
	})

	g.Describe("func HandlerConfig()", func() {
		g.It("returns the handler config", func() {
			Expect(ctrl.HandlerConfig()).To(BeIdenticalTo(config))
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
				_ dogma.Message,
			) error {
				// note, calls to ScheduleTimeout are NOT in chronological order
				s.ScheduleTimeout(MessageT3, t3Time)
				s.ScheduleTimeout(MessageT2, t2Time)
				s.ScheduleTimeout(MessageT1, t1Time)

				return nil
			}

			_, err := ctrl.Handle(
				context.Background(),
				fact.Ignore,
				createdTime,
				event,
			)
			Expect(err).ShouldNot(HaveOccurred())

			messageIDs.Reset() // reset after setup for a predictable ID.
		})

		g.It("returns timeouts that are ready to be handled", func() {
			timeouts, err := ctrl.Tick(
				context.Background(),
				fact.Ignore,
				t2Time, // advance time
			)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(timeouts).To(ConsistOf(
				event.NewTimeout(
					"3",
					MessageT1,
					createdTime,
					t1Time,
					envelope.Origin{
						Handler:     config,
						HandlerType: configkit.ProcessHandlerType,
						InstanceID:  "<instance-E1>",
					},
				),
				event.NewTimeout(
					"2",
					MessageT2,
					createdTime,
					t2Time,
					envelope.Origin{
						Handler:     config,
						HandlerType: configkit.ProcessHandlerType,
						InstanceID:  "<instance-E1>",
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

			Expect(err).ShouldNot(HaveOccurred())
			Expect(timeouts).To(HaveLen(2))

			timeouts, err = ctrl.Tick(
				context.Background(),
				fact.Ignore,
				t2Time, // advance time
			)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(timeouts).To(BeEmpty())
		})

		g.It("does not return timeouts for instances that have been ended", func() {
			secondInstanceEvent := envelope.NewEvent(
				"3000",
				MessageE2, // different message value = different instance
				time.Now(),
			)

			_, err := ctrl.Handle(
				context.Background(),
				fact.Ignore,
				createdTime,
				secondInstanceEvent,
			)
			Expect(err).ShouldNot(HaveOccurred())

			// end our original instance
			handler.HandleEventFunc = func(
				_ context.Context,
				_ dogma.ProcessRoot,
				s dogma.ProcessEventScope,
				_ dogma.Message,
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
			Expect(err).ShouldNot(HaveOccurred())

			// expect only the timeout from the E2 instance.
			timeouts, err := ctrl.Tick(
				context.Background(),
				fact.Ignore,
				t2Time,
			)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(timeouts).To(ConsistOf(
				secondInstanceEvent.NewTimeout(
					"3",
					MessageT1,
					createdTime,
					t1Time,
					envelope.Origin{
						Handler:     config,
						HandlerType: configkit.ProcessHandlerType,
						InstanceID:  "<instance-E2>", // E2, not E1!
					},
				),
				secondInstanceEvent.NewTimeout(
					"2",
					MessageT2,
					createdTime,
					t2Time,
					envelope.Origin{
						Handler:     config,
						HandlerType: configkit.ProcessHandlerType,
						InstanceID:  "<instance-E2>", // E2, not E1!
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
					m dogma.Message,
				) error {
					called = true
					Expect(m).To(Equal(MessageE1))
					return nil
				}

				_, err := ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					event,
				)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(called).To(BeTrue())
			})

			g.It("propagates handler errors", func() {
				expected := errors.New("<error>")

				handler.HandleEventFunc = func(
					_ context.Context,
					_ dogma.ProcessRoot,
					_ dogma.ProcessEventScope,
					_ dogma.Message,
				) error {
					return expected
				}

				_, err := ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					event,
				)

				Expect(err).To(Equal(expected))
			})

			g.It("returns both commands and timeouts", func() {
				now := time.Now()

				handler.HandleEventFunc = func(
					_ context.Context,
					_ dogma.ProcessRoot,
					s dogma.ProcessEventScope,
					_ dogma.Message,
				) error {
					s.ExecuteCommand(MessageC1)
					s.ScheduleTimeout(MessageT1, now) // timeouts at current time are "ready"
					return nil
				}

				envelopes, err := ctrl.Handle(
					context.Background(),
					fact.Ignore,
					now,
					event,
				)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(envelopes).To(ConsistOf(
					event.NewCommand(
						"1",
						MessageC1,
						now,
						envelope.Origin{
							Handler:     config,
							HandlerType: configkit.ProcessHandlerType,
							InstanceID:  "<instance-E1>",
						},
					),
					event.NewTimeout(
						"2",
						MessageT1,
						now,
						now,
						envelope.Origin{
							Handler:     config,
							HandlerType: configkit.ProcessHandlerType,
							InstanceID:  "<instance-E1>",
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
					_ dogma.Message,
				) error {
					s.ScheduleTimeout(MessageT1, now.Add(-1))
					return nil
				}

				envelopes, err := ctrl.Handle(
					context.Background(),
					fact.Ignore,
					now,
					event,
				)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(envelopes).To(HaveLen(1))
			})

			g.It("does not return timeouts scheduled in the future", func() {
				now := time.Now()

				handler.HandleEventFunc = func(
					_ context.Context,
					_ dogma.ProcessRoot,
					s dogma.ProcessEventScope,
					_ dogma.Message,
				) error {
					s.ScheduleTimeout(MessageT1, now.Add(1))
					return nil
				}

				envelopes, err := ctrl.Handle(
					context.Background(),
					fact.Ignore,
					now,
					event,
				)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(envelopes).To(BeEmpty())
			})

			g.It("uses the handler's timeout hint", func() {
				hint := 3 * time.Second
				handler.TimeoutHintFunc = func(dogma.Message) time.Duration {
					return hint
				}

				handler.HandleEventFunc = func(
					ctx context.Context,
					_ dogma.ProcessRoot,
					_ dogma.ProcessEventScope,
					_ dogma.Message,
				) error {
					dl, ok := ctx.Deadline()
					Expect(ok).To(BeTrue())
					Expect(dl).To(BeTemporally("~", time.Now().Add(hint)))
					return nil
				}

				_, err := ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					event,
				)

				Expect(err).ShouldNot(HaveOccurred())
			})

			g.When("the event is not routed to an instance", func() {
				g.BeforeEach(func() {
					handler.RouteEventToInstanceFunc = func(
						_ context.Context,
						_ dogma.Message,
					) (string, bool, error) {
						return "", false, nil
					}
				})

				g.It("does not forward the message to the handler", func() {
					handler.HandleEventFunc = func(
						context.Context,
						dogma.ProcessRoot,
						dogma.ProcessEventScope,
						dogma.Message,
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

					Expect(err).ShouldNot(HaveOccurred())
				})

				g.It("records a fact", func() {
					buf := &fact.Buffer{}
					_, err := ctrl.Handle(
						context.Background(),
						buf,
						time.Now(),
						event,
					)

					Expect(err).ShouldNot(HaveOccurred())
					Expect(buf.Facts()).To(ContainElement(
						fact.ProcessEventIgnored{
							Handler:  config,
							Envelope: event,
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
					dogma.Message,
				) error {
					return nil
				}

				_, err := ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					event,
				)
				Expect(err).ShouldNot(HaveOccurred())

				messageIDs.Reset() // reset after setup for a predictable ID.
			})

			g.It("forwards the message to the handler", func() {
				called := false
				handler.HandleTimeoutFunc = func(
					_ context.Context,
					_ dogma.ProcessRoot,
					_ dogma.ProcessTimeoutScope,
					m dogma.Message,
				) error {
					called = true
					Expect(m).To(Equal(MessageT1))
					return nil
				}

				_, err := ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					timeout,
				)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(called).To(BeTrue())
			})

			g.It("propagates handler errors", func() {
				expected := errors.New("<error>")

				handler.HandleTimeoutFunc = func(
					context.Context,
					dogma.ProcessRoot,
					dogma.ProcessTimeoutScope,
					dogma.Message,
				) error {
					return expected
				}

				_, err := ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					timeout,
				)

				Expect(err).To(Equal(expected))
			})

			g.It("returns both commands and timeouts", func() {
				now := time.Now()

				handler.HandleTimeoutFunc = func(
					_ context.Context,
					_ dogma.ProcessRoot,
					s dogma.ProcessTimeoutScope,
					_ dogma.Message,
				) error {
					s.ExecuteCommand(MessageC1)
					s.ScheduleTimeout(MessageT1, now) // timeouts at current time are "ready"
					return nil
				}

				envelopes, err := ctrl.Handle(
					context.Background(),
					fact.Ignore,
					now,
					timeout,
				)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(envelopes).To(ConsistOf(
					timeout.NewCommand(
						"1",
						MessageC1,
						now,
						envelope.Origin{
							Handler:     config,
							HandlerType: configkit.ProcessHandlerType,
							InstanceID:  "<instance-E1>",
						},
					),
					timeout.NewTimeout(
						"2",
						MessageT1,
						now,
						now,
						envelope.Origin{
							Handler:     config,
							HandlerType: configkit.ProcessHandlerType,
							InstanceID:  "<instance-E1>",
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
					_ dogma.Message,
				) error {
					s.ScheduleTimeout(MessageT2, now.Add(-1))
					return nil
				}

				envelopes, err := ctrl.Handle(
					context.Background(),
					fact.Ignore,
					now,
					timeout,
				)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(envelopes).To(HaveLen(1))
			})

			g.It("does not return timeouts scheduled in the future", func() {
				now := time.Now()

				handler.HandleTimeoutFunc = func(
					_ context.Context,
					_ dogma.ProcessRoot,
					s dogma.ProcessTimeoutScope,
					_ dogma.Message,
				) error {
					s.ScheduleTimeout(MessageT2, now.Add(1))
					return nil
				}

				envelopes, err := ctrl.Handle(
					context.Background(),
					fact.Ignore,
					now,
					event,
				)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(envelopes).To(BeEmpty())
			})

			g.It("uses the handler's timeout hint", func() {
				hint := 3 * time.Second
				handler.TimeoutHintFunc = func(dogma.Message) time.Duration {
					return hint
				}

				handler.HandleTimeoutFunc = func(
					ctx context.Context,
					_ dogma.ProcessRoot,
					_ dogma.ProcessTimeoutScope,
					_ dogma.Message,
				) error {
					dl, ok := ctx.Deadline()
					Expect(ok).To(BeTrue())
					Expect(dl).To(BeTemporally("~", time.Now().Add(hint)))
					return nil
				}

				_, err := ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					timeout,
				)

				Expect(err).ShouldNot(HaveOccurred())
			})

			g.When("the instance that created the timeout does not exist", func() {
				g.BeforeEach(func() {
					handler.HandleEventFunc = func(
						_ context.Context,
						_ dogma.ProcessRoot,
						s dogma.ProcessEventScope,
						_ dogma.Message,
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
					Expect(err).ShouldNot(HaveOccurred())

					messageIDs.Reset() // reset after setup for a predictable ID.
				})

				g.It("does not forward the message to the handler", func() {
					handler.HandleTimeoutFunc = func(
						context.Context,
						dogma.ProcessRoot,
						dogma.ProcessTimeoutScope,
						dogma.Message,
					) error {
						g.Fail("unexpected call to HandleEvent()")
						return nil
					}

					_, err := ctrl.Handle(
						context.Background(),
						fact.Ignore,
						time.Now(),
						timeout,
					)

					Expect(err).ShouldNot(HaveOccurred())
				})

				g.It("records a fact", func() {
					buf := &fact.Buffer{}
					_, err := ctrl.Handle(
						context.Background(),
						buf,
						time.Now(),
						timeout,
					)

					Expect(err).ShouldNot(HaveOccurred())
					Expect(buf.Facts()).To(ContainElement(
						fact.ProcessTimeoutIgnored{
							Handler:    config,
							InstanceID: "<instance-E1>",
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
				_ dogma.Message,
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

			Expect(err).To(Equal(expected))
		})

		g.It("panics when the handler routes to an empty instance ID", func() {
			handler.RouteEventToInstanceFunc = func(
				context.Context,
				dogma.Message,
			) (string, bool, error) {
				return "", true, nil
			}

			Expect(func() {
				ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					event,
				)
			}).To(PanicWith(
				MatchAllFields(
					Fields{
						"Handler":        Equal(config),
						"Interface":      Equal("ProcessMessageHandler"),
						"Method":         Equal("RouteEventToInstance"),
						"Implementation": Equal(config.Handler()),
						"Message":        Equal(event.Message),
						"Description":    Equal("routed an event of type fixtures.MessageE to an empty ID"),
						"Location": MatchAllFields(
							Fields{
								"Func": Not(BeEmpty()),
								"File": HaveSuffix("/fixtures/process.go"), // from dogmatiq/dogma module
								"Line": Not(BeZero()),
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

				Expect(err).ShouldNot(HaveOccurred())
				Expect(buf.Facts()).To(ContainElement(
					fact.ProcessInstanceNotFound{
						Handler:    config,
						InstanceID: "<instance-E1>",
						Envelope:   event,
					},
				))
				Expect(buf.Facts()).To(ContainElement(
					fact.ProcessInstanceBegun{
						Handler:    config,
						InstanceID: "<instance-E1>",
						Root:       &ProcessRoot{},
						Envelope:   event,
					},
				))
			})

			g.It("panics if New() returns nil", func() {
				handler.NewFunc = func() dogma.ProcessRoot {
					return nil
				}

				Expect(func() {
					ctrl.Handle(
						context.Background(),
						fact.Ignore,
						time.Now(),
						event,
					)
				}).To(PanicWith(
					MatchAllFields(
						Fields{
							"Handler":        Equal(config),
							"Interface":      Equal("ProcessMessageHandler"),
							"Method":         Equal("New"),
							"Implementation": Equal(config.Handler()),
							"Message":        Equal(event.Message),
							"Description":    Equal("returned a nil ProcessRoot"),
							"Location": MatchAllFields(
								Fields{
									"Func": Not(BeEmpty()),
									"File": HaveSuffix("/fixtures/process.go"), // from dogmatiq/dogma module
									"Line": Not(BeZero()),
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
					_ dogma.Message,
				) error {
					return nil
				}

				_, err := ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					event,
				)

				Expect(err).ShouldNot(HaveOccurred())

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

				Expect(err).ShouldNot(HaveOccurred())
				Expect(buf.Facts()).To(ContainElement(
					fact.ProcessInstanceLoaded{
						Handler:    config,
						InstanceID: "<instance-E1>",
						Root:       &ProcessRoot{},
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
				dogma.Message,
			) (string, bool, error) {
				panic(dogma.UnexpectedMessage)
			}

			Expect(func() {
				ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					event,
				)
			}).To(PanicWith(
				MatchFields(
					IgnoreExtras,
					Fields{
						"Handler":   Equal(config),
						"Interface": Equal("ProcessMessageHandler"),
						"Method":    Equal("RouteEventToInstance"),
						"Message":   Equal(event.Message),
					},
				),
			))
		})

		g.It("provides more context to UnexpectedMessage panics from HandleEvent()", func() {
			handler.HandleEventFunc = func(
				context.Context,
				dogma.ProcessRoot,
				dogma.ProcessEventScope,
				dogma.Message,
			) error {
				panic(dogma.UnexpectedMessage)
			}

			Expect(func() {
				ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					event,
				)
			}).To(PanicWith(
				MatchFields(
					IgnoreExtras,
					Fields{
						"Handler":   Equal(config),
						"Interface": Equal("ProcessMessageHandler"),
						"Method":    Equal("HandleEvent"),
						"Message":   Equal(event.Message),
					},
				),
			))
		})

		g.It("provides more context to UnexpectedMessage panics from HandleTimeout()", func() {
			handler.HandleEventFunc = func(
				context.Context,
				dogma.ProcessRoot,
				dogma.ProcessEventScope,
				dogma.Message,
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
				dogma.Message,
			) error {
				panic(dogma.UnexpectedMessage)
			}

			Expect(func() {
				ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					timeout,
				)
			}).To(PanicWith(
				MatchFields(
					IgnoreExtras,
					Fields{
						"Handler":   Equal(config),
						"Interface": Equal("ProcessMessageHandler"),
						"Method":    Equal("HandleTimeout"),
						"Message":   Equal(timeout.Message),
					},
				),
			))
		})

		g.It("provides more context to UnexpectedMessage panics from TimeoutHint()", func() {
			handler.TimeoutHintFunc = func(
				dogma.Message,
			) time.Duration {
				panic(dogma.UnexpectedMessage)
			}

			Expect(func() {
				ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					event,
				)
			}).To(PanicWith(
				MatchFields(
					IgnoreExtras,
					Fields{
						"Handler":   Equal(config),
						"Interface": Equal("ProcessMessageHandler"),
						"Method":    Equal("TimeoutHint"),
						"Message":   Equal(event.Message),
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
				dogma.Message,
			) error {
				return nil
			}

			_, err := ctrl.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				event,
			)

			Expect(err).ShouldNot(HaveOccurred())

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

			Expect(err).ShouldNot(HaveOccurred())
			Expect(buf.Facts()).NotTo(ContainElement(
				BeAssignableToTypeOf(fact.ProcessInstanceLoaded{}),
			))
		})
	})
})
