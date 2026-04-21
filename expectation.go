package testkit

import (
	"github.com/dogmatiq/enginekit/config"
	"github.com/dogmatiq/testkit/fact"
)

// An Expectation describes some criteria that may be met by an action.
type Expectation interface {
	// Caption returns the caption that should be used for this action in the
	// test report.
	Caption() string

	// Predicate returns a new predicate that checks that this expectation is
	// satisfied.
	//
	// The predicate must be closed by calling Done() once the action it tests
	// is completed.
	Predicate(s PredicateScope) Predicate
}

// Predicate tests whether a specific Action satisfies an Expectation.
type Predicate interface {
	fact.Observer

	// Ok returns true if the expectation tested by this predicate has been met.
	//
	// The return value may change as the predicate is notified of additional
	// facts. It must return a consistent value once Done() has been called.
	Ok() bool

	// Done finalizes the predicate.
	//
	// The behavior of the predicate is undefined if it is notified of any
	// additional facts after Done() has been called, or if Done() is called
	// more than once.
	Done()

	// Report returns a report describing whether or not the expectation was
	// met.
	//
	// The behavior of Report() is undefined if Done() has not been called.
	Report(ReportGenerationContext) *Report
}

// PredicateScope encapsulates the element's of a Test's state that may be
// inspected by Predicate implementations.
type PredicateScope struct {
	// App is the application being tested.
	App *config.Application

	// Options contains values that dictate how the predicate should behave.
	// The options are provided by the Test and the Action being performed.
	Options PredicateOptions
}

// PredicateOptions contains values that dictate how a predicate should behave.
type PredicateOptions struct {
	// MessageComparator is the comparator to use when testing two messages for
	// equality.
	MessageComparator MessageComparator

	// MatchDispatchCycleStartedFacts controls whether predicates that look for
	// specific messages should consider messages from DispatchCycleStarted
	// facts.
	//
	// If it is false, the predicate must only match against messages produced
	// by handlers.
	MatchDispatchCycleStartedFacts bool
}

// failingPredicate is a [Predicate] that always fails, reporting a
// pre-determined criteria and explanation.
type failingPredicate struct {
	criteria    string
	explanation string
}

func (p *failingPredicate) Notify(fact.Fact) {}
func (p *failingPredicate) Ok() bool         { return false }
func (p *failingPredicate) Done()            {}
func (p *failingPredicate) Report(ctx ReportGenerationContext) *Report {
	return &Report{
		TreeOk:      ctx.TreeOk,
		Ok:          false,
		Criteria:    p.criteria,
		Explanation: p.explanation,
	}
}
