package aggregate_test

import (
	"context"
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

var _ = g.Describe("type scope", func() {
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
				c.Identity("<name>", "fd88e430-32fe-49a6-888f-f678dcf924ef")
				c.Routes(
					dogma.HandlesCommand[CommandStub[TypeA]](),
					dogma.RecordsEvent[EventStub[TypeA]](),
				)
			},
			RouteCommandToInstanceFunc: func(m dogma.Command) string {
				switch m.(type) {
				case CommandStub[TypeA]:
					return "<instance>"
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

	g.When("the instance does not exist", func() {
		g.Describe("func RecordEvent()", func() {
			g.It("records facts about instance creation and the event", func() {
				handler.HandleCommandFunc = func(
					_ dogma.AggregateRoot,
					s dogma.AggregateCommandScope,
					_ dogma.Command,
				) {
					s.RecordEvent(EventA1)
				}

				now := time.Now()
				buf := &fact.Buffer{}
				_, err := ctrl.Handle(
					context.Background(),
					buf,
					now,
					command,
				)

				gm.Expect(err).ShouldNot(gm.HaveOccurred())
				gm.Expect(buf.Facts()).To(gm.ContainElement(
					fact.AggregateInstanceCreated{
						Handler:    cfg,
						InstanceID: "<instance>",
						Root: &AggregateRootStub{
							AppliedEvents: []dogma.Event{
								EventA1,
							},
						},
						Envelope: command,
					},
				))
				gm.Expect(buf.Facts()).To(gm.ContainElement(
					fact.EventRecordedByAggregate{
						Handler:    cfg,
						InstanceID: "<instance>",
						Root: &AggregateRootStub{
							AppliedEvents: []dogma.Event{
								EventA1,
							},
						},
						Envelope: command,
						EventEnvelope: command.NewEvent(
							"1",
							EventA1,
							now,
							envelope.Origin{
								Handler:     cfg,
								HandlerType: config.AggregateHandlerType,
								InstanceID:  "<instance>",
							},
							"aa9aa868-af3f-5dbb-a718-223782f4c77c",
							0,
						),
					},
				))
			})
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
				envelope.NewCommand(
					"2000",
					CommandA2, // use a different message to create the instance
					time.Now(),
				),
			)

			gm.Expect(err).ShouldNot(gm.HaveOccurred())

			messageIDs.Reset() // reset after setup for a predictable ID.
		})

		g.Describe("func RecordEvent()", func() {
			g.BeforeEach(func() {
				handler.HandleCommandFunc = func(
					_ dogma.AggregateRoot,
					s dogma.AggregateCommandScope,
					_ dogma.Command,
				) {
					s.RecordEvent(EventA1)
				}
			})

			g.It("records a fact", func() {
				messageIDs.Reset() // reset after setup for a predictable ID.

				buf := &fact.Buffer{}
				now := time.Now()
				_, err := ctrl.Handle(
					context.Background(),
					buf,
					now,
					command,
				)

				gm.Expect(err).ShouldNot(gm.HaveOccurred())
				gm.Expect(buf.Facts()).To(gm.ContainElement(
					fact.EventRecordedByAggregate{
						Handler:    cfg,
						InstanceID: "<instance>",
						Root: &AggregateRootStub{
							AppliedEvents: []dogma.Event{
								EventA1,
								EventA1,
							},
						},
						Envelope: command,
						EventEnvelope: command.NewEvent(
							"1",
							EventA1,
							now,
							envelope.Origin{
								Handler:     cfg,
								HandlerType: config.AggregateHandlerType,
								InstanceID:  "<instance>",
							},
							"aa9aa868-af3f-5dbb-a718-223782f4c77c",
							1,
						),
					},
				))
			})

			g.It("does not record a fact about instance creation", func() {
				buf := &fact.Buffer{}
				_, err := ctrl.Handle(
					context.Background(),
					buf,
					time.Now(),
					command,
				)

				gm.Expect(err).ShouldNot(gm.HaveOccurred())
				gm.Expect(buf.Facts()).NotTo(gm.ContainElement(
					gm.BeAssignableToTypeOf(fact.AggregateInstanceCreated{}),
				))
			})

			g.It("panics if the event type is not configured to be produced", func() {
				handler.HandleCommandFunc = func(
					_ dogma.AggregateRoot,
					s dogma.AggregateCommandScope,
					_ dogma.Command,
				) {
					s.RecordEvent(EventX1)
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
							"Method":         gm.Equal("HandleCommand"),
							"Implementation": gm.Equal(cfg.Source.Get()),
							"Message":        gm.Equal(command.Message),
							"Description":    gm.Equal("recorded an event of type stubs.EventStub[TypeX], which is not produced by this handler"),
							"Location": MatchAllFields(
								Fields{
									"Func": gm.Not(gm.BeEmpty()),
									"File": gm.HaveSuffix("/engine/internal/aggregate/scope_test.go"),
									"Line": gm.Not(gm.BeZero()),
								},
							),
						},
					),
				))
			})

			g.It("panics if the event is invalid", func() {
				handler.HandleCommandFunc = func(
					_ dogma.AggregateRoot,
					s dogma.AggregateCommandScope,
					_ dogma.Command,
				) {
					s.RecordEvent(EventStub[TypeA]{
						ValidationError: "<invalid>",
					})
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
							"Method":         gm.Equal("HandleCommand"),
							"Implementation": gm.Equal(cfg.Source.Get()),
							"Message":        gm.Equal(command.Message),
							"Description":    gm.Equal("recorded an invalid stubs.EventStub[TypeA] event: <invalid>"),
							"Location": MatchAllFields(
								Fields{
									"Func": gm.Not(gm.BeEmpty()),
									"File": gm.HaveSuffix("/engine/internal/aggregate/scope_test.go"),
									"Line": gm.Not(gm.BeZero()),
								},
							),
						},
					),
				))
			})
		})
	})

	g.Describe("func InstanceID()", func() {
		g.It("returns the instance ID", func() {
			called := false
			handler.HandleCommandFunc = func(
				_ dogma.AggregateRoot,
				s dogma.AggregateCommandScope,
				_ dogma.Command,
			) {
				called = true
				gm.Expect(s.InstanceID()).To(gm.Equal("<instance>"))
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
	})

	g.Describe("func Log()", func() {
		g.BeforeEach(func() {
			handler.HandleCommandFunc = func(
				_ dogma.AggregateRoot,
				s dogma.AggregateCommandScope,
				_ dogma.Command,
			) {
				s.Log("<format>", "<arg-1>", "<arg-2>")
			}
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
				fact.MessageLoggedByAggregate{
					Handler:    cfg,
					InstanceID: "<instance>",
					Root:       &AggregateRootStub{},
					Envelope:   command,
					LogFormat:  "<format>",
					LogArguments: []any{
						"<arg-1>",
						"<arg-2>",
					},
				},
			))
		})
	})
})
