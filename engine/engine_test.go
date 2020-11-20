package engine_test

import (
	"context"
	"errors"
	"time"

	"github.com/dogmatiq/configkit"
	. "github.com/dogmatiq/configkit/fixtures"
	"github.com/dogmatiq/configkit/message"
	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogma/fixtures"
	. "github.com/dogmatiq/testkit/engine"
	"github.com/dogmatiq/testkit/engine/envelope"
	"github.com/dogmatiq/testkit/engine/fact"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("type Engine", func() {
	var (
		aggregate   *AggregateMessageHandler
		process     *ProcessMessageHandler
		integration *IntegrationMessageHandler
		projection  *ProjectionMessageHandler
		app         *Application
		config      configkit.RichApplication
		engine      *Engine
	)

	BeforeEach(func() {
		aggregate = &AggregateMessageHandler{
			ConfigureFunc: func(c dogma.AggregateConfigurer) {
				c.Identity("<aggregate>", "<aggregate-key>")
				c.ConsumesCommandType(MessageA{})
				c.ProducesEventType(MessageE{})
			},
			RouteCommandToInstanceFunc: func(dogma.Message) string {
				return "<instance>"
			},
		}

		process = &ProcessMessageHandler{
			ConfigureFunc: func(c dogma.ProcessConfigurer) {
				c.Identity("<process>", "<process-key>")
				c.ConsumesEventType(MessageB{})
				c.ConsumesEventType(MessageE{}) // shared with <projection>
				c.ProducesCommandType(MessageC{})
			},
			RouteEventToInstanceFunc: func(context.Context, dogma.Message) (string, bool, error) {
				return "<instance>", true, nil
			},
		}

		integration = &IntegrationMessageHandler{
			ConfigureFunc: func(c dogma.IntegrationConfigurer) {
				c.Identity("<integration>", "<integration-key>")
				c.ConsumesCommandType(MessageC{})
				c.ProducesEventType(MessageF{})
			},
		}

		projection = &ProjectionMessageHandler{
			ConfigureFunc: func(c dogma.ProjectionConfigurer) {
				c.Identity("<projection>", "<projection-key>")
				c.ConsumesEventType(MessageD{})
				c.ConsumesEventType(MessageE{}) // shared with <process>
			},
		}

		app = &Application{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "<app-key>")
				c.RegisterAggregate(aggregate)
				c.RegisterProcess(process)
				c.RegisterIntegration(integration)
				c.RegisterProjection(projection)
			},
		}

		config = configkit.FromApplication(app)
		engine = MustNew(config)
	})

	Describe("func Dispatch()", func() {
		It("skips handlers that are disabled by type", func() {
			aggregate.HandleCommandFunc = func(
				dogma.AggregateRoot,
				dogma.AggregateCommandScope,
				dogma.Message,
			) {
				Fail("unexpected call")
			}

			now := time.Now()
			buf := &fact.Buffer{}
			err := engine.Dispatch(
				context.Background(),
				MessageA1,
				WithCurrentTime(now),
				WithObserver(buf),
				EnableAggregates(false),
			)
			Expect(err).ShouldNot(HaveOccurred())

			h, _ := config.RichHandlers().ByName("<aggregate>")
			Expect(buf.Facts()).To(ContainElement(
				fact.HandlingSkipped{
					Handler: h,
					Envelope: &envelope.Envelope{
						MessageID:     "1",
						CausationID:   "1",
						CorrelationID: "1",
						Message:       MessageA1,
						Type:          MessageAType,
						Role:          message.CommandRole,
						CreatedAt:     now,
					},
					Explicit: false,
				},
			))
		})

		It("skips handlers that are disabled by name", func() {
			aggregate.HandleCommandFunc = func(
				dogma.AggregateRoot,
				dogma.AggregateCommandScope,
				dogma.Message,
			) {
				Fail("unexpected call")
			}

			now := time.Now()
			buf := &fact.Buffer{}
			err := engine.Dispatch(
				context.Background(),
				MessageA1,
				WithCurrentTime(now),
				WithObserver(buf),
				EnableHandler("<aggregate>", false),
			)
			Expect(err).ShouldNot(HaveOccurred())

			h, _ := config.RichHandlers().ByName("<aggregate>")
			Expect(buf.Facts()).To(ContainElement(
				fact.HandlingSkipped{
					Handler: h,
					Envelope: &envelope.Envelope{
						MessageID:     "1",
						CausationID:   "1",
						CorrelationID: "1",
						Message:       MessageA1,
						Type:          MessageAType,
						Role:          message.CommandRole,
						CreatedAt:     now,
					},
					Explicit: true,
				},
			))
		})

		It("does not skip handlers that are enabled by name", func() {
			called := false
			aggregate.HandleCommandFunc = func(
				dogma.AggregateRoot,
				dogma.AggregateCommandScope,
				dogma.Message,
			) {
				called = true
			}

			err := engine.Dispatch(
				context.Background(),
				MessageA1,
				EnableHandler("<aggregate>", false),
				EnableHandler("<aggregate>", true),
			)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(called).To(BeTrue())
		})

		It("panics if the message is invalid", func() {
			Expect(func() {
				engine.Dispatch(
					context.Background(),
					MessageA{
						Value: errors.New("<invalid>"),
					},
				)
			}).To(PanicWith("can not dispatch invalid fixtures.MessageA message: <invalid>"))
		})

		It("panics if the message type is unrecognized", func() {
			Expect(func() {
				engine.Dispatch(context.Background(), MessageX1)
			}).To(PanicWith("the fixtures.MessageX message type is not consumed by any handlers"))
		})
	})

	Describe("func Tick()", func() {
		It("skips handlers that are disabled by type", func() {
			buf := &fact.Buffer{}
			err := engine.Tick(
				context.Background(),
				WithObserver(buf),
				EnableAggregates(false),
			)
			Expect(err).ShouldNot(HaveOccurred())

			h, _ := config.RichHandlers().ByName("<aggregate>")
			Expect(buf.Facts()).To(ContainElement(
				fact.TickSkipped{
					Handler:  h,
					Explicit: false,
				},
			))
		})

		It("skips handlers that are disabled by name", func() {
			buf := &fact.Buffer{}
			err := engine.Tick(
				context.Background(),
				WithObserver(buf),
				EnableHandler("<aggregate>", false),
			)
			Expect(err).ShouldNot(HaveOccurred())

			h, _ := config.RichHandlers().ByName("<aggregate>")
			Expect(buf.Facts()).To(ContainElement(
				fact.TickSkipped{
					Handler:  h,
					Explicit: true,
				},
			))
		})

		It("does not skip handlers that are enabled by name", func() {
			buf := &fact.Buffer{}
			err := engine.Tick(
				context.Background(),
				WithObserver(buf),
				EnableHandler("<aggregate>", false),
				EnableHandler("<aggregate>", true),
			)
			Expect(err).ShouldNot(HaveOccurred())

			h, _ := config.RichHandlers().ByName("<aggregate>")
			Expect(buf.Facts()).To(ContainElement(
				fact.TickBegun{
					Handler: h,
				},
			))
		})
	})
})
