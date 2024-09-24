package testkit_test

import (
	"context"
	"time"

	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/configkit/message"
	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/enginekit/enginetest/stubs"
	. "github.com/dogmatiq/testkit"
	"github.com/dogmatiq/testkit/engine"
	"github.com/dogmatiq/testkit/envelope"
	"github.com/dogmatiq/testkit/fact"
	"github.com/dogmatiq/testkit/internal/testingmock"
	g "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = g.Describe("func RecordEvent()", func() {
	var (
		app       *ApplicationStub
		t         *testingmock.T
		startTime time.Time
		buf       *fact.Buffer
		test      *Test
	)

	g.BeforeEach(func() {
		app = &ApplicationStub{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "38408e83-e8eb-4f82-abe1-7fa02cee0657")
				c.RegisterProcess(&ProcessMessageHandlerStub{
					ConfigureFunc: func(c dogma.ProcessConfigurer) {
						c.Identity("<process>", "1c0dd111-fe12-4dee-a8bc-64abea1dce8f")
						c.Routes(
							dogma.HandlesEvent[EventStub[TypeA]](),
							dogma.ExecutesCommand[CommandStub[TypeA]](),
						)
					},
					RouteEventToInstanceFunc: func(
						context.Context,
						dogma.Event,
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

	g.It("dispatches the message", func() {
		test.Prepare(
			RecordEvent(EventA1),
		)

		Expect(buf.Facts()).To(ContainElement(
			fact.DispatchCycleBegun{
				Envelope: &envelope.Envelope{
					MessageID:     "1",
					CausationID:   "1",
					CorrelationID: "1",
					Message:       EventA1,
					Type:          message.TypeOf(EventA1),
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

	g.It("fails the test if the message type is unrecognized", func() {
		t.FailSilently = true

		test.Prepare(
			RecordEvent(EventX1),
		)

		Expect(t.Failed()).To(BeTrue())
		Expect(t.Logs).To(ContainElement(
			"cannot record event, stubs.EventStub[TypeX] is a not a recognized message type",
		))
	})

	g.It("fails the test if the message type is not an event", func() {
		t.FailSilently = true

		test.Prepare(
			RecordEvent(CommandA1),
		)

		Expect(t.Failed()).To(BeTrue())
		Expect(t.Logs).To(ContainElement(
			"cannot record event, stubs.CommandStub[TypeA] is configured as a command",
		))
	})

	g.It("does not satisfy its own expectations", func() {
		t.FailSilently = true

		test.Expect(
			RecordEvent(EventA1),
			ToRecordEvent(EventA1),
		)

		Expect(t.Failed()).To(BeTrue())
	})

	g.It("produces the expected caption", func() {
		test.Prepare(
			RecordEvent(EventA1),
		)

		Expect(t.Logs).To(ContainElement(
			"--- recording stubs.EventStub[TypeA] event ---",
		))
	})

	g.It("panics if the message is nil", func() {
		Expect(func() {
			RecordEvent(nil)
		}).To(PanicWith("RecordEvent(<nil>): message must not be nil"))
	})

	g.It("captures the location that the action was created", func() {
		act := recordEvent(EventA1)
		Expect(act.Location()).To(MatchAllFields(
			Fields{
				"Func": Equal("github.com/dogmatiq/testkit_test.recordEvent"),
				"File": HaveSuffix("/action.linenumber_test.go"),
				"Line": Equal(53),
			},
		))
	})
})
