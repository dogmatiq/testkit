package assert

import (
	"io"

	"github.com/dogmatiq/dogmatest/compare"
	"github.com/dogmatiq/dogmatest/engine/fact"
	"github.com/dogmatiq/dogmatest/render"
)

// Assertion is a predicate that checks if some specific critiria was met during
// the execution of a test.
type Assertion interface {
	fact.Observer

	// Begin is called before the message-under-test is dispatched.
	Begin(compare.Comparator)

	// End is called after the message-under-test is dispatched.
	// The assertion writes a report of its result to w.
	// It returns true if the assertion passed.
	End(io.Writer, render.Renderer) bool
}
