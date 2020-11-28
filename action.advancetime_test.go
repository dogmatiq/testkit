package testkit_test

import (
	"fmt"
	"time"

	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogma/fixtures"
	. "github.com/dogmatiq/testkit"
	"github.com/dogmatiq/testkit/engine"
	"github.com/dogmatiq/testkit/fact"
	"github.com/dogmatiq/testkit/internal/testingmock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("func AdvanceTime()", func() {
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

	It("retains the virtual time between calls", func() {
		test.Prepare(
			AdvanceTime(ByDuration(1*time.Second)),
			AdvanceTime(ByDuration(1*time.Second)),
		)

		Expect(buf.Facts()).To(ContainElement(
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

	It("fails the test if time is reversed", func() {
		t.FailSilently = true

		target := startTime.Add(-1 * time.Second)

		test.Prepare(
			AdvanceTime(
				ToTime(target),
			),
		)

		Expect(t.Failed()).To(BeTrue())
		Expect(t.Logs).To(ContainElement(
			fmt.Sprintf(
				"adjusting the clock to %s would reverse time",
				target.Format(time.RFC3339),
			),
		))
	})

	It("panics if the adjustment is nil", func() {
		Expect(func() {
			AdvanceTime(nil)
		}).To(PanicWith("AdvanceTime(<nil>): adjustment must not be nil"))
	})

	It("captures the location that the action was created", func() {
		act := advanceTime(ByDuration(10 * time.Second))
		Expect(act.Location()).To(MatchAllFields(
			Fields{
				"Func": Equal("github.com/dogmatiq/testkit_test.advanceTime"),
				"File": HaveSuffix("/action.linenumber_test.go"),
				"Line": Equal(50),
			},
		))
	})

	When("passed a ToTime() adjustment", func() {
		targetTime := time.Date(2100, 1, 2, 3, 4, 5, 6, time.UTC)

		It("advances the clock to the provided time", func() {
			test.Prepare(
				AdvanceTime(ToTime(targetTime)),
			)

			Expect(buf.Facts()).To(ContainElement(
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

		It("produces the expected caption", func() {
			test.Prepare(
				AdvanceTime(ToTime(targetTime)),
			)

			Expect(t.Logs).To(ContainElement(
				"--- advancing time to 2100-01-02T03:04:05Z ---",
			))
		})
	})

	When("passed a ByDuration() adjustment", func() {
		It("advances the clock then performs a tick", func() {
			test.Prepare(
				AdvanceTime(ByDuration(3 * time.Second)),
			)

			Expect(buf.Facts()).To(ContainElement(
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

		It("produces the expected caption", func() {
			test.Prepare(
				AdvanceTime(ByDuration(3 * time.Second)),
			)

			Expect(t.Logs).To(ContainElement(
				"--- advancing time by 3s ---",
			))
		})

		It("panics if the duration is negative", func() {
			Expect(func() {
				ByDuration(-1 * time.Second)
			}).To(PanicWith("ByDuration(-1s): duration must not be negative"))
		})
	})
})
