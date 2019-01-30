package assert

import (
	"github.com/dogmatiq/dogmatest/compare"
	"github.com/dogmatiq/dogmatest/engine/fact"
	"github.com/dogmatiq/dogmatest/render"
)

// Assertion is a predicate that checks if some specific critiria was met during
// the execution of a test.
type Assertion interface {
	fact.Observer

	// Begin is called before the test is executed.
	//
	// c is the comparator used to compare messages and other entities.
	Begin(c compare.Comparator)

	// End is called after the test is executed.
	//
	// It returns the result of the assertion.
	End(r render.Renderer) *Result
}
