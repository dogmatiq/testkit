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
		caption:  fmt.Sprintf("to meet %d expectations", n),
		criteria: "all of",
		children: children,
		pred: func(passed int) (string, bool) {
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
		caption:  fmt.Sprintf("to meet at least one of %d expectations", n),
		criteria: "any of",
		children: children,
		pred: func(passed int) (string, bool) {
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

	caption := fmt.Sprintf("not to meet any of %d expectations", n)
	if n == 1 {
		caption = fmt.Sprintf("not %s", children[0].Caption())
	}

	return &compositeExpectation{
		caption:  caption,
		criteria: "none of",
		children: children,
		pred: func(passed int) (string, bool) {
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
		isInverted: true,
	}
}

// compositeExpectation is an Expectation that contains other expectations.
//
// It is the implementation used by AllOf(), AnyOf() and NoneOf().
//
// It uses a predicate function to determine whether the composite expectation
// is met based on how many of the "child" expectations are met.
type compositeExpectation struct {
	caption    string
	criteria   string
	children   []Expectation
	pred       func(passed int) (outcome string, ok bool)
	isInverted bool
}

func (e *compositeExpectation) Caption() string {
	return e.caption
}

func (e *compositeExpectation) Predicate(s PredicateScope) (Predicate, error) {
	var children []Predicate

	for _, c := range e.children {
		p, err := c.Predicate(s)
		if err != nil {
			return nil, err
		}

		children = append(children, p)
	}

	return &compositePredicate{
		criteria:   e.criteria,
		children:   children,
		pred:       e.pred,
		isInverted: e.isInverted,
	}, nil
}

// compositePredicate is the Predicate implementation for compositeExpectation.
type compositePredicate struct {
	criteria   string
	children   []Predicate
	pred       func(int) (string, bool)
	isInverted bool
}

func (p *compositePredicate) Notify(f fact.Fact) {
	for _, c := range p.children {
		c.Notify(f)
	}
}

func (p *compositePredicate) Ok() bool {
	_, ok := p.ok()
	return ok
}

func (p *compositePredicate) Done() {
	for _, c := range p.children {
		c.Done()
	}
}

func (p *compositePredicate) Report(ctx ReportGenerationContext) *Report {
	if p.isInverted {
		ctx.IsInverted = !ctx.IsInverted
	}

	m, ok := p.ok()

	rep := &Report{
		TreeOk:   ctx.TreeOk,
		Ok:       ok,
		Criteria: p.criteria,
		Outcome:  m,
	}

	for _, c := range p.children {
		rep.Append(c.Report(ctx))
	}

	return rep
}

func (p *compositePredicate) ok() (string, bool) {
	n := 0

	for _, c := range p.children {
		if c.Ok() {
			n++
		}
	}

	return p.pred(n)
}
