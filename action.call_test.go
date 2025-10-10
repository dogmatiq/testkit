package testkit_test

import (
	"context"
	"time"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/enginekit/config"
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

var _ = g.Describe("func Call()", func() {
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
				c.Identity("<app>", "b51cde16-aaec-4d75-ae14-06282e3a72c8")
				c.Routes(
					dogma.ViaIntegration(&IntegrationMessageHandlerStub{
						ConfigureFunc: func(c dogma.IntegrationConfigurer) {
							c.Identity("<integration>", "832d78d7-a006-414f-b6d7-3153aa7c9ab8")
							c.Routes(
								dogma.HandlesCommand[*CommandStub[TypeA]](),
							)
						},
					}),
				)
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
					CommandA1,
				)
			}),
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
				EnabledHandlerTypes: map[config.HandlerType]bool{
					config.AggregateHandlerType:   true,
					config.IntegrationHandlerType: false,
					config.ProcessHandlerType:     true,
					config.ProjectionHandlerType:  false,
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
					CommandA1,
				)
			}),
			ToExecuteCommand(CommandA1),
		)
	})

	g.It("produces the expected caption", func() {
		test.Prepare(
			Call(func() {}),
		)

		gm.Expect(t.Logs).To(gm.ContainElement(
			"--- calling user-defined function ---",
		))
	})

	g.It("panics if the function is nil", func() {
		gm.Expect(func() {
			Call(nil)
		}).To(gm.PanicWith("Call(<nil>): function must not be nil"))
	})

	g.It("captures the location that the action was created", func() {
		act := call(func() {})
		gm.Expect(act.Location()).To(MatchAllFields(
			Fields{
				"Func": gm.Equal("github.com/dogmatiq/testkit_test.call"),
				"File": gm.HaveSuffix("/action.linenumber_test.go"),
				"Line": gm.Equal(51),
			},
		))
	})
})
