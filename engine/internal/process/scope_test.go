package process_test

import (
	"context"
	"time"

	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/enginekit/enginetest/stubs"
	. "github.com/dogmatiq/testkit/engine/internal/process"
	"github.com/dogmatiq/testkit/envelope"
	"github.com/dogmatiq/testkit/fact"
	g "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = g.Describe("type scope", func() {
	var (
		messageIDs envelope.MessageIDGenerator
		handler    *ProcessMessageHandlerStub
		config     configkit.RichProcess
		ctrl       *Controller
		event      *envelope.Envelope
	)

	g.BeforeEach(func() {
		event = envelope.NewEvent(
			"1000",
			EventA1,
			time.Now(),
		)

		handler = &ProcessMessageHandlerStub{
			ConfigureFunc: func(c dogma.ProcessConfigurer) {
				c.Identity("<name>", "6901c34c-6e4d-4184-9414-780cb21a791a")
				c.Routes(
					dogma.HandlesEvent[EventStub[TypeA]](),
					dogma.ExecutesCommand[CommandStub[TypeA]](),
					dogma.SchedulesTimeout[TimeoutStub[TypeA]](),
				)
			},
			RouteEventToInstanceFunc: func(
				_ context.Context,
				m dogma.Event,
			) (string, bool, error) {
				switch m.(type) {
				case EventStub[TypeA]:
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

	g.Describe("func InstanceID()", func() {
		g.It("returns the instance ID", func() {
			called := false
			handler.HandleEventFunc = func(
				_ context.Context,
				_ dogma.ProcessRoot,
				s dogma.ProcessEventScope,
				_ dogma.Event,
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

	g.Describe("func End()", func() {
		g.It("records a fact", func() {
			handler.HandleEventFunc = func(
				_ context.Context,
				_ dogma.ProcessRoot,
				s dogma.ProcessEventScope,
				_ dogma.Event,
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
					Root:       &ProcessRootStub{},
					Envelope:   event,
				},
			))
		})

		g.It("does nothing if the instance has already been ended", func() {
			handler.HandleEventFunc = func(
				_ context.Context,
				_ dogma.ProcessRoot,
				s dogma.ProcessEventScope,
				_ dogma.Event,
			) error {
				s.End()
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
			Expect(buf.Facts()).To(HaveLen(3)) // not found, instance begun, instance ended (once)
		})
	})

	g.Describe("func ExecuteCommand()", func() {
		g.It("records a fact", func() {
			handler.HandleEventFunc = func(
				_ context.Context,
				_ dogma.ProcessRoot,
				s dogma.ProcessEventScope,
				_ dogma.Event,
			) error {
				s.ExecuteCommand(CommandA1)
				return nil
			}

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
					Root:       &ProcessRootStub{},
					Envelope:   event,
					CommandEnvelope: event.NewCommand(
						"1",
						CommandA1,
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

		g.It("panics if the command type is not configured to be produced", func() {
			handler.HandleEventFunc = func(
				_ context.Context,
				_ dogma.ProcessRoot,
				s dogma.ProcessEventScope,
				_ dogma.Event,
			) error {
				s.ExecuteCommand(CommandX1)
				return nil
			}

			Expect(func() {
				ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					event,
				)
			}).To(PanicWith(
				MatchAllFields(
					Fields{
						"Handler":        Equal(config),
						"Interface":      Equal("ProcessMessageHandler"),
						"Method":         Equal("HandleEvent"),
						"Implementation": Equal(config.Handler()),
						"Message":        Equal(event.Message),
						"Description":    Equal("executed a command of type stubs.CommandStub[TypeX], which is not produced by this handler"),
						"Location": MatchAllFields(
							Fields{
								"Func": Not(BeEmpty()),
								"File": HaveSuffix("/engine/internal/process/scope_test.go"),
								"Line": Not(BeZero()),
							},
						),
					},
				),
			))
		})

		g.It("panics if the command is invalid", func() {
			handler.HandleEventFunc = func(
				_ context.Context,
				_ dogma.ProcessRoot,
				s dogma.ProcessEventScope,
				_ dogma.Event,
			) error {
				s.ExecuteCommand(CommandStub[TypeA]{
					ValidationError: "<invalid>",
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
			}).To(PanicWith(
				MatchAllFields(
					Fields{
						"Handler":        Equal(config),
						"Interface":      Equal("ProcessMessageHandler"),
						"Method":         Equal("HandleEvent"),
						"Implementation": Equal(config.Handler()),
						"Message":        Equal(event.Message),
						"Description":    Equal("executed an invalid stubs.CommandStub[TypeA] command: <invalid>"),
						"Location": MatchAllFields(
							Fields{
								"Func": Not(BeEmpty()),
								"File": HaveSuffix("/engine/internal/process/scope_test.go"),
								"Line": Not(BeZero()),
							},
						),
					},
				),
			))
		})

		g.It("reverts a prior call to End()", func() {
			handler.HandleEventFunc = func(
				_ context.Context,
				_ dogma.ProcessRoot,
				s dogma.ProcessEventScope,
				_ dogma.Event,
			) error {
				s.End()
				s.ExecuteCommand(CommandA1)
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
				fact.ProcessInstanceEndingReverted{
					Handler:    config,
					InstanceID: "<instance>",
					Root:       &ProcessRootStub{},
					Envelope:   event,
				},
			))
		})
	})

	g.Describe("func ScheduleTimeout()", func() {
		g.It("records a fact", func() {
			t := time.Now().Add(10 * time.Second)

			handler.HandleEventFunc = func(
				_ context.Context,
				_ dogma.ProcessRoot,
				s dogma.ProcessEventScope,
				_ dogma.Event,
			) error {
				s.ScheduleTimeout(TimeoutA1, t)
				return nil
			}

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
					Root:       &ProcessRootStub{},
					Envelope:   event,
					TimeoutEnvelope: event.NewTimeout(
						"1",
						TimeoutA1,
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

		g.It("panics if the timeout type is not configured to be scheduled", func() {
			handler.HandleEventFunc = func(
				_ context.Context,
				_ dogma.ProcessRoot,
				s dogma.ProcessEventScope,
				_ dogma.Event,
			) error {
				s.ScheduleTimeout(TimeoutX1, time.Now())
				return nil
			}

			Expect(func() {
				ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					event,
				)
			}).To(PanicWith(
				MatchAllFields(
					Fields{
						"Handler":        Equal(config),
						"Interface":      Equal("ProcessMessageHandler"),
						"Method":         Equal("HandleEvent"),
						"Implementation": Equal(config.Handler()),
						"Message":        Equal(event.Message),
						"Description":    Equal("scheduled a timeout of type stubs.TimeoutStub[TypeX], which is not produced by this handler"),
						"Location": MatchAllFields(
							Fields{
								"Func": Not(BeEmpty()),
								"File": HaveSuffix("/engine/internal/process/scope_test.go"),
								"Line": Not(BeZero()),
							},
						),
					},
				),
			))
		})

		g.It("panics if the timeout is invalid", func() {
			handler.HandleEventFunc = func(
				_ context.Context,
				_ dogma.ProcessRoot,
				s dogma.ProcessEventScope,
				m dogma.Event,
			) error {
				s.ScheduleTimeout(
					TimeoutStub[TypeA]{
						ValidationError: "<invalid>",
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
			}).To(PanicWith(
				MatchAllFields(
					Fields{
						"Handler":        Equal(config),
						"Interface":      Equal("ProcessMessageHandler"),
						"Method":         Equal("HandleEvent"),
						"Implementation": Equal(config.Handler()),
						"Message":        Equal(event.Message),
						"Description":    Equal("scheduled an invalid stubs.TimeoutStub[TypeA] timeout: <invalid>"),
						"Location": MatchAllFields(
							Fields{
								"Func": Not(BeEmpty()),
								"File": HaveSuffix("/engine/internal/process/scope_test.go"),
								"Line": Not(BeZero()),
							},
						),
					},
				),
			))
		})

		g.It("reverts a prior call to End()", func() {
			handler.HandleEventFunc = func(
				_ context.Context,
				_ dogma.ProcessRoot,
				s dogma.ProcessEventScope,
				_ dogma.Event,
			) error {
				s.End()
				s.ScheduleTimeout(TimeoutA1, time.Now())
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
				fact.ProcessInstanceEndingReverted{
					Handler:    config,
					InstanceID: "<instance>",
					Root:       &ProcessRootStub{},
					Envelope:   event,
				},
			))
		})
	})

	g.Describe("func ScheduledFor()", func() {
		g.It("returns the time that the timeout message was scheduled to occur", func() {
			_, err := ctrl.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				event, // create the instance
			)

			timeout := event.NewTimeout(
				"2000",
				TimeoutA1,
				time.Now(),
				time.Now().Add(10*time.Second),
				envelope.Origin{
					Handler:     config,
					HandlerType: configkit.ProcessHandlerType,
					InstanceID:  "<instance>",
				},
			)

			handler.HandleTimeoutFunc = func(
				_ context.Context,
				_ dogma.ProcessRoot,
				s dogma.ProcessTimeoutScope,
				_ dogma.Timeout,
			) error {
				Expect(s.ScheduledFor()).To(BeTemporally("==", timeout.ScheduledFor))
				return nil
			}

			_, err = ctrl.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				timeout,
			)

			Expect(err).ShouldNot(HaveOccurred())
		})
	})

	g.Describe("func Log()", func() {
		g.BeforeEach(func() {
			handler.HandleEventFunc = func(
				_ context.Context,
				_ dogma.ProcessRoot,
				s dogma.ProcessEventScope,
				_ dogma.Event,
			) error {
				s.Log("<format>", "<arg-1>", "<arg-2>")
				return nil
			}
		})

		g.It("records a fact", func() {
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
					Root:       &ProcessRootStub{},
					Envelope:   event,
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
