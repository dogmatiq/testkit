package aggregate_test

import (
	"context"
	"fmt"
	"time"

	"github.com/dogmatiq/configkit"
	. "github.com/dogmatiq/configkit/fixtures"
	"github.com/dogmatiq/configkit/message"
	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogma/fixtures"
	"github.com/dogmatiq/testkit/engine/controller"
	. "github.com/dogmatiq/testkit/engine/controller/aggregate"
	"github.com/dogmatiq/testkit/engine/envelope"
	"github.com/dogmatiq/testkit/engine/fact"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ controller.Controller = &Controller{}

var _ = Describe("type Controller", func() {
	var (
		messageIDs envelope.MessageIDGenerator
		handler    *AggregateMessageHandler
		controller *Controller
		command    *envelope.Envelope
	)

	BeforeEach(func() {
		command = envelope.NewCommand(
			"1000",
			MessageC1,
			time.Now(),
		)

		handler = &AggregateMessageHandler{
			ConfigureFunc: func(c dogma.AggregateConfigurer) {
				c.Identity("<name>", "<key>")
				c.ConsumesCommandType(MessageC{})
				c.ProducesEventType(MessageX{})
			},
			// setup routes for "C" (command) messages to an instance ID based on the
			// message's content
			RouteCommandToInstanceFunc: func(m dogma.Message) string {
				switch x := m.(type) {
				case MessageC:
					return fmt.Sprintf(
						"<instance-%s>",
						x.Value.(string),
					)
				default:
					panic(dogma.UnexpectedMessage)
				}
			},
		}

		config := configkit.FromAggregate(handler)

		controller = NewController(
			config,
			&messageIDs,
			message.NewTypeSet(
				MessageEType,
			),
		)

		messageIDs.Reset() // reset after setup for a predictable ID.
	})

	Describe("func Identity()", func() {
		It("returns the handler identity", func() {
			Expect(controller.Identity()).To(Equal(
				configkit.MustNewIdentity("<name>", "<key>"),
			))
		})
	})

	Describe("func Type()", func() {
		It("returns configkit.AggregateHandlerType", func() {
			Expect(controller.Type()).To(Equal(configkit.AggregateHandlerType))
		})
	})

	Describe("func Tick()", func() {
		It("does not return any envelopes", func() {
			envelopes, err := controller.Tick(
				context.Background(),
				fact.Ignore,
				time.Now(),
			)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(envelopes).To(BeEmpty())
		})

		It("does not record any facts", func() {
			buf := &fact.Buffer{}
			_, err := controller.Tick(
				context.Background(),
				buf,
				time.Now(),
			)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(buf.Facts()).To(BeEmpty())
		})
	})

	Describe("func Handle()", func() {
		It("forwards the message to the handler", func() {
			called := false
			handler.HandleCommandFunc = func(
				_ dogma.AggregateCommandScope,
				m dogma.Message,
			) {
				called = true
				Expect(m).To(Equal(MessageC1))
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

		It("returns the recorded events", func() {
			handler.HandleCommandFunc = func(
				s dogma.AggregateCommandScope,
				_ dogma.Message,
			) {
				s.Create()
				s.RecordEvent(MessageE1)
				s.RecordEvent(MessageE2)
			}

			now := time.Now()
			events, err := controller.Handle(
				context.Background(),
				fact.Ignore,
				now,
				command,
			)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(events).To(ConsistOf(
				command.NewEvent(
					"1",
					MessageE1,
					now,
					envelope.Origin{
						HandlerName: "<name>",
						HandlerType: configkit.AggregateHandlerType,
						InstanceID:  "<instance-C1>",
					},
				),
				command.NewEvent(
					"2",
					MessageE2,
					now,
					envelope.Origin{
						HandlerName: "<name>",
						HandlerType: configkit.AggregateHandlerType,
						InstanceID:  "<instance-C1>",
					},
				),
			))
		})

		It("panics when the handler routes to an empty instance ID", func() {
			handler.RouteCommandToInstanceFunc = func(dogma.Message) string {
				return ""
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

		When("the instance does not exist", func() {
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
					fact.AggregateInstanceNotFound{
						HandlerName: "<name>",
						Handler:     handler,
						InstanceID:  "<instance-C1>",
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
						time.Now(),
						command,
					)
				}).To(Panic())
			})

			It("panics if the instance is created without recording an event", func() {
				handler.HandleCommandFunc = func(
					s dogma.AggregateCommandScope,
					_ dogma.Message,
				) {
					s.Create()
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

		When("the instance exists", func() {
			BeforeEach(func() {
				handler.HandleCommandFunc = func(
					s dogma.AggregateCommandScope,
					_ dogma.Message,
				) {
					s.Create()
					s.RecordEvent(MessageE1) // event must be recorded when creating
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
				buf := &fact.Buffer{}
				_, err := controller.Handle(
					context.Background(),
					buf,
					time.Now(),
					command,
				)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(buf.Facts()).To(ContainElement(
					fact.AggregateInstanceLoaded{
						HandlerName: "<name>",
						Handler:     handler,
						InstanceID:  "<instance-C1>",
						Root:        &AggregateRoot{},
						Envelope:    command,
					},
				))
			})

			It("panics if the instance is destroyed without recording an event", func() {
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
	})

	Describe("func Reset()", func() {
		BeforeEach(func() {
			handler.HandleCommandFunc = func(
				s dogma.AggregateCommandScope,
				m dogma.Message,
			) {
				s.Create()
				s.RecordEvent(MessageE1) // event must be recorded when creating
			}

			_, err := controller.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				command,
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
				command,
			)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(buf.Facts()).NotTo(ContainElement(
				BeAssignableToTypeOf(fact.AggregateInstanceLoaded{}),
			))
		})
	})
})
