package process_test

import (
	"context"
	"errors"
	"time"

	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogma/fixtures"
	. "github.com/dogmatiq/testkit/engine/controller/process"
	"github.com/dogmatiq/testkit/engine/envelope"
	"github.com/dogmatiq/testkit/engine/fact"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("type scope", func() {
	var (
		messageIDs envelope.MessageIDGenerator
		handler    *ProcessMessageHandler
		config     configkit.RichProcess
		ctrl       *Controller
		event      *envelope.Envelope
	)

	BeforeEach(func() {
		event = envelope.NewEvent(
			"1000",
			MessageA1,
			time.Now(),
		)

		handler = &ProcessMessageHandler{
			ConfigureFunc: func(c dogma.ProcessConfigurer) {
				c.Identity("<name>", "<key>")
				c.ConsumesEventType(MessageE{})
				c.ProducesCommandType(MessageC{})
				c.SchedulesTimeoutType(MessageT{})
			},
			RouteEventToInstanceFunc: func(
				_ context.Context,
				m dogma.Message,
			) (string, bool, error) {
				switch m.(type) {
				case MessageA:
					return "<instance>", true, nil
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

	When("the instance does not exist", func() {
		Describe("func HasBegun()", func() {
			It("returns false", func() {
				handler.HandleEventFunc = func(
					_ context.Context,
					s dogma.ProcessEventScope,
					_ dogma.Message,
				) error {
					Expect(s.HasBegun()).To(BeFalse())
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
		})

		Describe("func Root()", func() {
			It("panics", func() {
				handler.HandleEventFunc = func(
					_ context.Context,
					s dogma.ProcessEventScope,
					_ dogma.Message,
				) error {
					s.Root()
					return nil
				}

				Expect(func() {
					ctrl.Handle(
						context.Background(),
						fact.Ignore,
						time.Now(),
						event,
					)
				}).To(PanicWith("can not access process root of non-existent instance"))
			})
		})

		Describe("func Begin()", func() {
			It("returns true", func() {
				handler.HandleEventFunc = func(
					_ context.Context,
					s dogma.ProcessEventScope,
					_ dogma.Message,
				) error {
					Expect(s.Begin()).To(BeTrue())
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
				handler.HandleEventFunc = func(
					_ context.Context,
					s dogma.ProcessEventScope,
					_ dogma.Message,
				) error {
					s.Begin()
					return nil
				}

				buf := &fact.Buffer{}
				_, err := ctrl.Handle(
					context.Background(),
					buf,
					time.Now(),
					event,
				)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(buf.Facts()).To(ContainElement(
					fact.ProcessInstanceBegun{
						Handler:    config,
						InstanceID: "<instance>",
						Root:       &ProcessRoot{},
						Envelope:   event,
					},
				))
			})
		})

		Describe("func End()", func() {
			It("panics", func() {
				handler.HandleEventFunc = func(
					_ context.Context,
					s dogma.ProcessEventScope,
					_ dogma.Message,
				) error {
					s.End()
					return nil
				}

				Expect(func() {
					ctrl.Handle(
						context.Background(),
						fact.Ignore,
						time.Now(),
						event,
					)
				}).To(PanicWith("can not end non-existent instance"))
			})
		})

		Describe("func ExecuteCommand()", func() {
			It("panics", func() {
				handler.HandleEventFunc = func(
					_ context.Context,
					s dogma.ProcessEventScope,
					_ dogma.Message,
				) error {
					s.ExecuteCommand(MessageC1)
					return nil
				}

				Expect(func() {
					ctrl.Handle(
						context.Background(),
						fact.Ignore,
						time.Now(),
						event,
					)
				}).To(PanicWith("can not execute command against non-existent instance"))
			})
		})

		Describe("func ScheduleTimeout()", func() {
			It("panics", func() {
				handler.HandleEventFunc = func(
					_ context.Context,
					s dogma.ProcessEventScope,
					_ dogma.Message,
				) error {
					s.ScheduleTimeout(MessageT1, time.Now())
					return nil
				}

				Expect(func() {
					ctrl.Handle(
						context.Background(),
						fact.Ignore,
						time.Now(),
						event,
					)
				}).To(PanicWith("can not schedule timeout against non-existent instance"))
			})
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
				envelope.NewEvent(
					"2000",
					MessageA2, // use a different message to create the instance
					time.Now(),
				),
			)

			Expect(err).ShouldNot(HaveOccurred())

			messageIDs.Reset() // reset after setup for a predictable ID.
		})

		Describe("func HasBegun()", func() {
			It("returns true", func() {
				handler.HandleEventFunc = func(
					_ context.Context,
					s dogma.ProcessEventScope,
					_ dogma.Message,
				) error {
					Expect(s.HasBegun()).To(BeTrue())
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
		})

		Describe("func Root()", func() {
			It("returns the root", func() {
				handler.HandleEventFunc = func(
					_ context.Context,
					s dogma.ProcessEventScope,
					_ dogma.Message,
				) error {
					Expect(s.Root()).To(Equal(&ProcessRoot{}))
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
		})

		Describe("func Begin()", func() {
			It("returns false", func() {
				handler.HandleEventFunc = func(
					_ context.Context,
					s dogma.ProcessEventScope,
					_ dogma.Message,
				) error {
					Expect(s.Begin()).To(BeFalse())
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

			It("does not record a fact", func() {
				handler.HandleEventFunc = func(
					_ context.Context,
					s dogma.ProcessEventScope,
					_ dogma.Message,
				) error {
					s.Begin()
					return nil
				}

				buf := &fact.Buffer{}
				_, err := ctrl.Handle(
					context.Background(),
					buf,
					time.Now(),
					event,
				)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(buf.Facts()).NotTo(ContainElement(
					BeAssignableToTypeOf(fact.ProcessInstanceBegun{}),
				))
			})
		})

		Describe("func End()", func() {
			It("records a fact", func() {
				handler.HandleEventFunc = func(
					_ context.Context,
					s dogma.ProcessEventScope,
					_ dogma.Message,
				) error {
					s.End()
					return nil
				}

				buf := &fact.Buffer{}
				_, err := ctrl.Handle(
					context.Background(),
					buf,
					time.Now(),
					event,
				)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(buf.Facts()).To(ContainElement(
					fact.ProcessInstanceEnded{
						Handler:    config,
						InstanceID: "<instance>",
						Root:       &ProcessRoot{},
						Envelope:   event,
					},
				))
			})
		})

		Describe("func ExecuteCommand()", func() {
			BeforeEach(func() {
				fn := handler.HandleEventFunc
				handler.HandleEventFunc = func(
					ctx context.Context,
					s dogma.ProcessEventScope,
					m dogma.Message,
				) error {
					if err := fn(ctx, s, m); err != nil {
						return err
					}

					s.ExecuteCommand(MessageC1)
					return nil
				}
			})

			It("records a fact", func() {
				buf := &fact.Buffer{}
				now := time.Now()
				_, err := ctrl.Handle(
					context.Background(),
					buf,
					now,
					event,
				)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(buf.Facts()).To(ContainElement(
					fact.CommandExecutedByProcess{
						Handler:    config,
						InstanceID: "<instance>",
						Root:       &ProcessRoot{},
						Envelope:   event,
						CommandEnvelope: event.NewCommand(
							"1",
							MessageC1,
							now,
							envelope.Origin{
								Handler:     config,
								HandlerType: configkit.ProcessHandlerType,
								InstanceID:  "<instance>",
							},
						),
					},
				))
			})

			It("panics if the command type is not configured to be produced", func() {
				handler.HandleEventFunc = func(
					_ context.Context,
					s dogma.ProcessEventScope,
					m dogma.Message,
				) error {
					s.ExecuteCommand(MessageX1)
					return nil
				}

				Expect(func() {
					ctrl.Handle(
						context.Background(),
						fact.Ignore,
						time.Now(),
						event,
					)
				}).To(PanicWith("the '<name>' handler is not configured to execute commands of type fixtures.MessageX"))
			})

			It("panics if the command is invalid", func() {
				handler.HandleEventFunc = func(
					_ context.Context,
					s dogma.ProcessEventScope,
					m dogma.Message,
				) error {
					s.ExecuteCommand(MessageC{
						Value: errors.New("<invalid>"),
					})
					return nil
				}

				Expect(func() {
					ctrl.Handle(
						context.Background(),
						fact.Ignore,
						time.Now(),
						event,
					)
				}).To(PanicWith("can not execute command of type fixtures.MessageC, it is invalid: <invalid>"))
			})
		})

		Describe("func ScheduleTimeout()", func() {
			t := time.Now()

			BeforeEach(func() {
				fn := handler.HandleEventFunc
				handler.HandleEventFunc = func(
					ctx context.Context,
					s dogma.ProcessEventScope,
					m dogma.Message,
				) error {
					if err := fn(ctx, s, m); err != nil {
						return err
					}

					s.ScheduleTimeout(MessageT1, t)
					return nil
				}
			})

			It("records a fact", func() {
				buf := &fact.Buffer{}
				now := time.Now()
				_, err := ctrl.Handle(
					context.Background(),
					buf,
					now,
					event,
				)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(buf.Facts()).To(ContainElement(
					fact.TimeoutScheduledByProcess{
						Handler:    config,
						InstanceID: "<instance>",
						Root:       &ProcessRoot{},
						Envelope:   event,
						TimeoutEnvelope: event.NewTimeout(
							"1",
							MessageT1,
							now,
							t,
							envelope.Origin{
								Handler:     config,
								HandlerType: configkit.ProcessHandlerType,
								InstanceID:  "<instance>",
							},
						),
					},
				))
			})

			It("panics if the timeout type is not configured to be scheduled", func() {
				handler.HandleEventFunc = func(
					_ context.Context,
					s dogma.ProcessEventScope,
					_ dogma.Message,
				) error {
					s.ScheduleTimeout(MessageX1, time.Now())
					return nil
				}

				Expect(func() {
					ctrl.Handle(
						context.Background(),
						fact.Ignore,
						time.Now(),
						event,
					)
				}).To(PanicWith("the '<name>' handler is not configured to schedule timeouts of type fixtures.MessageX"))
			})

			It("panics if the timeout is invalid", func() {
				handler.HandleEventFunc = func(
					_ context.Context,
					s dogma.ProcessEventScope,
					m dogma.Message,
				) error {
					s.ScheduleTimeout(
						MessageT{
							Value: errors.New("<invalid>"),
						},
						time.Now(),
					)
					return nil
				}

				Expect(func() {
					ctrl.Handle(
						context.Background(),
						fact.Ignore,
						time.Now(),
						event,
					)
				}).To(PanicWith("can not schedule timeout of type fixtures.MessageT, it is invalid: <invalid>"))
			})
		})
	})

	Describe("func InstanceID()", func() {
		It("returns the instance ID", func() {
			called := false
			handler.HandleEventFunc = func(
				_ context.Context,
				s dogma.ProcessEventScope,
				_ dogma.Message,
			) error {
				called = true
				Expect(s.InstanceID()).To(Equal("<instance>"))
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
	})

	Describe("func Log()", func() {
		BeforeEach(func() {
			handler.HandleEventFunc = func(
				_ context.Context,
				s dogma.ProcessEventScope,
				_ dogma.Message,
			) error {
				s.Log("<format>", "<arg-1>", "<arg-2>")
				return nil
			}
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
				fact.MessageLoggedByProcess{
					Handler:    config,
					InstanceID: "<instance>",
					Root:       &ProcessRoot{},
					Envelope:   event,
					LogFormat:  "<format>",
					LogArguments: []interface{}{
						"<arg-1>",
						"<arg-2>",
					},
				},
			))
		})
	})
})
