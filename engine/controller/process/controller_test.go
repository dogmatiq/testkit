package process_test

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/dogmatest/engine/controller"
	. "github.com/dogmatiq/dogmatest/engine/controller/process"
	"github.com/dogmatiq/dogmatest/engine/envelope"
	"github.com/dogmatiq/dogmatest/engine/fact"
	"github.com/dogmatiq/enginekit/fixtures"
	handlerkit "github.com/dogmatiq/enginekit/handler"
	"github.com/dogmatiq/enginekit/message"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ controller.Controller = &Controller{}

var _ = Describe("type Controller", func() {
	var (
		now        = time.Now()
		handler    *fixtures.ProcessMessageHandler
		controller *Controller
		event      = envelope.New(
			fixtures.MessageE1,
			message.EventRole,
		)
		timeout = event.NewTimeout(
			fixtures.MessageT1,
			now,
			envelope.Origin{
				HandlerName: "<name>",
				HandlerType: handlerkit.ProcessType,
				InstanceID:  "<instance-E1>",
			},
		)
	)

	BeforeEach(func() {
		handler = &fixtures.ProcessMessageHandler{
			// setup routes for "E" (event) messages to an instance ID based on the
			// message's content
			RouteEventToInstanceFunc: func(
				_ context.Context,
				m dogma.Message,
			) (string, bool, error) {
				switch x := m.(type) {
				case fixtures.MessageE:
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
		controller = NewController("<name>", handler)
	})

	Describe("func Name()", func() {
		It("returns the handler name", func() {
			Expect(controller.Name()).To(Equal("<name>"))
		})
	})

	Describe("func Type()", func() {
		It("returns handler.ProcessType", func() {
			Expect(controller.Type()).To(Equal(handlerkit.ProcessType))
		})
	})

	Describe("func Tick()", func() {
		BeforeEach(func() {
			handler.HandleEventFunc = func(
				_ context.Context,
				s dogma.ProcessEventScope,
				_ dogma.Message,
			) error {
				s.Begin()

				// note, calls to ScheduleTimeout are NOT in chronological order
				s.ScheduleTimeout(fixtures.MessageT3, now.Add(3*time.Hour))
				s.ScheduleTimeout(fixtures.MessageT2, now.Add(2*time.Hour))
				s.ScheduleTimeout(fixtures.MessageT1, now.Add(1*time.Hour))

				return nil
			}

			_, err := controller.Handle(
				context.Background(),
				fact.Ignore,
				now,
				event,
			)
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("returns timeouts that are ready to be handled", func() {
			timeouts, err := controller.Tick(
				context.Background(),
				fact.Ignore,
				now.Add(2*time.Hour),
			)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(timeouts).To(ConsistOf(
				event.NewTimeout(
					fixtures.MessageT1,
					now.Add(1*time.Hour),
					envelope.Origin{
						HandlerName: "<name>",
						HandlerType: handlerkit.ProcessType,
						InstanceID:  "<instance-E1>",
					},
				),
				event.NewTimeout(
					fixtures.MessageT2,
					now.Add(2*time.Hour),
					envelope.Origin{
						HandlerName: "<name>",
						HandlerType: handlerkit.ProcessType,
						InstanceID:  "<instance-E1>",
					},
				),
			))
		})

		It("does not return the same timeouts multiple times", func() {
			t := now.Add(2 * time.Hour)

			timeouts, err := controller.Tick(
				context.Background(),
				fact.Ignore,
				t,
			)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(timeouts).To(HaveLen(2))

			timeouts, err = controller.Tick(
				context.Background(),
				fact.Ignore,
				t,
			)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(timeouts).To(BeEmpty())
		})

		It("does not return timeouts for instances that have been deleted", func() {
			_, err := controller.Handle(
				context.Background(),
				fact.Ignore,
				now,
				envelope.New(
					fixtures.MessageE2, // different message value = different instance
					message.EventRole,
				),
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

			_, err = controller.Handle(
				context.Background(),
				fact.Ignore,
				now,
				event,
			)
			Expect(err).ShouldNot(HaveOccurred())

			// expect only the timeout from the E2 instance.
			timeouts, err := controller.Tick(
				context.Background(),
				fact.Ignore,
				now.Add(2*time.Hour),
			)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(timeouts).To(ConsistOf(
				event.NewTimeout(
					fixtures.MessageT1,
					now.Add(1*time.Hour),
					envelope.Origin{
						HandlerName: "<name>",
						HandlerType: handlerkit.ProcessType,
						InstanceID:  "<instance-E2>", // E2, not E1!
					},
				),
				event.NewTimeout(
					fixtures.MessageT2,
					now.Add(2*time.Hour),
					envelope.Origin{
						HandlerName: "<name>",
						HandlerType: handlerkit.ProcessType,
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
					Expect(m).To(Equal(fixtures.MessageE1))
					return nil
				}

				_, err := controller.Handle(
					context.Background(),
					fact.Ignore,
					now,
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

				_, err := controller.Handle(
					context.Background(),
					fact.Ignore,
					now,
					event,
				)

				Expect(err).To(Equal(expected))
			})

			It("returns both commands and timeouts", func() {
				handler.HandleEventFunc = func(
					_ context.Context,
					s dogma.ProcessEventScope,
					_ dogma.Message,
				) error {
					s.Begin()
					s.ExecuteCommand(fixtures.MessageC1)
					s.ScheduleTimeout(fixtures.MessageT1, now) // timeouts at current time are "ready"
					return nil
				}

				envelopes, err := controller.Handle(
					context.Background(),
					fact.Ignore,
					now,
					event,
				)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(envelopes).To(ConsistOf(
					event.NewCommand(
						fixtures.MessageC1,
						envelope.Origin{
							HandlerName: "<name>",
							HandlerType: handlerkit.ProcessType,
							InstanceID:  "<instance-E1>",
						},
					),
					event.NewTimeout(
						fixtures.MessageT1,
						now,
						envelope.Origin{
							HandlerName: "<name>",
							HandlerType: handlerkit.ProcessType,
							InstanceID:  "<instance-E1>",
						},
					),
				))
			})

			It("returns timeouts scheduled in the past", func() {
				handler.HandleEventFunc = func(
					_ context.Context,
					s dogma.ProcessEventScope,
					_ dogma.Message,
				) error {
					s.Begin()
					s.ScheduleTimeout(fixtures.MessageT1, now.Add(-1))
					return nil
				}

				envelopes, err := controller.Handle(
					context.Background(),
					fact.Ignore,
					now,
					event,
				)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(envelopes).To(HaveLen(1))
			})

			It("does not return timeouts scheduled in the future", func() {
				handler.HandleEventFunc = func(
					_ context.Context,
					s dogma.ProcessEventScope,
					_ dogma.Message,
				) error {
					s.Begin()
					s.ScheduleTimeout(fixtures.MessageT1, now.Add(1))
					return nil
				}

				envelopes, err := controller.Handle(
					context.Background(),
					fact.Ignore,
					now,
					event,
				)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(envelopes).To(BeEmpty())
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

					_, err := controller.Handle(
						context.Background(),
						fact.Ignore,
						now,
						event,
					)

					Expect(err).ShouldNot(HaveOccurred())
				})

				It("records a fact", func() {
					buf := &fact.Buffer{}
					_, err := controller.Handle(
						context.Background(),
						buf,
						now,
						event,
					)

					Expect(err).ShouldNot(HaveOccurred())
					Expect(buf.Facts).To(ContainElement(
						fact.ProcessEventIgnored{
							HandlerName: "<name>",
							Envelope:    event,
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

				_, err := controller.Handle(
					context.Background(),
					fact.Ignore,
					now,
					event,
				)
				Expect(err).ShouldNot(HaveOccurred())
			})

			It("forwards the message to the handler", func() {
				called := false
				handler.HandleTimeoutFunc = func(
					_ context.Context,
					_ dogma.ProcessTimeoutScope,
					m dogma.Message,
				) error {
					called = true
					Expect(m).To(Equal(fixtures.MessageT1))
					return nil
				}

				_, err := controller.Handle(
					context.Background(),
					fact.Ignore,
					now,
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

				_, err := controller.Handle(
					context.Background(),
					fact.Ignore,
					now,
					timeout,
				)

				Expect(err).To(Equal(expected))
			})

			It("returns both commands and timeouts", func() {
				handler.HandleTimeoutFunc = func(
					_ context.Context,
					s dogma.ProcessTimeoutScope,
					_ dogma.Message,
				) error {
					s.ExecuteCommand(fixtures.MessageC1)
					s.ScheduleTimeout(fixtures.MessageT1, now) // timeouts at current time are "ready"
					return nil
				}

				envelopes, err := controller.Handle(
					context.Background(),
					fact.Ignore,
					now,
					timeout,
				)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(envelopes).To(ConsistOf(
					event.NewCommand(
						fixtures.MessageC1,
						envelope.Origin{
							HandlerName: "<name>",
							HandlerType: handlerkit.ProcessType,
							InstanceID:  "<instance-E1>",
						},
					),
					event.NewTimeout(
						fixtures.MessageT1,
						now,
						envelope.Origin{
							HandlerName: "<name>",
							HandlerType: handlerkit.ProcessType,
							InstanceID:  "<instance-E1>",
						},
					),
				))
			})

			It("returns timeouts scheduled in the past", func() {
				handler.HandleTimeoutFunc = func(
					_ context.Context,
					s dogma.ProcessTimeoutScope,
					_ dogma.Message,
				) error {
					s.ScheduleTimeout(fixtures.MessageB2, now.Add(-1))
					return nil
				}

				envelopes, err := controller.Handle(
					context.Background(),
					fact.Ignore,
					now,
					timeout,
				)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(envelopes).To(HaveLen(1))
			})

			It("does not return timeouts scheduled in the future", func() {
				handler.HandleTimeoutFunc = func(
					_ context.Context,
					s dogma.ProcessTimeoutScope,
					_ dogma.Message,
				) error {
					s.ScheduleTimeout(fixtures.MessageB2, now.Add(1))
					return nil
				}

				envelopes, err := controller.Handle(
					context.Background(),
					fact.Ignore,
					now,
					event,
				)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(envelopes).To(BeEmpty())
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

					_, err := controller.Handle(
						context.Background(),
						fact.Ignore,
						now,
						event,
					)
					Expect(err).ShouldNot(HaveOccurred())
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

					_, err := controller.Handle(
						context.Background(),
						fact.Ignore,
						now,
						timeout,
					)

					Expect(err).ShouldNot(HaveOccurred())
				})

				It("records a fact", func() {
					buf := &fact.Buffer{}
					_, err := controller.Handle(
						context.Background(),
						buf,
						now,
						timeout,
					)

					Expect(err).ShouldNot(HaveOccurred())
					Expect(buf.Facts).To(ContainElement(
						fact.ProcessTimeoutIgnored{
							HandlerName: "<name>",
							Envelope:    timeout,
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

			_, err := controller.Handle(
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
				controller.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					event,
				)
			}).To(Panic())
		})

		When("the instance does not exist", func() {
			It("records a fact", func() {
				buf := &fact.Buffer{}
				_, err := controller.Handle(
					context.Background(),
					buf,
					time.Now(),
					event,
				)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(buf.Facts).To(ContainElement(
					fact.ProcessInstanceNotFound{
						HandlerName: "<name>",
						InstanceID:  "<instance-E1>",
						Envelope:    event,
					},
				))
			})

			It("panics if New() returns nil", func() {
				handler.NewFunc = func() dogma.ProcessRoot {
					return nil
				}

				Expect(func() {
					controller.Handle(
						context.Background(),
						fact.Ignore,
						time.Now(),
						event,
					)
				}).To(Panic())
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

				_, err := controller.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					event,
				)

				Expect(err).ShouldNot(HaveOccurred())
			})

			It("records a fact", func() {
				buf := &fact.Buffer{}
				_, err := controller.Handle(
					context.Background(),
					buf,
					time.Now(),
					event,
				)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(buf.Facts).To(ContainElement(
					fact.ProcessInstanceLoaded{
						HandlerName: "<name>",
						InstanceID:  "<instance-E1>",
						Root:        &fixtures.ProcessRoot{},
						Envelope:    event,
					},
				))
			})

			It("does not call New()", func() {
				handler.NewFunc = func() dogma.ProcessRoot {
					Fail("unexpected call to New()")
					return nil
				}

				controller.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					event,
				)
			})
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

			_, err := controller.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				event,
			)

			Expect(err).ShouldNot(HaveOccurred())
		})

		It("removes all instances", func() {
			controller.Reset()

			buf := &fact.Buffer{}
			_, err := controller.Handle(
				context.Background(),
				buf,
				time.Now(),
				event,
			)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(buf.Facts).NotTo(ContainElement(
				BeAssignableToTypeOf(fact.ProcessInstanceLoaded{}),
			))
		})
	})
})
