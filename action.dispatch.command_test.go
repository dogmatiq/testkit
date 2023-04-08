package testkit_test

import (
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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("func ExecuteCommand()", func() {
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
				c.Identity("<app>", "a84b2620-4675-4024-b55b-cd5dbeb6e293")
				c.RegisterAggregate(&AggregateMessageHandler{
					ConfigureFunc: func(c dogma.AggregateConfigurer) {
						c.Identity("<aggregate>", "d1cf3af1-6c20-4125-8e68-192a6075d0b4")
						c.ConsumesCommandType(MessageC{})
						c.ProducesEventType(MessageE{})
					},
					RouteCommandToInstanceFunc: func(
						dogma.Message,
					) string {
						return "<instance>"
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

	It("dispatches the message", func() {
		test.Prepare(
			ExecuteCommand(MessageC1),
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

	It("fails the test if the message type is unrecognized", func() {
		t.FailSilently = true

		test.Prepare(
			ExecuteCommand(MessageX1),
		)

		Expect(t.Failed()).To(BeTrue())
		Expect(t.Logs).To(ContainElement(
			"can not execute command, fixtures.MessageX is a not a recognized message type",
		))
	})

	It("fails the test if the message type is not a command", func() {
		t.FailSilently = true

		test.Prepare(
			ExecuteCommand(MessageE1),
		)

		Expect(t.Failed()).To(BeTrue())
		Expect(t.Logs).To(ContainElement(
			"can not execute command, fixtures.MessageE is configured as an event",
		))
	})

	It("does not satisfy its own expectations", func() {
		t.FailSilently = true

		test.Expect(
			ExecuteCommand(MessageC1),
			ToExecuteCommand(MessageC1),
		)

		Expect(t.Failed()).To(BeTrue())
	})

	It("produces the expected caption", func() {
		test.Prepare(
			ExecuteCommand(MessageC1),
		)

		Expect(t.Logs).To(ContainElement(
			"--- executing fixtures.MessageC command ---",
		))
	})

	It("panics if the message is nil", func() {
		Expect(func() {
			ExecuteCommand(nil)
		}).To(PanicWith("ExecuteCommand(<nil>): message must not be nil"))
	})

	It("captures the location that the action was created", func() {
		act := executeCommand(MessageC1)
		Expect(act.Location()).To(MatchAllFields(
			Fields{
				"Func": Equal("github.com/dogmatiq/testkit_test.executeCommand"),
				"File": HaveSuffix("/action.linenumber_test.go"),
				"Line": Equal(52),
			},
		))
	})
})
