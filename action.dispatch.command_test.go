package testkit_test

import (
	"time"

	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/enginekit/enginetest/stubs"
	. "github.com/dogmatiq/testkit"
	"github.com/dogmatiq/testkit/engine"
	"github.com/dogmatiq/testkit/envelope"
	"github.com/dogmatiq/testkit/fact"
	"github.com/dogmatiq/testkit/internal/testingmock"
	g "github.com/onsi/ginkgo/v2"
	gm "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = g.Describe("func ExecuteCommand()", func() {
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
				c.Identity("<app>", "a84b2620-4675-4024-b55b-cd5dbeb6e293")
				c.RegisterIntegration(&IntegrationMessageHandlerStub{
					ConfigureFunc: func(c dogma.IntegrationConfigurer) {
						c.Identity("<integration>", "d1cf3af1-6c20-4125-8e68-192a6075d0b4")
						c.Routes(
							dogma.HandlesCommand[CommandStub[TypeA]](),
							dogma.RecordsEvent[EventStub[TypeA]](),
						)
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
			ExecuteCommand(CommandA1),
		)

		gm.Expect(buf.Facts()).To(gm.ContainElement(
			fact.DispatchCycleBegun{
				Envelope: &envelope.Envelope{
					MessageID:     "1",
					CausationID:   "1",
					CorrelationID: "1",
					Message:       CommandA1,
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
			ExecuteCommand(CommandX1),
		)

		gm.Expect(t.Failed()).To(gm.BeTrue())
		gm.Expect(t.Logs).To(gm.ContainElement(
			"cannot execute command, stubs.CommandStub[TypeX] is a not a recognized message type",
		))
	})

	g.It("does not satisfy its own expectations", func() {
		t.FailSilently = true

		test.Expect(
			ExecuteCommand(CommandA1),
			ToExecuteCommand(CommandA1),
		)

		gm.Expect(t.Failed()).To(gm.BeTrue())
	})

	g.It("produces the expected caption", func() {
		test.Prepare(
			ExecuteCommand(CommandA1),
		)

		gm.Expect(t.Logs).To(gm.ContainElement(
			"--- executing stubs.CommandStub[TypeA] command ---",
		))
	})

	g.It("panics if the message is nil", func() {
		gm.Expect(func() {
			ExecuteCommand(nil)
		}).To(gm.PanicWith("ExecuteCommand(<nil>): message must not be nil"))
	})

	g.It("captures the location that the action was created", func() {
		act := executeCommand(CommandA1)
		gm.Expect(act.Location()).To(MatchAllFields(
			Fields{
				"Func": gm.Equal("github.com/dogmatiq/testkit_test.executeCommand"),
				"File": gm.HaveSuffix("/action.linenumber_test.go"),
				"Line": gm.Equal(52),
			},
		))
	})
})
