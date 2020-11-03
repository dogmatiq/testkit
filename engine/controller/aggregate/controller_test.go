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
	. "github.com/onsi/gomega/gstruct"
)

var _ controller.Controller = &Controller{}

var _ = Describe("type Controller", func() {
	var (
		messageIDs envelope.MessageIDGenerator
		handler    *AggregateMessageHandler
		config     configkit.RichAggregate
		ctrl       *Controller
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

		config = configkit.FromAggregate(handler)

		ctrl = NewController(
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
			Expect(ctrl.Identity()).To(Equal(
				configkit.MustNewIdentity("<name>", "<key>"),
			))
		})
	})

	Describe("func Type()", func() {
		It("returns configkit.AggregateHandlerType", func() {
			Expect(ctrl.Type()).To(Equal(configkit.AggregateHandlerType))
		})
	})

	Describe("func Tick()", func() {
		It("does not return any envelopes", func() {
			envelopes, err := ctrl.Tick(
				context.Background(),
				fact.Ignore,
				time.Now(),
			)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(envelopes).To(BeEmpty())
		})

		It("does not record any facts", func() {
			buf := &fact.Buffer{}
			_, err := ctrl.Tick(
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

			_, err := ctrl.Handle(
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
				s.RecordEvent(MessageE1)
				s.RecordEvent(MessageE2)
			}

			now := time.Now()
			events, err := ctrl.Handle(
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
				ctrl.Handle(
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
				_, err := ctrl.Handle(
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
					ctrl.Handle(
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
					s.RecordEvent(MessageE1) // record event to create the instance
				}

				_, err := ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					command,
				)

				Expect(err).ShouldNot(HaveOccurred())
			})

			It("records a fact", func() {
				buf := &fact.Buffer{}
				_, err := ctrl.Handle(
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
					ctrl.Handle(
						context.Background(),
						fact.Ignore,
						time.Now(),
						command,
					)
				}).To(Panic())
			})
		})

		It("provides more context to UnexpectedMessage panics from RouteCommandToInstance()", func() {
			handler.RouteCommandToInstanceFunc = func(dogma.Message) string {
				panic(dogma.UnexpectedMessage)
			}

			Expect(func() {
				ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					command,
				)
			}).To(PanicWith(
				MatchFields(
					IgnoreExtras,
					Fields{
						"Handler":   Equal(config),
						"Interface": Equal("AggregateMessageHandler"),
						"Method":    Equal("RouteCommandToInstance"),
						"Message":   Equal(command.Message),
					},
				),
			))
		})

		It("provides more context to UnexpectedMessage panics from HandleCommand()", func() {
			handler.HandleCommandFunc = func(
				dogma.AggregateCommandScope,
				dogma.Message,
			) {
				panic(dogma.UnexpectedMessage)
			}

			Expect(func() {
				ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					command,
				)
			}).To(PanicWith(
				MatchFields(
					IgnoreExtras,
					Fields{
						"Handler":   Equal(config),
						"Interface": Equal("AggregateMessageHandler"),
						"Method":    Equal("HandleCommand"),
						"Message":   Equal(command.Message),
					},
				),
			))
		})

		It("provides more context to UnexpectedMessage panics from ApplyEvent() when called with new events", func() {
			handler.HandleCommandFunc = func(
				s dogma.AggregateCommandScope,
				_ dogma.Message,
			) {
				s.RecordEvent(MessageE1)
			}

			handler.NewFunc = func() dogma.AggregateRoot {
				return &AggregateRoot{
					ApplyEventFunc: func(dogma.Message, interface{}) {
						panic(dogma.UnexpectedMessage)
					},
				}
			}

			Expect(func() {
				ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					command,
				)
			}).To(PanicWith(
				MatchFields(
					IgnoreExtras,
					Fields{
						"Handler":   Equal(config),
						"Interface": Equal("AggregateRoot"),
						"Method":    Equal("ApplyEvent"),
						"Message":   Equal(MessageE1),
					},
				),
			))
		})

		It("provides more context to UnexpectedMessage panics from ApplyEvent() when called with historical events", func() {
			handler.HandleCommandFunc = func(
				s dogma.AggregateCommandScope,
				_ dogma.Message,
			) {
				s.RecordEvent(MessageE1)
			}

			ctrl.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				command,
			)

			handler.HandleCommandFunc = nil
			handler.NewFunc = func() dogma.AggregateRoot {
				return &AggregateRoot{
					ApplyEventFunc: func(dogma.Message, interface{}) {
						panic(dogma.UnexpectedMessage)
					},
				}
			}

			Expect(func() {
				ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					command,
				)
			}).To(PanicWith(
				MatchFields(
					IgnoreExtras,
					Fields{
						"Handler":   Equal(config),
						"Interface": Equal("AggregateRoot"),
						"Method":    Equal("ApplyEvent"),
						"Message":   Equal(MessageE1),
					},
				),
			))
		})
	})

	Describe("func Reset()", func() {
		BeforeEach(func() {
			handler.HandleCommandFunc = func(
				s dogma.AggregateCommandScope,
				m dogma.Message,
			) {
				s.RecordEvent(MessageE1) // record event to create the instance
			}

			_, err := ctrl.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				command,
			)

			Expect(err).ShouldNot(HaveOccurred())
		})

		It("removes all instances", func() {
			ctrl.Reset()

			buf := &fact.Buffer{}
			_, err := ctrl.Handle(
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
