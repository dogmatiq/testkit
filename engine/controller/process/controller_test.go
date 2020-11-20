package process_test

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogma/fixtures"
	"github.com/dogmatiq/testkit/engine/controller"
	. "github.com/dogmatiq/testkit/engine/controller/process"
	"github.com/dogmatiq/testkit/engine/envelope"
	"github.com/dogmatiq/testkit/engine/fact"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ controller.Controller = &Controller{}

var _ = Describe("type Controller", func() {
	var (
		messageIDs envelope.MessageIDGenerator
		handler    *ProcessMessageHandler
		config     configkit.RichProcess
		ctrl       *Controller
		event      *envelope.Envelope
		timeout    *envelope.Envelope
	)

	BeforeEach(func() {
		event = envelope.NewEvent(
			"1000",
			MessageE1,
			time.Now(),
		)

		timeout = event.NewTimeout(
			"2000",
			MessageT1,
			time.Now(),
			time.Now(),
			envelope.Origin{
				Handler:     config,
				HandlerType: configkit.ProcessHandlerType,
				InstanceID:  "<instance-E1>",
			},
		)

		handler = &ProcessMessageHandler{
			ConfigureFunc: func(c dogma.ProcessConfigurer) {
				c.Identity("<name>", "<key>")
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

	Describe("func HandlerConfig()", func() {
		It("returns the handler config", func() {
			Expect(ctrl.HandlerConfig()).To(BeIdenticalTo(config))
		})
	})

	Describe("func Tick()", func() {
		var (
			createdTime time.Time
			t1Time      time.Time
			t2Time      time.Time
			t3Time      time.Time
		)

		BeforeEach(func() {
			createdTime = time.Now()

			t1Time = createdTime.Add(1 * time.Hour)
			t2Time = createdTime.Add(2 * time.Hour)
			t3Time = createdTime.Add(3 * time.Hour)

			handler.HandleEventFunc = func(
				_ context.Context,
				s dogma.ProcessEventScope,
				_ dogma.Message,
			) error {
				s.Begin()

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

		It("returns timeouts that are ready to be handled", func() {
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

		It("does not return the same timeouts multiple times", func() {
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

		It("does not return timeouts for instances that have been deleted", func() {
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

	Describe("func Handle()", func() {
		When("handling an event", func() {
			It("forwards the message to the handler", func() {
				called := false
				handler.HandleEventFunc = func(
					_ context.Context,
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

			It("propagates handler errors", func() {
				expected := errors.New("<error>")

				handler.HandleEventFunc = func(
					_ context.Context,
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

			It("returns both commands and timeouts", func() {
				now := time.Now()

				handler.HandleEventFunc = func(
					_ context.Context,
					s dogma.ProcessEventScope,
					_ dogma.Message,
				) error {
					s.Begin()
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

			It("returns timeouts scheduled in the past", func() {
				now := time.Now()

				handler.HandleEventFunc = func(
					_ context.Context,
					s dogma.ProcessEventScope,
					_ dogma.Message,
				) error {
					s.Begin()
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

			It("does not return timeouts scheduled in the future", func() {
				now := time.Now()

				handler.HandleEventFunc = func(
					_ context.Context,
					s dogma.ProcessEventScope,
					_ dogma.Message,
				) error {
					s.Begin()
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

			It("uses the handler's timeout hint", func() {
				hint := 3 * time.Second
				handler.TimeoutHintFunc = func(dogma.Message) time.Duration {
					return hint
				}

				handler.HandleEventFunc = func(
					ctx context.Context,
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

			When("the event is not routed to an instance", func() {
				BeforeEach(func() {
					handler.RouteEventToInstanceFunc = func(
						_ context.Context,
						_ dogma.Message,
					) (string, bool, error) {
						return "", false, nil
					}
				})

				It("does not forward the message to the handler", func() {
					handler.HandleEventFunc = func(
						context.Context,
						dogma.ProcessEventScope,
						dogma.Message,
					) error {
						Fail("unexpected call to HandleEvent()")
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

				It("records a fact", func() {
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

		When("handling a timeout", func() {
			BeforeEach(func() {
				handler.HandleEventFunc = func(
					_ context.Context,
					s dogma.ProcessEventScope,
					_ dogma.Message,
				) error {
					s.Begin()
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

			It("forwards the message to the handler", func() {
				called := false
				handler.HandleTimeoutFunc = func(
					_ context.Context,
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

			It("propagates handler errors", func() {
				expected := errors.New("<error>")

				handler.HandleTimeoutFunc = func(
					_ context.Context,
					_ dogma.ProcessTimeoutScope,
					_ dogma.Message,
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

			It("returns both commands and timeouts", func() {
				now := time.Now()

				handler.HandleTimeoutFunc = func(
					_ context.Context,
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

			It("returns timeouts scheduled in the past", func() {
				now := time.Now()

				handler.HandleTimeoutFunc = func(
					_ context.Context,
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

			It("does not return timeouts scheduled in the future", func() {
				now := time.Now()

				handler.HandleTimeoutFunc = func(
					_ context.Context,
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

			It("uses the handler's timeout hint", func() {
				hint := 3 * time.Second
				handler.TimeoutHintFunc = func(dogma.Message) time.Duration {
					return hint
				}

				handler.HandleTimeoutFunc = func(
					ctx context.Context,
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

			When("the instance that created the timeout no longer exists", func() {
				BeforeEach(func() {
					handler.HandleEventFunc = func(
						_ context.Context,
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

				It("does not forward the message to the handler", func() {
					handler.HandleTimeoutFunc = func(
						context.Context,
						dogma.ProcessTimeoutScope,
						dogma.Message,
					) error {
						Fail("unexpected call to HandleEvent()")
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

				It("records a fact", func() {
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

		It("propagates routing errors", func() {
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

		It("panics when the handler routes to an empty instance ID", func() {
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
			}).To(PanicWith("the '<name>' process message handler attempted to route a fixtures.MessageE event to an empty instance ID"))
		})

		When("the instance does not exist", func() {
			It("records a fact", func() {
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
			})

			It("panics if New() returns nil", func() {
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
				}).To(PanicWith("the '<name>' process message handler returned a nil root from New()"))
			})
		})

		When("the instance exists", func() {
			BeforeEach(func() {
				handler.HandleEventFunc = func(
					_ context.Context,
					s dogma.ProcessEventScope,
					_ dogma.Message,
				) error {
					s.Begin()
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

			It("records a fact", func() {
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

			It("does not call New()", func() {
				handler.NewFunc = func() dogma.ProcessRoot {
					Fail("unexpected call to New()")
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

		It("provides more context to UnexpectedMessage panics from RouteEventToInstance()", func() {
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

		It("provides more context to UnexpectedMessage panics from HandleEvent()", func() {
			handler.HandleEventFunc = func(
				context.Context,
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

		It("provides more context to UnexpectedMessage panics from HandleTimeout()", func() {
			handler.HandleEventFunc = func(
				_ context.Context,
				s dogma.ProcessEventScope,
				_ dogma.Message,
			) error {
				s.Begin()
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

		It("provides more context to UnexpectedMessage panics from TimeoutHint()", func() {
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

	Describe("func Reset()", func() {
		BeforeEach(func() {
			handler.HandleEventFunc = func(
				_ context.Context,
				s dogma.ProcessEventScope,
				m dogma.Message,
			) error {
				s.Begin()
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

		It("removes all instances", func() {
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
