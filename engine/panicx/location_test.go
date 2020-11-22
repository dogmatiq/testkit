package panicx_test

import (
	. "github.com/dogmatiq/testkit/engine/panicx"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("type Location", func() {
	Describe("func LocationOfFunc()", func() {
		It("returns the expected location", func() {
			loc := LocationOfFunc(doNothing)

			Expect(loc).To(MatchAllFields(
				Fields{
					"Func": Equal("github.com/dogmatiq/testkit/engine/panicx_test.doNothing"),
					"File": HaveSuffix("/engine/panicx/linenumber_test.go"),
					"Line": Equal(50),
				},
			))
		})

		It("returns an empty location if the value is not a function", func() {
			Expect(func() {
				LocationOfFunc("<not a function>")
			}).To(PanicWith("fn must be a function"))
		})
	})

	Describe("func LocationOfCall()", func() {
		It("returns the expected location", func() {
			loc := locationOfCallLayer2()

			Expect(loc).To(MatchAllFields(
				Fields{
					"Func": Equal("github.com/dogmatiq/testkit/engine/panicx_test.locationOfCallLayer2"),
					"File": HaveSuffix("/engine/panicx/linenumber_test.go"),
					"Line": Equal(53),
				},
			))
		})
	})

	Describe("func String()", func() {
		DescribeTable(
			"it returns the expectation string",
			func(s string, l Location) {
				Expect(l.String()).To(Equal(s))
			},
			Entry("empty", "<unknown>", Location{}),
			Entry("function name only", "<function>()", Location{Func: "<function>"}),
			Entry("file location only", "<file>:123", Location{File: "<file>", Line: 123}),
			Entry("both", "<function>() <file>:123", Location{Func: "<function>", File: "<file>", Line: 123}),
		)
	})
})
