package panicx_test

import (
	. "github.com/dogmatiq/testkit/engine/panicx"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("type Location", func() {
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
