package location_test

import (
	. "github.com/dogmatiq/testkit/location"
	g "github.com/onsi/ginkgo/v2"
	gm "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = g.Describe("type Location", func() {
	g.Describe("func OfFunc()", func() {
		g.It("returns the expected location", func() {
			loc := OfFunc(doNothing)

			gm.Expect(loc).To(MatchAllFields(
				Fields{
					"Func": gm.Equal("github.com/dogmatiq/testkit/location_test.doNothing"),
					"File": gm.HaveSuffix("/location/linenumber_test.go"),
					"Line": gm.Equal(50),
				},
			))
		})

		g.It("panics if the value is not a function", func() {
			gm.Expect(func() {
				OfFunc("<not a function>")
			}).To(gm.PanicWith("fn must be a function"))
		})
	})

	g.Describe("func OfMethod()", func() {
		g.It("returns the expected location", func() {
			loc := OfMethod(ofMethodT{}, "Method")

			gm.Expect(loc).To(MatchAllFields(
				Fields{
					"Func": gm.Equal("github.com/dogmatiq/testkit/location_test.ofMethodT.Method"),
					"File": gm.HaveSuffix("/location/linenumber_test.go"),
					"Line": gm.Equal(57),
				},
			))
		})

		g.It("panics if the methods does not exist", func() {
			gm.Expect(func() {
				OfMethod(ofMethodT{}, "DoesNotExist")
			}).To(gm.PanicWith("method does not exist"))
		})
	})

	g.Describe("func OfCall()", func() {
		g.It("returns the expected location", func() {
			loc := ofCallLayer2()

			gm.Expect(loc).To(MatchAllFields(
				Fields{
					"Func": gm.Equal("github.com/dogmatiq/testkit/location_test.ofCallLayer2"),
					"File": gm.HaveSuffix("/location/linenumber_test.go"),
					"Line": gm.Equal(53),
				},
			))
		})
	})

	g.Describe("func OfPanic()", func() {
		g.It("returns the expected location", func() {
			defer func() {
				recover()
				loc := OfPanic()

				gm.Expect(loc).To(MatchAllFields(
					Fields{
						"Func": gm.Equal("github.com/dogmatiq/testkit/location_test.doPanic"),
						"File": gm.HaveSuffix("/location/linenumber_test.go"),
						"Line": gm.Equal(51),
					},
				))
			}()

			doPanic()
		})
	})

	g.Describe("func String()", func() {
		g.DescribeTable(
			"it returns the expected string",
			func(s string, l Location) {
				gm.Expect(l.String()).To(gm.Equal(s))
			},
			g.Entry("empty", "<unknown>", Location{}),
			g.Entry("function name only", "<function>(...)", Location{Func: "<function>"}),
			g.Entry("function name only (global closure)", "<function glob..>(...)", Location{Func: "<function glob..>"}),
			g.Entry("file location only", "<file>:123", Location{File: "<file>", Line: 123}),
			g.Entry("both", "<file>:123 [<function>(...)]", Location{Func: "<function>", File: "<file>", Line: 123}),
			g.Entry("both (global closure)", "<file>:123", Location{Func: "<function glob..>", File: "<file>", Line: 123}),
		)
	})
})
