package report_test

import (
	"strings"

	. "github.com/dogmatiq/testkit/internal/report"
	g "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = g.Describe("func WriteDiff()", func() {
	g.It("produces a word-diff of the input", func() {
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
