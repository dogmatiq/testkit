package testkit_test

import (
	"fmt"
	"time"

	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/enginekit/enginetest/stubs"
	. "github.com/dogmatiq/testkit"
	"github.com/dogmatiq/testkit/engine"
	"github.com/dogmatiq/testkit/fact"
	"github.com/dogmatiq/testkit/internal/testingmock"
	g "github.com/onsi/ginkgo/v2"
	gm "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = g.Describe("func AdvanceTime()", func() {
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
				c.Identity("<app>", "140ca29b-7a05-4f26-968b-6285255e6d8a")
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

	g.It("retains the virtual time between calls", func() {
		test.Prepare(
			AdvanceTime(ByDuration(1*time.Second)),
			AdvanceTime(ByDuration(1*time.Second)),
		)

		gm.Expect(buf.Facts()).To(gm.ContainElement(
			fact.TickCycleBegun{
				EngineTime: startTime.Add(2 * time.Second),
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

	g.It("fails the test if time is reversed", func() {
		t.FailSilently = true

		target := startTime.Add(-1 * time.Second)

		test.Prepare(
			AdvanceTime(
				ToTime(target),
			),
		)

		gm.Expect(t.Failed()).To(gm.BeTrue())
		gm.Expect(t.Logs).To(gm.ContainElement(
			fmt.Sprintf(
				"adjusting the clock to %s would reverse time",
				target.Format(time.RFC3339),
			),
		))
	})

	g.It("panics if the adjustment is nil", func() {
		gm.Expect(func() {
			AdvanceTime(nil)
		}).To(gm.PanicWith("AdvanceTime(<nil>): adjustment must not be nil"))
	})

	g.It("captures the location that the action was created", func() {
		act := advanceTime(ByDuration(10 * time.Second))
		gm.Expect(act.Location()).To(MatchAllFields(
			Fields{
				"Func": gm.Equal("github.com/dogmatiq/testkit_test.advanceTime"),
				"File": gm.HaveSuffix("/action.linenumber_test.go"),
				"Line": gm.Equal(50),
			},
		))
	})

	g.When("passed a ToTime() adjustment", func() {
		targetTime := time.Date(2100, 1, 2, 3, 4, 5, 6, time.UTC)

		g.It("advances the clock to the provided time", func() {
			test.Prepare(
				AdvanceTime(ToTime(targetTime)),
			)

			gm.Expect(buf.Facts()).To(gm.ContainElement(
				fact.TickCycleBegun{
					EngineTime: targetTime,
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

		g.It("produces the expected caption", func() {
			test.Prepare(
				AdvanceTime(ToTime(targetTime)),
			)

			gm.Expect(t.Logs).To(gm.ContainElement(
				"--- advancing time to 2100-01-02T03:04:05Z ---",
			))
		})
	})

	g.When("passed a ByDuration() adjustment", func() {
		g.It("advances the clock then performs a tick", func() {
			test.Prepare(
				AdvanceTime(ByDuration(3 * time.Second)),
			)

			gm.Expect(buf.Facts()).To(gm.ContainElement(
				fact.TickCycleBegun{
					EngineTime: startTime.Add(3 * time.Second),
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

		g.It("produces the expected caption", func() {
			test.Prepare(
				AdvanceTime(ByDuration(3 * time.Second)),
			)

			gm.Expect(t.Logs).To(gm.ContainElement(
				"--- advancing time by 3s ---",
			))
		})

		g.It("panics if the duration is negative", func() {
			gm.Expect(func() {
				ByDuration(-1 * time.Second)
			}).To(gm.PanicWith("ByDuration(-1s): duration must not be negative"))
		})
	})
})
