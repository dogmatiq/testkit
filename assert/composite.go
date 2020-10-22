package assert

import (
	"fmt"

	"github.com/dogmatiq/testkit/compare"
	"github.com/dogmatiq/testkit/engine/fact"
	"github.com/dogmatiq/testkit/render"
)

// AllOf returns an assertion that passes if all of the given sub-assertions
// pass.
func AllOf(subs ...Assertion) Assertion {
	n := len(subs)

	if n == 0 {
		panic("no sub-assertions provided")
	}

	if n == 1 {
		return subs[0]
	}

	return &compositeAssertion{
		Criteria:      "all of",
		SubAssertions: subs,
		Predicate: func(p int) (string, bool) {
			n := len(subs)

			if p == n {
				return "", true
			}

			return fmt.Sprintf(
				"%d of the sub-assertions failed",
				n-p,
			), false
		},
	}
}

// AnyOf returns an assertion that passes if at least one of the given
// sub-assertions passes.
func AnyOf(subs ...Assertion) Assertion {
	n := len(subs)

	if n == 0 {
		panic("no sub-assertions provided")
	}

	if n == 1 {
		return subs[0]
	}

	return &compositeAssertion{
		Criteria:      "any of",
		SubAssertions: subs,
		Predicate: func(p int) (string, bool) {
			if p > 0 {
				return "", true
			}

			return fmt.Sprintf(
				"all %d of the sub-assertions failed",
				n,
			), false
		},
	}
}

// NoneOf returns an assertion that passes if all of the given sub-assertions
// fail.
func NoneOf(subs ...Assertion) Assertion {
	n := len(subs)

	if n == 0 {
		panic("no sub-assertions provided")
	}

	return &compositeAssertion{
		Criteria:      "none of",
		SubAssertions: subs,
		Predicate: func(p int) (string, bool) {
			if p == 0 {
				return "", true
			}

			if n == 1 {
				return "the sub-assertion passed unexpectedly", false
			}

			return fmt.Sprintf(
				"%d of the sub-assertions passed unexpectedly",
				p,
			), false
		},
	}
}

// compositeAssertion is an assertion that is a container for other assertions.
type compositeAssertion struct {
	// Criteria is a brief description of the assertion's requirement to pass.
	Criteria string

	// SubAssertions is the set of assertions in the container.
	SubAssertions []Assertion

	// Predicate is a function that determines whether or not the assertion
	// passes, based on the number of child assertions that passed.
	//
	// It returns true if the assertion passed, and may optionally return a
	// message to be displayed in either case.
	Predicate func(int) (string, bool)

	// ok is a pointer to the assertion's result.
	//
	// It is nil if Ok() has not been called.
	ok *bool

	// outcome is the messagestring from the predicate function.
	//
	// It is populated by Ok().
	outcome string
}

// Notify notifies the assertion of the occurrence of a fact.
func (a *compositeAssertion) Notify(f fact.Fact) {
	for _, sub := range a.SubAssertions {
		sub.Notify(f)
	}
}

// Begin is called to prepare the assertion for a new test.
//
// c is the comparator used to compare messages and other entities.
func (a *compositeAssertion) Begin(c compare.Comparator) {
	a.ok = nil
	a.outcome = ""

	for _, sub := range a.SubAssertions {
		sub.Begin(c)
	}
}

// End is called once the test is complete.
func (a *compositeAssertion) End() {
	for _, sub := range a.SubAssertions {
		sub.End()
	}
}

// Ok returns true if the assertion passed.
//
// If asserted is false, the assertion was a no-op and the value of pass is
// meaningless.
func (a *compositeAssertion) Ok() (ok bool, asserted bool) {
	if a.ok != nil {
		return *a.ok, true
	}

	n := 0

	for _, sub := range a.SubAssertions {
		if sub.MustOk() {
			n++
		}
	}

	m, ok := a.Predicate(n)

	a.ok = &ok
	a.outcome = m

	return *a.ok, true
}

// MustOk returns true if the assertion passed.
func (a *compositeAssertion) MustOk() bool {
	ok, _ := a.Ok()
	return ok
}

// BuildReport generates a report about the assertion.
//
// ok is true if the assertion is considered to have passed. This may not be the
// same value as returned from Ok() when this assertion is used as a
// sub-assertion inside a composite.
func (a *compositeAssertion) BuildReport(ok, verbose bool, r render.Renderer) *Report {
	a.MustOk() // populate a.ok and a.outcome

	rep := &Report{
		TreeOk:   ok,
		Ok:       *a.ok,
		Criteria: a.Criteria,
		Outcome:  a.outcome,
	}

	for _, sub := range a.SubAssertions {
		rep.Append(
			sub.BuildReport(ok, verbose, r),
		)
	}

	return rep
}
