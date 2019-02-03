package aggregate_test

import (
	"context"
	"time"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogmatest/engine/controller/aggregate"
	"github.com/dogmatiq/dogmatest/engine/envelope"
	"github.com/dogmatiq/dogmatest/engine/fact"
	"github.com/dogmatiq/enginekit/fixtures"
	handlerkit "github.com/dogmatiq/enginekit/handler"
	"github.com/dogmatiq/enginekit/message"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("type scope", func() {
	var (
		messageIDs envelope.MessageIDGenerator
		handler    *fixtures.AggregateMessageHandler
		controller *Controller
		command    *envelope.Envelope
	)

	BeforeEach(func() {
		command = envelope.New(
			"1000",
			fixtures.MessageA1,
			message.CommandRole,
		)

		handler = &fixtures.AggregateMessageHandler{
			RouteCommandToInstanceFunc: func(m dogma.Message) string {
				switch m.(type) {
				case fixtures.MessageA:
					return "<instance>"
				default:
					panic(dogma.UnexpectedMessage)
				}
			},
		}

		controller = NewController("<name>", handler, &messageIDs)

		messageIDs.Reset() // reset after setup for a predictable ID.
	})

	When("the instance does not exist", func() {
		Describe("func Root", func() {
			It("panics", func() {
				handler.HandleCommandFunc = func(
					s dogma.AggregateCommandScope,
					_ dogma.Message,
				) {
					s.Root()
				}

				Expect(func() {
					controller.Handle(
						context.Background(),
						fact.Ignore,
						time.Now(),
						command,
					)
				}).To(Panic())
			})
		})

		Describe("func Create", func() {
			It("returns true", func() {
				handler.HandleCommandFunc = func(
					s dogma.AggregateCommandScope,
					_ dogma.Message,
				) {
					Expect(s.Create()).To(BeTrue())
					s.RecordEvent(fixtures.MessageE1) // event must be recorded when creating
				}

				_, err := controller.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					command,
				)

				Expect(err).ShouldNot(HaveOccurred())
			})

			It("records a fact", func() {
				handler.HandleCommandFunc = func(
					s dogma.AggregateCommandScope,
					_ dogma.Message,
				) {
					s.Create()
					s.RecordEvent(fixtures.MessageE1) // event must be recorded when creating
				}

				buf := &fact.Buffer{}
				_, err := controller.Handle(
					context.Background(),
					buf,
					time.Now(),
					command,
				)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(buf.Facts()).To(ContainElement(
					fact.AggregateInstanceCreated{
						HandlerName: "<name>",
						Handler:     handler,
						InstanceID:  "<instance>",
						Root:        &fixtures.AggregateRoot{},
						Envelope:    command,
					},
				))
			})
		})

		Describe("func Destroy", func() {
			It("panics", func() {
				handler.HandleCommandFunc = func(
					s dogma.AggregateCommandScope,
					_ dogma.Message,
				) {
					s.Destroy()
				}

				Expect(func() {
					controller.Handle(
						context.Background(),
						fact.Ignore,
						time.Now(),
						command,
					)
				}).To(Panic())
			})
		})

		Describe("func RecordEvent", func() {
			It("panics", func() {
				handler.HandleCommandFunc = func(
					s dogma.AggregateCommandScope,
					_ dogma.Message,
				) {
					s.RecordEvent(fixtures.MessageB1)
				}

				Expect(func() {
					controller.Handle(
						context.Background(),
						fact.Ignore,
						time.Now(),
						command,
					)
				}).To(Panic())
			})
		})
	})

	When("the instance exists", func() {
		BeforeEach(func() {
			handler.HandleCommandFunc = func(
				s dogma.AggregateCommandScope,
				_ dogma.Message,
			) {
				s.Create()
				s.RecordEvent(fixtures.MessageE1) // event must be recorded when creating
			}

			_, err := controller.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				envelope.New(
					"2000",
					fixtures.MessageA2, // use a different message to create the instance
					message.CommandRole,
				),
			)

			Expect(err).ShouldNot(HaveOccurred())

			messageIDs.Reset() // reset after setup for a predictable ID.
		})

		Describe("func Root", func() {
			It("returns the root", func() {
				handler.HandleCommandFunc = func(
					s dogma.AggregateCommandScope,
					_ dogma.Message,
				) {
					Expect(s.Root()).To(Equal(&fixtures.AggregateRoot{}))
				}

				_, err := controller.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					command,
				)

				Expect(err).ShouldNot(HaveOccurred())
			})
		})

		Describe("func Create", func() {
			It("returns false", func() {
				handler.HandleCommandFunc = func(
					s dogma.AggregateCommandScope,
					_ dogma.Message,
				) {
					Expect(s.Create()).To(BeFalse())
					s.RecordEvent(fixtures.MessageE1) // event must be recorded when creating
				}

				_, err := controller.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					command,
				)

				Expect(err).ShouldNot(HaveOccurred())
			})

			It("does not record a fact", func() {
				handler.HandleCommandFunc = func(
					s dogma.AggregateCommandScope,
					_ dogma.Message,
				) {
					s.Create()
					s.RecordEvent(fixtures.MessageE1) // event must be recorded when creating
				}

				buf := &fact.Buffer{}
				_, err := controller.Handle(
					context.Background(),
					buf,
					time.Now(),
					command,
				)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(buf.Facts()).NotTo(ContainElement(
					BeAssignableToTypeOf(fact.AggregateInstanceCreated{}),
				))
			})
		})

		Describe("func Destroy", func() {
			It("records a fact", func() {
				handler.HandleCommandFunc = func(
					s dogma.AggregateCommandScope,
					_ dogma.Message,
				) {
					s.RecordEvent(fixtures.MessageE1) // event must be recorded when destroying
					s.Destroy()
				}

				buf := &fact.Buffer{}
				_, err := controller.Handle(
					context.Background(),
					buf,
					time.Now(),
					command,
				)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(buf.Facts()).To(ContainElement(
					fact.AggregateInstanceDestroyed{
						HandlerName: "<name>",
						Handler:     handler,
						InstanceID:  "<instance>",
						Root:        &fixtures.AggregateRoot{},
						Envelope:    command,
					},
				))
			})
		})

		Describe("func RecordEvent", func() {
			BeforeEach(func() {
				handler.HandleCommandFunc = func(
					s dogma.AggregateCommandScope,
					m dogma.Message,
				) {
					s.RecordEvent(fixtures.MessageB1)
				}
			})

			It("records a fact", func() {
				messageIDs.Reset() // reset after setup for a predictable ID.

				buf := &fact.Buffer{}
				_, err := controller.Handle(
					context.Background(),
					buf,
					time.Now(),
					command,
				)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(buf.Facts()).To(ContainElement(
					fact.EventRecordedByAggregate{
						HandlerName: "<name>",
						Handler:     handler,
						InstanceID:  "<instance>",
						Root:        &fixtures.AggregateRoot{},
						Envelope:    command,
						EventEnvelope: command.NewEvent(
							"1",
							fixtures.MessageB1,
							envelope.Origin{
								HandlerName: "<name>",
								HandlerType: handlerkit.AggregateType,
								InstanceID:  "<instance>",
							},
						),
					},
				))
			})
		})
	})

	Describe("func InstanceID", func() {
		It("returns the instance ID", func() {
			called := false
			handler.HandleCommandFunc = func(
				s dogma.AggregateCommandScope,
				_ dogma.Message,
			) {
				called = true
				Expect(s.InstanceID()).To(Equal("<instance>"))
			}

			_, err := controller.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				command,
			)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(called).To(BeTrue())
		})
	})

	Describe("func Log", func() {
		BeforeEach(func() {
			handler.HandleCommandFunc = func(
				s dogma.AggregateCommandScope,
				_ dogma.Message,
			) {
				s.Log("<format>", "<arg-1>", "<arg-2>")
			}
		})

		It("records a fact", func() {
			buf := &fact.Buffer{}
			_, err := controller.Handle(
				context.Background(),
				buf,
				time.Now(),
				command,
			)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(buf.Facts()).To(ContainElement(
				fact.MessageLoggedByAggregate{
					HandlerName: "<name>",
					Handler:     handler,
					InstanceID:  "<instance>",
					Root:        &fixtures.AggregateRoot{},
					Envelope:    command,
					LogFormat:   "<format>",
					LogArguments: []interface{}{
						"<arg-1>",
						"<arg-2>",
					},
				},
			))
		})
	})
})
