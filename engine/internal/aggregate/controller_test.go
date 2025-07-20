package aggregate_test

import (
	"context"
	"fmt"
	"time"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/enginekit/config"
	"github.com/dogmatiq/enginekit/config/runtimeconfig"
	. "github.com/dogmatiq/enginekit/enginetest/stubs"
	. "github.com/dogmatiq/testkit/engine/internal/aggregate"
	"github.com/dogmatiq/testkit/envelope"
	"github.com/dogmatiq/testkit/fact"
	g "github.com/onsi/ginkgo/v2"
	gm "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = g.Describe("type Controller", func() {
	var (
		messageIDs envelope.MessageIDGenerator
		handler    *AggregateMessageHandlerStub
		cfg        *config.Aggregate
		ctrl       *Controller
		command    *envelope.Envelope
	)

	g.BeforeEach(func() {
		command = envelope.NewCommand(
			"1000",
			CommandA1,
			time.Now(),
		)

		handler = &AggregateMessageHandlerStub{
			ConfigureFunc: func(c dogma.AggregateConfigurer) {
				c.Identity("<name>", "e8fd6bd4-c3a3-4eb4-bf0f-56862a123229")
				c.Routes(
					dogma.HandlesCommand[CommandStub[TypeA]](),
					dogma.RecordsEvent[EventStub[TypeA]](),
				)
			},
			RouteCommandToInstanceFunc: func(m dogma.Command) string {
				switch x := m.(type) {
				case CommandStub[TypeA]:
					return fmt.Sprintf(
						"<instance-%s>",
						x.Content,
					)
				default:
					panic(dogma.UnexpectedMessage)
				}
			},
		}

		cfg = runtimeconfig.FromAggregate(handler)

		ctrl = &Controller{
			Config:     cfg,
			MessageIDs: &messageIDs,
		}

		messageIDs.Reset() // reset after setup for a predictable ID.
	})

	g.Describe("func HandlerConfig()", func() {
		g.It("returns the handler config", func() {
			gm.Expect(ctrl.HandlerConfig()).To(gm.BeIdenticalTo(cfg))
		})
	})

	g.Describe("func Tick()", func() {
		g.It("does not return any envelopes", func() {
			envelopes, err := ctrl.Tick(
				context.Background(),
				fact.Ignore,
				time.Now(),
			)
			gm.Expect(err).ShouldNot(gm.HaveOccurred())
			gm.Expect(envelopes).To(gm.BeEmpty())
		})

		g.It("does not record any facts", func() {
			buf := &fact.Buffer{}
			_, err := ctrl.Tick(
				context.Background(),
				buf,
				time.Now(),
			)
			gm.Expect(err).ShouldNot(gm.HaveOccurred())
			gm.Expect(buf.Facts()).To(gm.BeEmpty())
		})
	})

	g.Describe("func Handle()", func() {
		g.It("forwards the message to the handler", func() {
			called := false
			handler.HandleCommandFunc = func(
				_ dogma.AggregateRoot,
				_ dogma.AggregateCommandScope,
				m dogma.Command,
			) {
				called = true
				gm.Expect(m).To(gm.Equal(CommandA1))
			}

			_, err := ctrl.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				command,
			)

			gm.Expect(err).ShouldNot(gm.HaveOccurred())
			gm.Expect(called).To(gm.BeTrue())
		})

		g.It("returns the recorded events", func() {
			handler.HandleCommandFunc = func(
				_ dogma.AggregateRoot,
				s dogma.AggregateCommandScope,
				_ dogma.Command,
			) {
				s.RecordEvent(EventA1)
				s.RecordEvent(EventA2)
			}

			now := time.Now()
			events, err := ctrl.Handle(
				context.Background(),
				fact.Ignore,
				now,
				command,
			)

			gm.Expect(err).ShouldNot(gm.HaveOccurred())
			gm.Expect(events).To(gm.ConsistOf(
				command.NewEvent(
					"1",
					EventA1,
					now,
					envelope.Origin{
						Handler:     cfg,
						HandlerType: config.AggregateHandlerType,
						InstanceID:  "<instance-A1>",
					},
					"78e27a08-0ae8-52cf-8f46-79e448ed5bf6",
					0,
				),
				command.NewEvent(
					"2",
					EventA2,
					now,
					envelope.Origin{
						Handler:     cfg,
						HandlerType: config.AggregateHandlerType,
						InstanceID:  "<instance-A1>",
					},
					"78e27a08-0ae8-52cf-8f46-79e448ed5bf6",
					1,
				),
			))
		})

		g.It("panics when the handler routes to an empty instance ID", func() {
			handler.RouteCommandToInstanceFunc = func(dogma.Command) string {
				return ""
			}

			gm.Expect(func() {
				ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					command,
				)
			}).To(gm.PanicWith(
				MatchAllFields(
					Fields{
						"Handler":        gm.Equal(cfg),
						"Interface":      gm.Equal("AggregateMessageHandler"),
						"Method":         gm.Equal("RouteCommandToInstance"),
						"Implementation": gm.Equal(cfg.Source.Get()),
						"Message":        gm.Equal(command.Message),
						"Description":    gm.Equal("routed a command of type stubs.CommandStub[TypeA] to an empty ID"),
						"Location": MatchAllFields(
							Fields{
								"Func": gm.Not(gm.BeEmpty()),
								"File": gm.HaveSuffix("/stubs/aggregate.go"), // from dogmatiq/enginekit module
								"Line": gm.Not(gm.BeZero()),
							},
						),
					},
				),
			))
		})

		g.When("the instance does not exist", func() {
			g.It("records a fact", func() {
				buf := &fact.Buffer{}
				_, err := ctrl.Handle(
					context.Background(),
					buf,
					time.Now(),
					command,
				)

				gm.Expect(err).ShouldNot(gm.HaveOccurred())
				gm.Expect(buf.Facts()).To(gm.ContainElement(
					fact.AggregateInstanceNotFound{
						Handler:    cfg,
						InstanceID: "<instance-A1>",
						Envelope:   command,
					},
				))
			})

			g.It("passes a new aggregate root", func() {
				handler.HandleCommandFunc = func(
					r dogma.AggregateRoot,
					s dogma.AggregateCommandScope,
					_ dogma.Command,
				) {
					gm.Expect(r).To(gm.Equal(&AggregateRootStub{}))
				}

				_, err := ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					command,
				)

				gm.Expect(err).ShouldNot(gm.HaveOccurred())
			})

			g.It("panics if New() returns nil", func() {
				handler.NewFunc = func() dogma.AggregateRoot {
					return nil
				}

				gm.Expect(func() {
					ctrl.Handle(
						context.Background(),
						fact.Ignore,
						time.Now(),
						command,
					)
				}).To(gm.PanicWith(
					MatchAllFields(
						Fields{
							"Handler":        gm.Equal(cfg),
							"Interface":      gm.Equal("AggregateMessageHandler"),
							"Method":         gm.Equal("New"),
							"Implementation": gm.Equal(cfg.Source.Get()),
							"Message":        gm.Equal(command.Message),
							"Description":    gm.Equal("returned a nil AggregateRoot"),
							"Location": MatchAllFields(
								Fields{
									"Func": gm.Not(gm.BeEmpty()),
									"File": gm.HaveSuffix("/stubs/aggregate.go"), // from dogmatiq/enginekit module
									"Line": gm.Not(gm.BeZero()),
								},
							),
						},
					),
				))
			})
		})

		g.When("the instance exists", func() {
			g.BeforeEach(func() {
				handler.HandleCommandFunc = func(
					_ dogma.AggregateRoot,
					s dogma.AggregateCommandScope,
					_ dogma.Command,
				) {
					s.RecordEvent(EventA1) // record event to create the instance
				}

				_, err := ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					command,
				)

				gm.Expect(err).ShouldNot(gm.HaveOccurred())

				handler.HandleCommandFunc = nil
			})

			g.It("records a fact", func() {
				buf := &fact.Buffer{}
				_, err := ctrl.Handle(
					context.Background(),
					buf,
					time.Now(),
					command,
				)

				gm.Expect(err).ShouldNot(gm.HaveOccurred())
				gm.Expect(buf.Facts()).To(gm.ContainElement(
					fact.AggregateInstanceLoaded{
						Handler:    cfg,
						InstanceID: "<instance-A1>",
						Root: &AggregateRootStub{
							AppliedEvents: []dogma.Event{
								EventA1,
							},
						},
						Envelope: command,
					},
				))
			})

			g.It("passes an aggregate root with historical events applied", func() {
				handler.HandleCommandFunc = func(
					r dogma.AggregateRoot,
					s dogma.AggregateCommandScope,
					_ dogma.Command,
				) {
					gm.Expect(r).To(gm.Equal(
						&AggregateRootStub{
							AppliedEvents: []dogma.Event{
								EventA1,
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

				gm.Expect(err).ShouldNot(gm.HaveOccurred())
			})
		})

		g.It("provides more context to UnexpectedMessage panics from RouteCommandToInstance()", func() {
			handler.RouteCommandToInstanceFunc = func(dogma.Command) string {
				panic(dogma.UnexpectedMessage)
			}

			gm.Expect(func() {
				ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					command,
				)
			}).To(gm.PanicWith(
				MatchFields(
					IgnoreExtras,
					Fields{
						"Handler":   gm.Equal(cfg),
						"Interface": gm.Equal("AggregateMessageHandler"),
						"Method":    gm.Equal("RouteCommandToInstance"),
						"Message":   gm.Equal(command.Message),
					},
				),
			))
		})

		g.It("provides more context to UnexpectedMessage panics from HandleCommand()", func() {
			handler.HandleCommandFunc = func(
				dogma.AggregateRoot,
				dogma.AggregateCommandScope,
				dogma.Command,
			) {
				panic(dogma.UnexpectedMessage)
			}

			gm.Expect(func() {
				ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					command,
				)
			}).To(gm.PanicWith(
				MatchFields(
					IgnoreExtras,
					Fields{
						"Handler":   gm.Equal(cfg),
						"Interface": gm.Equal("AggregateMessageHandler"),
						"Method":    gm.Equal("HandleCommand"),
						"Message":   gm.Equal(command.Message),
					},
				),
			))
		})

		g.It("provides more context to UnexpectedMessage panics from ApplyEvent() when called with new events", func() {
			handler.HandleCommandFunc = func(
				_ dogma.AggregateRoot,
				s dogma.AggregateCommandScope,
				_ dogma.Command,
			) {
				s.RecordEvent(EventA1)
			}

			handler.NewFunc = func() dogma.AggregateRoot {
				return &AggregateRootStub{
					ApplyEventFunc: func(dogma.Event) {
						panic(dogma.UnexpectedMessage)
					},
				}
			}

			gm.Expect(func() {
				ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					command,
				)
			}).To(gm.PanicWith(
				MatchFields(
					IgnoreExtras,
					Fields{
						"Handler":   gm.Equal(cfg),
						"Interface": gm.Equal("AggregateRoot"),
						"Method":    gm.Equal("ApplyEvent"),
						"Message":   gm.Equal(EventA1),
					},
				),
			))
		})

		g.It("provides more context to UnexpectedMessage panics from ApplyEvent() when called with historical events", func() {
			handler.HandleCommandFunc = func(
				_ dogma.AggregateRoot,
				s dogma.AggregateCommandScope,
				_ dogma.Command,
			) {
				s.RecordEvent(EventA1)
			}

			ctrl.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				command,
			)

			handler.HandleCommandFunc = nil
			handler.NewFunc = func() dogma.AggregateRoot {
				return &AggregateRootStub{
					ApplyEventFunc: func(dogma.Event) {
						panic(dogma.UnexpectedMessage)
					},
				}
			}

			gm.Expect(func() {
				ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					command,
				)
			}).To(gm.PanicWith(
				MatchFields(
					IgnoreExtras,
					Fields{
						"Handler":   gm.Equal(cfg),
						"Interface": gm.Equal("AggregateRoot"),
						"Method":    gm.Equal("ApplyEvent"),
						"Message":   gm.Equal(EventA1),
					},
				),
			))
		})
	})

	g.Describe("func Reset()", func() {
		g.BeforeEach(func() {
			handler.HandleCommandFunc = func(
				_ dogma.AggregateRoot,
				s dogma.AggregateCommandScope,
				_ dogma.Command,
			) {
				s.RecordEvent(EventA1) // record event to create the instance
			}

			_, err := ctrl.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				command,
			)

			gm.Expect(err).ShouldNot(gm.HaveOccurred())
		})

		g.It("removes all instances", func() {
			ctrl.Reset()

			buf := &fact.Buffer{}
			_, err := ctrl.Handle(
				context.Background(),
				buf,
				time.Now(),
				command,
			)

			gm.Expect(err).ShouldNot(gm.HaveOccurred())
			gm.Expect(buf.Facts()).NotTo(gm.ContainElement(
				gm.BeAssignableToTypeOf(fact.AggregateInstanceLoaded{}),
			))
		})
	})
})
