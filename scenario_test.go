package testkit_test

import (
	"context"
	"time"

	"github.com/dogmatiq/configkit"
	. "github.com/dogmatiq/configkit/fixtures"
	"github.com/dogmatiq/configkit/message"
	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogma/fixtures"
	. "github.com/dogmatiq/testkit"
	"github.com/dogmatiq/testkit/engine"
	"github.com/dogmatiq/testkit/envelope"
	"github.com/dogmatiq/testkit/fact"
	"github.com/dogmatiq/testkit/internal/testingmock"
	g "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = g.Describe("func Scenario()", func() {
	var (
		app       *Application
		t         *testingmock.T
		startTime time.Time
		buf       *fact.Buffer
		test      *Test
	)

	g.BeforeEach(func() {
		app = &Application{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "a84b2620-4675-4024-b55b-cd5dbeb6e293")
				c.RegisterAggregate(&AggregateMessageHandler{
					ConfigureFunc: func(c dogma.AggregateConfigurer) {
						c.Identity("<aggregate>", "d1cf3af1-6c20-4125-8e68-192a6075d0b4")
						c.Routes(
							dogma.HandlesCommand[MessageC](),
							dogma.RecordsEvent[MessageE](),
						)
					},
					RouteCommandToInstanceFunc: func(
						dogma.Message,
					) string {
						return "<instance>"
					},
				})
				c.RegisterProcess(&ProcessMessageHandler{
					ConfigureFunc: func(c dogma.ProcessConfigurer) {
						c.Identity("<process>", "1c0dd111-fe12-4dee-a8bc-64abea1dce8f")
						c.Routes(
							dogma.HandlesEvent[MessageE](),
							dogma.ExecutesCommand[MessageC](),
						)
					},
					RouteEventToInstanceFunc: func(
						context.Context,
						dogma.Message,
					) (string, bool, error) {
						return "<instance>", true, nil
					},
				})
			},
		}

		t = &testingmock.T{}
		startTime = time.Now()
		buf = &fact.Buffer{}

		test = Begin(
			t,
			app,
			StartTimeAt(startTime),
			WithUnsafeOperationOptions(
				engine.WithObserver(buf),
			),
		)
	})

	g.It("executes command message", func() {
		test.Given(
			Scenario("<scenario>").ExecuteCommand(MessageC1),
		)

		Expect(buf.Facts()).To(ContainElement(
			fact.DispatchCycleBegun{
				Envelope: &envelope.Envelope{
					MessageID:     "1",
					CausationID:   "1",
					CorrelationID: "1",
					Message:       MessageC1,
					Type:          MessageCType,
					Role:          message.CommandRole,
					CreatedAt:     startTime,
				},
				EngineTime: startTime,
				EnabledHandlerTypes: map[configkit.HandlerType]bool{
					configkit.AggregateHandlerType:   true,
					configkit.IntegrationHandlerType: false,
					configkit.ProcessHandlerType:     true,
					configkit.ProjectionHandlerType:  false,
				},
				EnabledHandlers: map[string]bool{},
			},
		))
	})

	g.It("records the event message", func() {
		test.Given(
			Scenario("<scenario>").RecordEvent(MessageE1),
		)

		Expect(buf.Facts()).To(ContainElement(
			fact.DispatchCycleBegun{
				Envelope: &envelope.Envelope{
					MessageID:     "1",
					CausationID:   "1",
					CorrelationID: "1",
					Message:       MessageE1,
					Type:          MessageEType,
					Role:          message.EventRole,
					CreatedAt:     startTime,
				},
				EngineTime: startTime,
				EnabledHandlerTypes: map[configkit.HandlerType]bool{
					configkit.AggregateHandlerType:   true,
					configkit.IntegrationHandlerType: false,
					configkit.ProcessHandlerType:     true,
					configkit.ProjectionHandlerType:  false,
				},
				EnabledHandlers: map[string]bool{},
			},
		))
	})

	g.It("advances the time", func() {
		test.Given(
			Scenario("<scenario>").AdvanceTime(ByDuration(1 * time.Second)),
		)

		Expect(buf.Facts()).To(ContainElement(
			fact.TickCycleBegun{
				EngineTime: startTime.Add(1 * time.Second),
				EnabledHandlerTypes: map[configkit.HandlerType]bool{
					configkit.AggregateHandlerType:   true,
					configkit.IntegrationHandlerType: false,
					configkit.ProcessHandlerType:     true,
					configkit.ProjectionHandlerType:  false,
				},
				EnabledHandlers: map[string]bool{},
			},
		))
	})

	g.It("executes the command via the call", func() {
		e := test.CommandExecutor()

		test.Given(
			Scenario("<scenario>").Call(func() {
				e.ExecuteCommand(
					context.Background(),
					MessageC1,
				)
			}),
		)

		Expect(buf.Facts()).To(ContainElement(
			fact.DispatchCycleBegun{
				Envelope: &envelope.Envelope{
					MessageID:     "1",
					CausationID:   "1",
					CorrelationID: "1",
					Message:       MessageC1,
					Type:          MessageCType,
					Role:          message.CommandRole,
					CreatedAt:     startTime,
				},
				EngineTime: startTime,
				EnabledHandlerTypes: map[configkit.HandlerType]bool{
					configkit.AggregateHandlerType:   true,
					configkit.IntegrationHandlerType: false,
					configkit.ProcessHandlerType:     true,
					configkit.ProjectionHandlerType:  false,
				},
				EnabledHandlers: map[string]bool{},
			},
		))
	})
})
