package render_test

import (
	"io"

	. "github.com/dogmatiq/dogmatest/render"
	"github.com/dogmatiq/iago"
	"github.com/dogmatiq/iago/iotest"
	. "github.com/onsi/ginkgo"
)

var _ = Describe("func Diff", func() {
	It("produces a word-diff of the input", func() {
		iotest.TestWrite(
			GinkgoT(),
			func(w io.Writer) int {
				return iago.Must(
					WriteDiff(
						w,
						"foo bar baz",
						"foo qux baz",
					),
				)
			},
			"foo [-bar-]{+qux+} baz",
		)
	})
})
