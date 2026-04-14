package report_test

import (
	"strings"
	"testing"

	. "github.com/dogmatiq/testkit/internal/report"
	"github.com/dogmatiq/testkit/internal/test"
)

func TestWriteDiff(t *testing.T) {
	t.Run("it produces a word-diff of the input", func(t *testing.T) {
		var w strings.Builder

		WriteDiff(
			&w,
			"foo bar baz",
			"foo qux baz",
		)

		test.Expect(
			t,
			"unexpected diff",
			w.String(),
			"foo [-bar-]{+qux+} baz",
		)
	})
}
