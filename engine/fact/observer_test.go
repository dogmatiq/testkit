package fact_test

import (
	. "github.com/dogmatiq/dogmatest/engine/fact"
	"github.com/dogmatiq/enginekit/fixtures"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("type ObserverGroup", func() {
	Describe("func Notify()", func() {
		It("notifies each of the observers in the group", func() {
			f := DispatchCycleSkipped{
				Message: fixtures.MessageA1,
			}
			n := 0
			g := ObserverGroup{
				ObserverFunc(func(of Fact) {
					n++
					Expect(of).To(Equal(f))
				}),
				ObserverFunc(func(of Fact) {
					n++
					Expect(of).To(Equal(f))
				}),
			}

			g.Notify(f)

			Expect(n).To(Equal(2))
		})
	})
})

var _ = Describe("type Buffer", func() {
	Describe("func Notify()", func() {
		It("appends the fact to the buffer", func() {
			f1 := DispatchCycleSkipped{
				Message: fixtures.MessageA1,
			}
			f2 := DispatchCycleSkipped{
				Message: fixtures.MessageA2,
			}
			b := &Buffer{}

			b.Notify(f1)
			b.Notify(f2)

			Expect(b.Facts()).To(Equal([]Fact{
				f1,
				f2,
			}))
		})
	})
})

var _ = Describe("var Ignore", func() {
	Describe("func Notify()", func() {
		It("does nothing", func() {
			Ignore.Notify(DispatchCycleSkipped{
				Message: fixtures.MessageA1,
			})
		})
	})
})
