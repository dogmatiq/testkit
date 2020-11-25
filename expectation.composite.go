package testkit

import (
	"fmt"

	"github.com/dogmatiq/testkit/fact"
)

// AllOf is an expectation that passes only if all of its children pass.
func AllOf(children ...Expectation) Expectation {
	n := len(children)

	if n == 0 {
		panic("AllOf(): at least one child expectation must be provided")
	}

	if n == 1 {
		return children[0]
	}

	return &compositeExpectation{
		BannerText: fmt.Sprintf("TO MEET %d EXPECTATIONS", n),
		Criteria:   "all of",
		Children:   children,
		Predicate: func(passed int) (string, bool) {
			if passed == n {
				return "", true
			}

			return fmt.Sprintf(
				"%d of the expectations failed",
				n-passed,
			), false
		},
	}
}

// AnyOf is an expectation that passes if any of its children pass.
func AnyOf(children ...Expectation) Expectation {
	n := len(children)

	if n == 0 {
		panic("AnyOf(): at least one child expectation must be provided")
	}

	if n == 1 {
		return children[0]
	}

	return &compositeExpectation{
		BannerText: fmt.Sprintf("TO MEET AT LEAST ONE OF %d EXPECTATIONS", n),
		Criteria:   "any of",
		Children:   children,
		Predicate: func(passed int) (string, bool) {
			if passed > 0 {
				return "", true
			}

			return fmt.Sprintf(
				"all %d of the expectations failed",
				n,
			), false
		},
	}
}

// NoneOf is an expectation that passes only if all of its children fail.
func NoneOf(children ...Expectation) Expectation {
	n := len(children)

	if n == 0 {
		panic("NoneOf(): at least one child expectation must be provided")
	}

	banner := fmt.Sprintf("NOT TO MEET ANY OF %d EXPECTATIONS", n)
	if n == 1 {
		banner = fmt.Sprintf("NOT %s", children[0].Banner())
	}

	return &compositeExpectation{
		BannerText: banner,
		Criteria:   "none of",
		Children:   children,
		Predicate: func(passed int) (string, bool) {
			if passed == 0 {
				return "", true
			}

			if n == 1 {
				return "the expectation passed unexpectedly", false
			}

			return fmt.Sprintf(
				"%d of the expectations passed unexpectedly",
				passed,
			), false
		},
	}
}

// compositeExpectation is an expectation that runs several child expectations.
// Its final result is determined by a predicate function.
type compositeExpectation struct {
	// BannerText is a human-readable banner to display in the logs when this
	// expectation is used.
	BannerText string

	// Criteria is a brief description of the expectation that must be met.
	Criteria string

	// Children is the expectation's child expectations.
	Children []Expectation

	// Predicate is a function that determines whether or not the expectation
	// passes, based on the number of children that passed.
	//
	// It returns true if the expectation passed, and may optionally return a
	// message to be displayed in either case.
	Predicate func(int) (string, bool)

	// ok is a cache of the expectation's result.
	//
	// It is populated the first time Ok() is called.
	ok *bool

	// outcome is the message string returned by the predicate function.
	//
	// It is populated the first time Ok() is called.
	outcome string
}

// Banner returns a human-readable banner to display in the logs when this
// expectation is used.
//
// The banner text should be in uppercase, and complete the sentence "The
// application is expected ...". For example, "TO DO A THING".
func (e *compositeExpectation) Banner() string {
	return e.BannerText
}

// Notify notifies the expectation of the occurrence of a fact.
func (e *compositeExpectation) Notify(f fact.Fact) {
	for _, c := range e.Children {
		c.Notify(f)
	}
}

// Begin is called to prepare the expectation for a new test.
func (e *compositeExpectation) Begin(o ExpectOptionSet) {
	e.ok = nil
	e.outcome = ""

	for _, c := range e.Children {
		c.Begin(o)
	}
}

// End is called once the test is complete.
func (e *compositeExpectation) End() {
	for _, c := range e.Children {
		c.End()
	}
}

// Ok returns true if the expectation passed.
func (e *compositeExpectation) Ok() bool {
	if e.ok != nil {
		return *e.ok
	}

	n := 0

	for _, c := range e.Children {
		if c.Ok() {
			n++
		}
	}

	m, ok := e.Predicate(n)

	e.ok = &ok
	e.outcome = m

	return *e.ok
}

// BuildReport generates a report about the expectation.
//
// ok is true if the expectation is considered to have passed. This may not be
// the same value as returned from Ok() when this expectation is used as a child
// of a composite expectation.
func (e *compositeExpectation) BuildReport(ok bool) *Report {
	e.Ok() // populate e.ok and e.outcome

	rep := &Report{
		TreeOk:   ok,
		Ok:       *e.ok,
		Criteria: e.Criteria,
		Outcome:  e.outcome,
	}

	for _, c := range e.Children {
		rep.Append(
			c.BuildReport(ok),
		)
	}

	return rep
}
