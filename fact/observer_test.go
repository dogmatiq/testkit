package fact_test

import (
	"time"

	. "github.com/dogmatiq/testkit/fact"
	g "github.com/onsi/ginkgo/v2"
	gm "github.com/onsi/gomega"
)

var _ = g.Describe("type ObserverGroup", func() {
	g.Describe("func Notify()", func() {
		g.It("notifies each of the observers in the group", func() {
			f := TickCycleBegun{}

			n := 0
			g := ObserverGroup{
				ObserverFunc(func(of Fact) {
					n++
					gm.Expect(of).To(gm.Equal(f))
				}),
				ObserverFunc(func(of Fact) {
					n++
					gm.Expect(of).To(gm.Equal(f))
				}),
			}

			g.Notify(f)

			gm.Expect(n).To(gm.Equal(2))
		})
	})
})

var _ = g.Describe("type Buffer", func() {
	g.Describe("func Notify()", func() {
		g.It("appends the fact to the buffer", func() {
			f1 := TickCycleBegun{
				EngineTime: time.Now(),
			}
			f2 := TickCycleBegun{
				EngineTime: time.Now().Add(1 * time.Second),
			}
			b := &Buffer{}

			b.Notify(f1)
			b.Notify(f2)

			gm.Expect(b.Facts()).To(gm.Equal([]Fact{
				f1,
				f2,
			}))
		})
	})
})

var _ = g.Describe("var Ignore", func() {
	g.Describe("func Notify()", func() {
		g.It("does nothing", func() {
			Ignore.Notify(TickCycleBegun{})
		})
	})
})
