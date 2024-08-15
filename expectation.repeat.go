package testkit

import (
	"fmt"

	"github.com/dogmatiq/testkit/fact"
)

// ToRepeatedly is an expectation that repeats a fixed number of similar
// expectations.
//
// desc is a human-readable description of the expectations. It should be
// phrased as an imperative statement, such as "debit a customer".
//
// f is a factory function that produces the expectations. It is called n times.
// The parameter i indicates the current iteration starting with i == 0, until i
// == n-1. n must be 1 or greater.
func ToRepeatedly(
	desc string,
	n int,
	f func(i int) Expectation,
) Expectation {
	if f == nil {
		panic(fmt.Sprintf("ToRepeatedly(%#v, %d, <nil>): function must not be nil", desc, n))
	}

	if n < 1 {
		panic(fmt.Sprintf("ToRepeatedly(%#v, %d, <func>): n must be 1 or greater", desc, n))
	}

	if desc == "" {
		panic(fmt.Sprintf("ToRepeatedly(%#v, %d, <func>): description must not be empty", desc, n))
	}

	return &repeatExpectation{
		criteria: desc,
		count:    n,
		factory:  f,
	}
}

type repeatExpectation struct {
	criteria string
	count    int
	factory  func(i int) Expectation
}

func (e *repeatExpectation) Caption() string {
	return fmt.Sprintf("to %s", e.criteria)
}

func (e *repeatExpectation) Predicate(s PredicateScope) (Predicate, error) {
	var predicates []Predicate

	for i := 0; i < e.count; i++ {
		x := e.factory(i)

		if x == nil {
			return nil, fmt.Errorf("on iteration %d: factory returned a nil expectation", i)
		}

		p, err := x.Predicate(s)
		if err != nil {
			return nil, fmt.Errorf("on iteration %d: %w", i, err)
		}

		predicates = append(predicates, p)
	}

	return &repeatPredicate{
		criteria: e.criteria,
		children: predicates,
	}, nil
}

// repeatPredicate is the Predicate implementation for repeatExpectation.
type repeatPredicate struct {
	criteria string
	children []Predicate
}

func (p *repeatPredicate) Notify(f fact.Fact) {
	for _, c := range p.children {
		c.Notify(f)
	}
}

func (p *repeatPredicate) Ok() bool {
	for _, c := range p.children {
		if !c.Ok() {
			return false
		}
	}

	return true
}

func (p *repeatPredicate) Done() {
	for _, c := range p.children {
		c.Done()
	}
}

func (p *repeatPredicate) Report(treeOk, isInverted bool) *Report {
	rep := &Report{
		TreeOk:   treeOk,
		Ok:       true,
		Criteria: p.criteria,
	}

	var (
		failureCount int // number of failed iterations
		failureShown int // index of the failed iteration shown on the report
	)

	for i, c := range p.children {
		if c.Ok() {
			continue
		}

		failureCount++

		if failureCount == 1 {
			failureShown = i

			// We only append the failure report from the first failed predicate
			// to the report, as the assumption is the factory produces
			// potentially thousands of very similar expectations which would
			// pollute the report and make it harder to find the problem.
			rep.Append(
				c.Report(treeOk, isInverted),
			)

		}

	}

	if failureCount != 0 {
		rep.Ok = false
		rep.Outcome = fmt.Sprintf(
			"%d of %d iteration(s) failed, iteration #%d shown",
			failureCount,
			len(p.children),
			failureShown,
		)
	}

	return rep
}
