package assert

import (
	"github.com/dogmatiq/dogmatest/compare"
	"github.com/dogmatiq/dogmatest/engine/fact"
	"github.com/dogmatiq/dogmatest/render"
)

// CompositeAssertion is an assertion that is a container for other assertions.
type CompositeAssertion struct {
	// Criteria is a brief description of the assertion's requirement to pass.
	Criteria string

	// SubAssertions is the set of assertions in the container.
	SubAssertions []Assertion

	// Predicate is a function that determines whether or not the assertion passes,
	// based on the number of child assertions that passed.
	//
	// It returns true if the assertion passed, and may optionally return a message
	// to be displayed in either case.
	Predicate func(int) (string, bool)
}

// Notify notifies the assertion of the occurrence of a fact.
func (a *CompositeAssertion) Notify(f fact.Fact) {
	for _, sub := range a.SubAssertions {
		sub.Notify(f)
	}
}

// Begin is called before the test is executed.
//
// c is the comparator used to compare messages and other entities.
func (a *CompositeAssertion) Begin(c compare.Comparator) {
	for _, sub := range a.SubAssertions {
		sub.Begin(c)
	}
}

// End is called after the test is executed.
//
// It returns the result of the assertion.
func (a *CompositeAssertion) End(r render.Renderer) *Result {
	res := &Result{
		Criteria: a.Criteria,
	}

	n := 0

	for _, sub := range a.SubAssertions {
		sr := sub.End(r)

		if sr.Ok {
			n++
		}

		res.AppendResult(sr)
	}

	res.Outcome, res.Ok = a.Predicate(n)

	return res
}
