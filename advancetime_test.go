package testkit_test

import (
	"time"

	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogma/fixtures"
	. "github.com/dogmatiq/testkit"
	"github.com/dogmatiq/testkit/engine"
	"github.com/dogmatiq/testkit/engine/fact"
	"github.com/dogmatiq/testkit/internal/testingmock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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

		GinkgoT()

		t = &testingmock.T{
			FailSilently: true,
		}

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

	It("retains the virtual time between calls", func() {
		test.Prepare(
			AdvanceTime(ByDuration(1*time.Second)),
			AdvanceTime(ByDuration(1*time.Second)),
		)

		Expect(buf.Facts()).To(ContainElement(
			fact.TickCycleBegun{
				EngineTime: startTime.Add(2 * time.Second),
				EnabledHandlers: map[configkit.HandlerType]bool{
					configkit.AggregateHandlerType:   true,
					configkit.IntegrationHandlerType: false,
					configkit.ProcessHandlerType:     true,
					configkit.ProjectionHandlerType:  false,
				},
			},
		))
	})

	It("panics if the mutation produces a time in the past", func() {
		test.Prepare(
			AdvanceTime(
				ByDuration(-1 * time.Second),
			),
		)

		Expect(t.Failed).To(BeTrue())
		Expect(t.Logs).To(ContainElement(
			"adjusting the clock by -1s would reverse time",
		))
	})

	When("passed a ToTime() mutation", func() {
		targetTime := time.Date(2100, 1, 2, 3, 4, 5, 6, time.UTC)

		It("advances the clock to the provided time", func() {
			test.Prepare(
				AdvanceTime(ToTime(targetTime)),
			)

			Expect(buf.Facts()).To(ContainElement(
				fact.TickCycleBegun{
					EngineTime: targetTime,
					EnabledHandlers: map[configkit.HandlerType]bool{
						configkit.AggregateHandlerType:   true,
						configkit.IntegrationHandlerType: false,
						configkit.ProcessHandlerType:     true,
						configkit.ProjectionHandlerType:  false,
					},
				},
			))
		})

		It("logs a suitable heading", func() {
			test.Prepare(
				AdvanceTime(ToTime(targetTime)),
			)

			Expect(t.Logs).To(ContainElement(
				"--- PREPARE: ADVANCING TIME (to 2100-01-02T03:04:05Z) ---",
			))
		})
	})

	When("passed a ByDuration() mutation", func() {
		It("advances the clock then performs a tick", func() {
			test.Prepare(
				AdvanceTime(ByDuration(3 * time.Second)),
			)

			Expect(buf.Facts()).To(ContainElement(
				fact.TickCycleBegun{
					EngineTime: startTime.Add(3 * time.Second),
					EnabledHandlers: map[configkit.HandlerType]bool{
						configkit.AggregateHandlerType:   true,
						configkit.IntegrationHandlerType: false,
						configkit.ProcessHandlerType:     true,
						configkit.ProjectionHandlerType:  false,
					},
				},
			))
		})

		It("logs a suitable heading", func() {
			test.Prepare(
				AdvanceTime(ByDuration(3 * time.Second)),
			)

			Expect(t.Logs).To(ContainElement(
				"--- PREPARE: ADVANCING TIME (by 3s) ---",
			))
		})
	})
})
