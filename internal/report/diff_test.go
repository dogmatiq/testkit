package report_test

import (
	"strings"

	. "github.com/dogmatiq/testkit/internal/report"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("func WriteDiff()", func() {
	It("produces a word-diff of the input", func() {
		var w strings.Builder

		WriteDiff(
			&w,
			"foo bar baz",
			"foo qux baz",
		)

		Expect(w.String()).To(
			Equal("foo [-bar-]{+qux+} baz"),
		)
	})
})
