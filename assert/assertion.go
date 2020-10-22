package assert

import (
	"github.com/dogmatiq/testkit/compare"
	"github.com/dogmatiq/testkit/engine/fact"
	"github.com/dogmatiq/testkit/render"
)

// OptionalAssertion is an interface that accept all Assertion types, as well as
// the Nothing value.
type OptionalAssertion interface {
	fact.Observer

	// Begin is called to prepare the assertion for a new test.
	//
	// c is the comparator used to compare messages and other entities.
	Begin(c compare.Comparator)

	// End is called once the test is complete.
	End()

	// TryOk returns true if the assertion passed.
	//
	// If asserted is false, the assertion was a no-op and ok is meaningless.
	TryOk() (ok bool, asserted bool)

	// BuildReport generates a report about the assertion.
	//
	// ok is true if the assertion is considered to have passed. This may not be
	// the same value as returned from Ok() when this assertion is used as a
	// sub-assertion inside a composite.
	BuildReport(ok, verbose bool, r render.Renderer) *Report
}

// An Assertion is a predicate for determining whether some specific criteria
// was met during a test.
type Assertion interface {
	OptionalAssertion

	// Ok returns true if the assertion passed.
	Ok() bool
}

// Nothing is an "optional assertion" that always passes and does not build an
// report.
var Nothing OptionalAssertion = noopAssertion{}

type noopAssertion struct{}

func (noopAssertion) Notify(fact.Fact)         {}
func (noopAssertion) Begin(compare.Comparator) {}
func (noopAssertion) End()                     {}
func (noopAssertion) TryOk() (bool, bool)      { return false, false }
func (noopAssertion) BuildReport(bool, bool, render.Renderer) *Report {
	panic("not implemented")
}
