package aggregate_test

import (
	"context"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogmatest/engine/controller/aggregate"
	"github.com/dogmatiq/dogmatest/engine/envelope"
	"github.com/dogmatiq/dogmatest/engine/fact"
	"github.com/dogmatiq/dogmatest/internal/enginekit/fixtures"
	"github.com/dogmatiq/dogmatest/internal/enginekit/message"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("type commandScope", func() {
	var (
		handler    *fixtures.AggregateMessageHandler
		controller *Controller
		command    = envelope.New(
			fixtures.MessageA1,
			message.CommandRole,
		)
	)

	BeforeEach(func() {
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
		controller = NewController("<name>", handler)
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
					command,
				)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(buf.Facts).To(ContainElement(
					fact.AggregateInstanceCreated{
						HandlerName: "<name>",
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
				envelope.New(
					fixtures.MessageA2, // use a different message to create the instance
					message.CommandRole,
				),
			)

			Expect(err).ShouldNot(HaveOccurred())
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
					command,
				)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(buf.Facts).NotTo(ContainElement(
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
					command,
				)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(buf.Facts).To(ContainElement(
					fact.AggregateInstanceDestroyed{
						HandlerName: "<name>",
						InstanceID:  "<instance>",
						Root:        &fixtures.AggregateRoot{},
						Envelope:    command,
					},
				))
			})
		})

		Describe("func RecordEvent", func() {
			BeforeEach(func() {
				fn := handler.HandleCommandFunc
				handler.HandleCommandFunc = func(
					s dogma.AggregateCommandScope,
					m dogma.Message,
				) {
					fn(s, m)
					s.RecordEvent(fixtures.MessageB1)
				}
			})

			It("records a fact", func() {
				buf := &fact.Buffer{}
				_, err := controller.Handle(
					context.Background(),
					buf,
					command,
				)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(buf.Facts).To(ContainElement(
					fact.EventRecordedByAggregate{
						HandlerName: "<name>",
						InstanceID:  "<instance>",
						Root:        &fixtures.AggregateRoot{},
						Envelope:    command,
						EventEnvelope: command.NewEvent(
							fixtures.MessageB1,
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
				command,
			)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(buf.Facts).To(ContainElement(
				fact.MessageLoggedByAggregate{
					HandlerName: "<name>",
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
