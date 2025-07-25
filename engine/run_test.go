package engine_test

import (
	"context"
	"time"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/enginekit/config/runtimeconfig"
	. "github.com/dogmatiq/enginekit/enginetest/stubs"
	. "github.com/dogmatiq/testkit/engine"
	"github.com/dogmatiq/testkit/fact"
	g "github.com/onsi/ginkgo/v2"
	gm "github.com/onsi/gomega"
)

var _ = g.Describe("func Run()", func() {
	var (
		app    *ApplicationStub
		engine *Engine
	)

	g.BeforeEach(func() {
		app = &ApplicationStub{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "9e55f1ed-1f9a-46d9-a01f-e57638f74eb7")
			},
		}

		engine = MustNew(
			runtimeconfig.FromApplication(app),
		)
	})

	g.It("calls tick repeatedly", func() {
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
		gm.Expect(len(facts)).To(gm.BeNumerically(">=", 6))

		for i := 0; i < 6; i += 2 {
			gm.Expect(facts[i]).To(gm.BeAssignableToTypeOf(fact.TickCycleBegun{}))
			gm.Expect(facts[i+1]).To(gm.BeAssignableToTypeOf(fact.TickCycleCompleted{}))
		}
	})

	g.It("returns an error if the context is canceled while ticking", func() {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := Run(ctx, engine, 0)
		gm.Expect(err).To(gm.Equal(context.Canceled))
	})

	g.It("returns an error if the context is canceled between ticks", func() {
		ctx, cancel := context.WithCancel(context.Background())

		go func() {
			time.Sleep(10 * time.Millisecond)
			cancel()
		}()

		err := Run(ctx, engine, 0)
		gm.Expect(err).To(gm.Equal(context.Canceled))
	})
})

var _ = g.Describe("func RunTimeScaled()", func() {
	var (
		app    *ApplicationStub
		engine *Engine
	)

	g.BeforeEach(func() {
		app = &ApplicationStub{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "4f06c58d-b854-41e9-92ee-d4e4ba137670")
			},
		}

		engine = MustNew(
			runtimeconfig.FromApplication(app),
		)
	})

	g.It("scales type by the given factor", func() {
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
		gm.Expect(len(facts)).To(gm.BeNumerically(">=", 6))

		t := facts[0].(fact.TickCycleBegun).EngineTime
		gm.Expect(t).To(gm.BeTemporally(">=", epoch))
		gm.Expect(t).To(gm.BeTemporally("<", epoch.Add(5*time.Millisecond)))

		t = facts[2].(fact.TickCycleBegun).EngineTime
		gm.Expect(t).To(gm.BeTemporally(">=", epoch.Add(5*time.Millisecond)))
		gm.Expect(t).To(gm.BeTemporally("<", epoch.Add(10*time.Millisecond)))

		t = facts[4].(fact.TickCycleBegun).EngineTime
		gm.Expect(t).To(gm.BeTemporally(">=", epoch.Add(10*time.Millisecond)))
		gm.Expect(t).To(gm.BeTemporally("<", epoch.Add(15*time.Millisecond)))
	})
})
