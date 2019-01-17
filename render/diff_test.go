package render_test

import (
	"io"
	"testing"

	. "github.com/dogmatiq/dogmatest/render"
	"github.com/dogmatiq/iago"
	"github.com/dogmatiq/iago/iotest"
)

func TestDiff(t *testing.T) {
	iotest.TestWrite(
		t,
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
}
