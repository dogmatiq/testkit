package assert

import (
	"fmt"

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

// AllOf returns an assertion that passes if all of the given sub-assertions pass.
func AllOf(subs ...Assertion) Assertion {
	n := len(subs)

	if n == 0 {
		panic("no sub-assertions provided")
	}

	if n == 1 {
		return subs[0]
	}

	return &CompositeAssertion{
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

	return &CompositeAssertion{
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

	return &CompositeAssertion{
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
