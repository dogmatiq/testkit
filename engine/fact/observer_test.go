package fact_test

import (
	"github.com/dogmatiq/configkit"
	. "github.com/dogmatiq/testkit/engine/fact"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("type ObserverGroup", func() {
	Describe("func Notify()", func() {
		It("notifies each of the observers in the group", func() {
			f := HandlingBegun{
				HandlerIdentity: configkit.MustNewIdentity("<handler-1>", "<handler-1-key>"),
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
			f1 := HandlingBegun{
				HandlerIdentity: configkit.MustNewIdentity("<handler-1>", "<handler-1-key>"),
			}
			f2 := HandlingBegun{
				HandlerIdentity: configkit.MustNewIdentity("<handler-2>", "<handler-2-key>"),
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
			Ignore.Notify(HandlingBegun{
				HandlerIdentity: configkit.MustNewIdentity("<handler-1>", "<handler-1-key>"),
			})
		})
	})
})
