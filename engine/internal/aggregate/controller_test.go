package aggregate_test

import (
	"context"
	"fmt"
	"time"

	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogma/fixtures"
	. "github.com/dogmatiq/testkit/engine/internal/aggregate"
	"github.com/dogmatiq/testkit/envelope"
	"github.com/dogmatiq/testkit/fact"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

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
				c.ProducesEventType(MessageE{})
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
				_ dogma.AggregateRoot,
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
				_ dogma.AggregateRoot,
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
						Handler:     config,
						HandlerType: configkit.AggregateHandlerType,
						InstanceID:  "<instance-C1>",
					},
				),
				command.NewEvent(
					"2",
					MessageE2,
					now,
					envelope.Origin{
						Handler:     config,
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
			}).To(PanicWith(
				MatchAllFields(
					Fields{
						"Handler":        Equal(config),
						"Interface":      Equal("AggregateMessageHandler"),
						"Method":         Equal("RouteCommandToInstance"),
						"Implementation": Equal(config.Handler()),
						"Message":        Equal(command.Message),
						"Description":    Equal("routed a command of type fixtures.MessageC to an empty instance ID"),
						"Location": MatchAllFields(
							Fields{
								"Func": Not(BeEmpty()),
								"File": HaveSuffix("/fixtures/aggregate.go"), // from dogmatiq/dogma module
								"Line": Not(BeZero()),
							},
						),
					},
				),
			))
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
						Handler:    config,
						InstanceID: "<instance-C1>",
						Envelope:   command,
					},
				))
			})

			It("passes a new aggregate root", func() {
				handler.HandleCommandFunc = func(
					r dogma.AggregateRoot,
					s dogma.AggregateCommandScope,
					_ dogma.Message,
				) {
					Expect(r).To(Equal(&AggregateRoot{}))
				}

				_, err := ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					command,
				)

				Expect(err).ShouldNot(HaveOccurred())
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
				}).To(PanicWith(
					MatchAllFields(
						Fields{
							"Handler":        Equal(config),
							"Interface":      Equal("AggregateMessageHandler"),
							"Method":         Equal("New"),
							"Implementation": Equal(config.Handler()),
							"Message":        Equal(command.Message),
							"Description":    Equal("returned a nil AggregateRoot"),
							"Location": MatchAllFields(
								Fields{
									"Func": Not(BeEmpty()),
									"File": HaveSuffix("/fixtures/aggregate.go"), // from dogmatiq/dogma module
									"Line": Not(BeZero()),
								},
							),
						},
					),
				))
			})
		})

		When("the instance exists", func() {
			BeforeEach(func() {
				handler.HandleCommandFunc = func(
					_ dogma.AggregateRoot,
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

				handler.HandleCommandFunc = nil
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
						Handler:    config,
						InstanceID: "<instance-C1>",
						Root: &AggregateRoot{
							AppliedEvents: []dogma.Message{
								MessageE1,
							},
						},
						Envelope: command,
					},
				))
			})

			It("passes an aggregate root with historical events applied", func() {
				handler.HandleCommandFunc = func(
					r dogma.AggregateRoot,
					s dogma.AggregateCommandScope,
					_ dogma.Message,
				) {
					Expect(r).To(Equal(
						&AggregateRoot{
							AppliedEvents: []dogma.Message{
								MessageE1,
							},
						},
					))
				}

				_, err := ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					command,
				)

				Expect(err).ShouldNot(HaveOccurred())
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
				dogma.AggregateRoot,
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
				_ dogma.AggregateRoot,
				s dogma.AggregateCommandScope,
				_ dogma.Message,
			) {
				s.RecordEvent(MessageE1)
			}

			handler.NewFunc = func() dogma.AggregateRoot {
				return &AggregateRoot{
					ApplyEventFunc: func(dogma.Message) {
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
				_ dogma.AggregateRoot,
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
					ApplyEventFunc: func(dogma.Message) {
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
				_ dogma.AggregateRoot,
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
