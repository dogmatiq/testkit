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
	. "github.com/onsi/gomega/gstruct"
)

var _ = g.Describe("func Call()", func() {
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
				c.Identity("<app>", "b51cde16-aaec-4d75-ae14-06282e3a72c8")
				c.RegisterAggregate(&AggregateMessageHandler{
					ConfigureFunc: func(c dogma.AggregateConfigurer) {
						c.Identity("<aggregate>", "832d78d7-a006-414f-b6d7-3153aa7c9ab8")
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
						c.Identity("<process>", "b64cdd19-782e-4e4e-9e5f-a95a943d6340")
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

		test = Begin(
			t,
			app,
			StartTimeAt(startTime),
			WithUnsafeOperationOptions(
				engine.WithObserver(buf),
			),
		)
	})

	g.It("allows use of the test's executor", func() {
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

	g.It("allows use of the test's recorder", func() {
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

	g.It("allows expectations to match commands executed via the test's executor", func() {
		e := test.CommandExecutor()

		test.Expect(
			Call(func() {
				e.ExecuteCommand(
					context.Background(),
					MessageC1,
				)
			}),
			ToExecuteCommand(MessageC1),
		)
	})

	g.It("allows expectations to match events recorded via the test's recorder", func() {
		r := test.EventRecorder()

		test.Expect(
			Call(func() {
				r.RecordEvent(
					context.Background(),
					MessageE1,
				)
			}),
			ToRecordEvent(MessageE1),
		)
	})

	g.It("produces the expected caption", func() {
		test.Prepare(
			Call(func() {}),
		)

		Expect(t.Logs).To(ContainElement(
			"--- calling user-defined function ---",
		))
	})

	g.It("panics if the function is nil", func() {
		Expect(func() {
			Call(nil)
		}).To(PanicWith("Call(<nil>): function must not be nil"))
	})

	g.It("captures the location that the action was created", func() {
		act := call(func() {})
		Expect(act.Location()).To(MatchAllFields(
			Fields{
				"Func": Equal("github.com/dogmatiq/testkit_test.call"),
				"File": HaveSuffix("/action.linenumber_test.go"),
				"Line": Equal(51),
			},
		))
	})
})
