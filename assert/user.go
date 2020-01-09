package assert

import (
	"github.com/dogmatiq/testkit/compare"
	"github.com/dogmatiq/testkit/engine/fact"
	"github.com/dogmatiq/testkit/render"
)

// AssertionContext is passed to user-defined assertion functions.
type AssertionContext struct {
	// Comparator provides logic for comparing messages and application state.
	Comparator compare.Comparator

	// Facts is an ordered slice of the facts that occurred.
	Facts []fact.Fact
}

// Should returns an assertion that uses a user-defined function to check for
// specific criteria.
//
// cr is a human-readable description of the expectation of the assertion. It
// should be phrased as an imperative statement, such as "insert a customer".
//
// fn is the function that performs the assertion logic. It returns a non-nil
// error to indicate an assertion failure. It is passed an AssertionContext
// which contains dependencies and engine state that can be used to implement
// the assertion logic.
func Should(
	cr string,
	fn func(AssertionContext) error,
) Assertion {
	return &userAssertion{
		criteria: cr,
		assert:   fn,
	}
}

// userAssertion is a user-defined assertion.
type userAssertion struct {
	criteria string
	assert   func(AssertionContext) error
	ctx      AssertionContext
	done     bool
	err      error
}

// Notify the observer of a fact.
func (a *userAssertion) Notify(f fact.Fact) {
	a.ctx.Facts = append(a.ctx.Facts, f)
}

// Prepare is called to prepare the assertion for a new test.
//
// c is the comparator used to compare messages and other entities.
func (a *userAssertion) Prepare(c compare.Comparator) {
	a.ctx.Comparator = c
}

// Ok returns true if the assertion passed.
func (a *userAssertion) Ok() bool {
	if !a.done {
		a.err = a.assert(a.ctx)
		a.done = true
	}

	return a.err == nil
}

// BuildReport generates a report about the assertion.
//
// ok is true if the assertion is considered to have passed. This may not be
// the same value as returned from Ok() when this assertion is used as a
// sub-assertion inside a composite.
func (a *userAssertion) BuildReport(ok bool, r render.Renderer) *Report {
	rep := &Report{
		TreeOk:   ok,
		Ok:       a.Ok(),
		Criteria: a.criteria,
	}

	if ok || a.Ok() {
		return rep
	}

	rep.Outcome = "the user-defined assertion returned a non-nil error"
	rep.Explanation = a.err.Error()

	return rep
}
