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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("func RecordEvent()", func() {
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
				c.Identity("<app>", "38408e83-e8eb-4f82-abe1-7fa02cee0657")
				c.RegisterProcess(&ProcessMessageHandler{
					ConfigureFunc: func(c dogma.ProcessConfigurer) {
						c.Identity("<process>", "1c0dd111-fe12-4dee-a8bc-64abea1dce8f")
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

	It("dispatches the message", func() {
		test.Prepare(
			RecordEvent(MessageE1),
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

	It("fails the test if the message type is unrecognized", func() {
		t.FailSilently = true

		test.Prepare(
			RecordEvent(MessageX1),
		)

		Expect(t.Failed()).To(BeTrue())
		Expect(t.Logs).To(ContainElement(
			"can not record event, fixtures.MessageX is a not a recognized message type",
		))
	})

	It("fails the test if the message type is not an event", func() {
		t.FailSilently = true

		test.Prepare(
			RecordEvent(MessageC1),
		)

		Expect(t.Failed()).To(BeTrue())
		Expect(t.Logs).To(ContainElement(
			"can not record event, fixtures.MessageC is configured as a command",
		))
	})

	It("does not satisfy its own expectations", func() {
		t.FailSilently = true

		test.Expect(
			RecordEvent(MessageE1),
			ToRecordEvent(MessageE1),
		)

		Expect(t.Failed()).To(BeTrue())
	})

	It("produces the expected caption", func() {
		test.Prepare(
			RecordEvent(MessageE1),
		)

		Expect(t.Logs).To(ContainElement(
			"--- recording fixtures.MessageE event ---",
		))
	})

	It("panics if the message is nil", func() {
		Expect(func() {
			RecordEvent(nil)
		}).To(PanicWith("RecordEvent(<nil>): message must not be nil"))
	})

	It("captures the location that the action was created", func() {
		act := recordEvent(MessageE1)
		Expect(act.Location()).To(MatchAllFields(
			Fields{
				"Func": Equal("github.com/dogmatiq/testkit_test.recordEvent"),
				"File": HaveSuffix("/action.linenumber_test.go"),
				"Line": Equal(53),
			},
		))
	})
})
