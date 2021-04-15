package location_test

import (
	. "github.com/dogmatiq/testkit/location"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("type Location", func() {
	Describe("func OfFunc()", func() {
		It("returns the expected location", func() {
			loc := OfFunc(doNothing)

			Expect(loc).To(MatchAllFields(
				Fields{
					"Func": Equal("github.com/dogmatiq/testkit/location_test.doNothing"),
					"File": HaveSuffix("/location/linenumber_test.go"),
					"Line": Equal(50),
				},
			))
		})

		It("panics if the value is not a function", func() {
			Expect(func() {
				OfFunc("<not a function>")
			}).To(PanicWith("fn must be a function"))
		})
	})

	Describe("func OfMethod()", func() {
		It("returns the expected location", func() {
			loc := OfMethod(ofMethodT{}, "Method")

			Expect(loc).To(MatchAllFields(
				Fields{
					"Func": Equal("github.com/dogmatiq/testkit/location_test.ofMethodT.Method"),
					"File": HaveSuffix("/location/linenumber_test.go"),
					"Line": Equal(57),
				},
			))
		})

		It("panics if the methods does not exist", func() {
			Expect(func() {
				OfMethod(ofMethodT{}, "DoesNotExist")
			}).To(PanicWith("method does not exist"))
		})
	})

	Describe("func OfCall()", func() {
		It("returns the expected location", func() {
			loc := ofCallLayer2()

			Expect(loc).To(MatchAllFields(
				Fields{
					"Func": Equal("github.com/dogmatiq/testkit/location_test.ofCallLayer2"),
					"File": HaveSuffix("/location/linenumber_test.go"),
					"Line": Equal(53),
				},
			))
		})
	})

	Describe("func OfPanic()", func() {
		It("returns the expected location", func() {
			defer func() {
				recover()
				loc := OfPanic()

				Expect(loc).To(MatchAllFields(
					Fields{
						"Func": Equal("github.com/dogmatiq/testkit/location_test.doPanic"),
						"File": HaveSuffix("/location/linenumber_test.go"),
						"Line": Equal(51),
					},
				))
			}()

			doPanic()
		})
	})

	Describe("func String()", func() {
		DescribeTable(
			"it returns the expected string",
			func(s string, l Location) {
				Expect(l.String()).To(Equal(s))
			},
			Entry("empty", "<unknown>", Location{}),
			Entry("function name only", "<function>(...)", Location{Func: "<function>"}),
			Entry("function name only (global closure)", "<function glob..>(...)", Location{Func: "<function glob..>"}),
			Entry("file location only", "<file>:123", Location{File: "<file>", Line: 123}),
			Entry("both", "<file>:123 [<function>(...)]", Location{Func: "<function>", File: "<file>", Line: 123}),
			Entry("both (global closure)", "<file>:123", Location{Func: "<function glob..>", File: "<file>", Line: 123}),
		)
	})
})
