package render_test

import (
	"io"

	"github.com/dogmatiq/iago/iotest"
	"github.com/dogmatiq/iago/must"
	. "github.com/dogmatiq/testkit/render"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("func WriteDiff()", func() {
	It("produces a word-diff of the input", func() {
		iotest.TestWrite(
			GinkgoT(),
			func(w io.Writer) int {
				return must.Must(
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

var _ = Describe("func Diff()", func() {
	It("produces a word-diff of the input", func() {
		Expect(
			Diff(
				"foo bar baz",
				"foo qux baz",
			),
		).To(Equal(
			"foo [-bar-]{+qux+} baz",
		))
	})
})
