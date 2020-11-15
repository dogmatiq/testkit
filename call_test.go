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
	"github.com/dogmatiq/testkit/assert"
	"github.com/dogmatiq/testkit/engine"
	"github.com/dogmatiq/testkit/engine/envelope"
	"github.com/dogmatiq/testkit/engine/fact"
	"github.com/dogmatiq/testkit/internal/testingmock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("func Call()", func() {
	var (
		app       *Application
		t         *testingmock.T
		startTime time.Time
		buf       *fact.Buffer
		test      *Test
	)

	BeforeEach(func() {
		app = &Application{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "<app-key>")
				c.RegisterAggregate(&AggregateMessageHandler{
					ConfigureFunc: func(c dogma.AggregateConfigurer) {
						c.Identity("<aggregate>", "<aggregate-key>")
						c.ConsumesCommandType(MessageC{})
						c.ProducesEventType(MessageE{})
					},
					RouteCommandToInstanceFunc: func(
						dogma.Message,
					) string {
						return "<instance>"
					},
				})
				c.RegisterProcess(&ProcessMessageHandler{
					ConfigureFunc: func(c dogma.ProcessConfigurer) {
						c.Identity("<process>", "<process-key>")
						c.ConsumesEventType(MessageE{})
						c.ProducesCommandType(MessageC{})
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

		test = New(app).Begin(
			t,
			WithStartTime(startTime),
			WithOperationOptions(
				engine.WithObserver(buf),
			),
		)
	})

	It("allows use of the test's executor", func() {
		e := test.CommandExecutor()

		test.Prepare(
			Call(func() {
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
				EnabledHandlers: map[configkit.HandlerType]bool{
					configkit.AggregateHandlerType:   true,
					configkit.IntegrationHandlerType: false,
					configkit.ProcessHandlerType:     true,
					configkit.ProjectionHandlerType:  false,
				},
			},
		))
	})

	It("allows use of the test's recorder", func() {
		r := test.EventRecorder()

		test.Prepare(
			Call(func() {
				r.RecordEvent(
					context.Background(),
					MessageE1,
				)
			}),
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
				EnabledHandlers: map[configkit.HandlerType]bool{
					configkit.AggregateHandlerType:   true,
					configkit.IntegrationHandlerType: false,
					configkit.ProcessHandlerType:     true,
					configkit.ProjectionHandlerType:  false,
				},
			},
		))
	})

	It("allows expectations to match commands executed via the test's executor", func() {
		e := test.CommandExecutor()

		test.Expect(
			Call(func() {
				e.ExecuteCommand(
					context.Background(),
					MessageC1,
				)
			}),
			assert.CommandExecuted(MessageC1),
		)
	})

	It("allows expectations to match events recorded via the test's recorder", func() {
		r := test.EventRecorder()

		test.Expect(
			Call(func() {
				r.RecordEvent(
					context.Background(),
					MessageE1,
				)
			}),
			assert.EventRecorded(MessageE1),
		)
	})

	It("logs a suitable heading", func() {
		test.Prepare(
			Call(func() {}),
		)

		Expect(t.Logs).To(ContainElement(
			"--- PREPARE: CALLING USER-DEFINED FUNCTION ---",
		))
	})

	It("panics if the function is nil", func() {
		Expect(func() {
			Call(nil)
		}).To(PanicWith("Call(): function must not be nil"))
	})
})
