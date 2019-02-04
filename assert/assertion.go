package assert

import (
	"github.com/dogmatiq/testkit/compare"
	"github.com/dogmatiq/testkit/engine/fact"
	"github.com/dogmatiq/testkit/render"
)

// Assertion is a predicate that checks if some specific critiria was met during
// the execution of a test.
type Assertion interface {
	fact.Observer

	// Prepare is called to prepare the assertion for a new test.
	//
	// c is the comparator used to compare messages and other entities.
	Prepare(c compare.Comparator)

	// Ok returns true if the assertion passed.
	Ok() bool

	// BuildReport generates a report about the assertion.
	//
	// ok is true if the assertion is considered to have passed. This may not be
	// the same value as returned from Ok() when this assertion is used as
	// sub-assertion inside a composite.
	BuildReport(ok bool, r render.Renderer) *Report
}
