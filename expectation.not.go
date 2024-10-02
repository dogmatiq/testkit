package testkit

import (
	"fmt"

	"github.com/dogmatiq/testkit/fact"
)

// Not is an expectation that passes only if the given expectation fails.
func Not(expectation Expectation) Expectation {
	return &notExpectation{
		caption:     fmt.Sprintf("not %s", expectation.Caption()),
		expectation: expectation,
	}
}

// notExpectation is an [Expectation] that inverts another expectation.
//
// It uses a predicate function to determine whether the not expectation is met.
type notExpectation struct {
	caption     string
	expectation Expectation
}

func (e *notExpectation) Caption() string {
	return e.caption
}

func (e *notExpectation) Predicate(s PredicateScope) (Predicate, error) {
	p, err := e.expectation.Predicate(s)
	if err != nil {
		return nil, err
	}

	return &notPredicate{
		expectation: p,
	}, nil
}

// notPredicate is the [Predicate] implementation for [notExpectation].
type notPredicate struct {
	expectation Predicate
}

func (p *notPredicate) Notify(f fact.Fact) {
	p.expectation.Notify(f)
}

func (p *notPredicate) Ok() bool {
	return !p.expectation.Ok()
}

func (p *notPredicate) Done() {
	p.expectation.Done()
}

func (p *notPredicate) Report(ctx ReportGenerationContext) *Report {
	ctx.IsInverted = !ctx.IsInverted

	r := p.expectation.Report(ctx)

	rep := &Report{
		TreeOk:   ctx.TreeOk,
		Ok:       p.Ok(),
		Criteria: "do not " + r.Criteria,
	}

	return rep
}
