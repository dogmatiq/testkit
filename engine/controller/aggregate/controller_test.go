package aggregate_test

import (
	"context"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogmatest/engine/controller/aggregate"
	"github.com/dogmatiq/dogmatest/engine/envelope"
	"github.com/dogmatiq/dogmatest/engine/fact"
	"github.com/dogmatiq/dogmatest/internal/enginekit/fixtures"
	handlerkit "github.com/dogmatiq/dogmatest/internal/enginekit/handler"
	"github.com/dogmatiq/dogmatest/internal/enginekit/message"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("type Controller", func() {
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

	Describe("func Name()", func() {
		It("returns the handler name", func() {
			Expect(controller.Name()).To(Equal("<name>"))
		})
	})

	Describe("func Type()", func() {
		It("returns handler.AggregateType", func() {
			Expect(controller.Type()).To(Equal(handlerkit.AggregateType))
		})
	})

	Describe("func Handle()", func() {
		It("forwards the message to the handler", func() {
			called := false
			handler.HandleCommandFunc = func(
				_ dogma.AggregateCommandScope,
				hm dogma.Message,
			) {
				called = true
				Expect(hm).To(Equal(fixtures.MessageA1))
			}

			_, err := controller.Handle(
				context.Background(),
				fact.Ignore,
				command,
			)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(called).To(BeTrue())
		})

		When("the handler returns an empty instance ID", func() {
			It("panics when the handler routes to an empty instance ID", func() {
				handler.RouteCommandToInstanceFunc = func(m dogma.Message) string {
					return ""
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

		When("the instance does not already exist", func() {
			It("records a fact", func() {
				buf := &fact.Buffer{}
				_, err := controller.Handle(
					context.Background(),
					buf,
					command,
				)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(buf.Facts).To(ContainElement(
					fact.AggregateInstanceNotFound{
						HandlerName: "<name>",
						InstanceID:  "<instance>",
						Envelope:    command,
					},
				))
			})

			It("panics if New() returns nil", func() {
				handler.NewFunc = func() dogma.AggregateRoot {
					return nil
				}

				Expect(func() {
					controller.Handle(
						context.Background(),
						fact.Ignore,
						command,
					)
				}).To(Panic())
			})

			When("the handler does not create the instance first", func() {
				It("panics if the handler access the root", func() {
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

				It("panics if the handler records an event", func() {
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

				It("panics if the handler destroys the instance", func() {
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
		})

		When("the handler records an event", func() {
			event := command.NewEvent(
				fixtures.MessageB1,
			)

			BeforeEach(func() {
				handler.HandleCommandFunc = func(
					s dogma.AggregateCommandScope,
					_ dogma.Message,
				) {
					s.Create() // must be created first
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
						HandlerName:   "<name>",
						InstanceID:    "<instance>",
						Root:          &fixtures.AggregateRoot{},
						Envelope:      command,
						EventEnvelope: event,
					},
				))
			})
		})

		When("the handler logs a message", func() {
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
})
