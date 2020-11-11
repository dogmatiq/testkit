package engine_test

import (
	"context"
	"time"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogma/fixtures"
	. "github.com/dogmatiq/testkit/engine"
	"github.com/dogmatiq/testkit/engine/fact"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("func Run()", func() {
	var (
		app    *Application
		engine *Engine
	)

	BeforeEach(func() {
		app = &Application{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "<app-key>")
			},
		}

		var err error
		engine, err = New(app)
		Expect(err).ShouldNot(HaveOccurred())
	})

	It("calls tick repeatedly", func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func() {
			time.Sleep(16 * time.Millisecond)
			cancel()
		}()

		buf := &fact.Buffer{}
		Run(
			ctx,
			engine,
			5*time.Millisecond,
			WithObserver(buf),
		)

		facts := buf.Facts()
		Expect(len(facts)).To(BeNumerically(">=", 6))

		for i := 0; i < 6; i += 2 {
			Expect(facts[i]).To(BeAssignableToTypeOf(fact.TickCycleBegun{}))
			Expect(facts[i+1]).To(BeAssignableToTypeOf(fact.TickCycleCompleted{}))
		}
	})

	It("returns an error if the context is canceled while ticking", func() {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := Run(ctx, engine, 0)
		Expect(err).To(Equal(context.Canceled))
	})

	It("returns an error if the context is canceled between ticks", func() {
		ctx, cancel := context.WithCancel(context.Background())

		go func() {
			time.Sleep(10 * time.Millisecond)
			cancel()
		}()

		err := Run(ctx, engine, 0)
		Expect(err).To(Equal(context.Canceled))
	})
})

var _ = Describe("func RunTimeScaled()", func() {
	var (
		app    *Application
		engine *Engine
	)

	BeforeEach(func() {
		app = &Application{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "<app-key>")
			},
		}

		var err error
		engine, err = New(app)
		Expect(err).ShouldNot(HaveOccurred())
	})

	It("scales type by the given factor", func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func() {
			time.Sleep(30 * time.Millisecond)
			cancel()
		}()

		epoch := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

		buf := &fact.Buffer{}

		RunTimeScaled(
			ctx,
			engine,
			10*time.Millisecond,
			0.5,
			epoch,
			WithObserver(buf),
		)

		facts := buf.Facts()
		Expect(len(facts)).To(BeNumerically(">=", 6))

		t := facts[0].(fact.TickCycleBegun).EngineTime
		Expect(t).To(BeTemporally(">=", epoch))
		Expect(t).To(BeTemporally("<", epoch.Add(5*time.Millisecond)))

		t = facts[2].(fact.TickCycleBegun).EngineTime
		Expect(t).To(BeTemporally(">=", epoch.Add(5*time.Millisecond)))
		Expect(t).To(BeTemporally("<", epoch.Add(10*time.Millisecond)))

		t = facts[4].(fact.TickCycleBegun).EngineTime
		Expect(t).To(BeTemporally(">=", epoch.Add(10*time.Millisecond)))
		Expect(t).To(BeTemporally("<", epoch.Add(15*time.Millisecond)))
	})
})
