package testkit

import (
	"github.com/dogmatiq/testkit/fact"
)

// An Expectation is a predicate for determining whether some specific criteria
// was met while performing an action.
type Expectation interface {
	fact.Observer

	// Banner returns a human-readable banner to display in the logs when this
	// expectation is used.
	//
	// The banner text should be in uppercase, and complete the sentence "The
	// application is expected ...". For example, "TO DO A THING".
	Banner() string

	// Begin is called to prepare the expectation for a new test.
	Begin(o ExpectOptionSet)

	// End is called once the test is complete.
	End()

	// Ok returns true if the expectation passed.
	Ok() bool

	// BuildReport generates a report about the expectation.
	//
	// ok is true if the expectation is considered to have passed. This may not be
	// the same value as returned from Ok() when this expectation is used as a child
	// of a composite expectation.
	BuildReport(ok bool) *Report
}

type predicateBasedExpectation interface {
	Expectation

	// Predicate returns a new predicate that checks that this expectation is
	// satisfied.
	//
	// The predicate must be closed by calling Done() once the action it tests
	// is completed.
	Predicate(o PredicateOptions) Predicate
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

	// Report returns a report describing whether or not the expectation
	// was met.
	//
	// treeOk is true if the entire "tree" of expectations is considered to have
	// passed. This may not be the same value as returned from Ok() when this
	// expectation is used as a child of a composite expectation.
	//
	// The behavior of Report() is undefined if Done() has not been called.
	Report(treeOk bool) *Report
}

// ExpectOptionSet TODO REMOVE
type ExpectOptionSet = PredicateOptions

// PredicateOptions contains values that dictate how a predicate should behave.
type PredicateOptions struct {
	// MatchDispatchCycleStartedFacts controls whether predicates that look for
	// specific messages should consider messages from DispatchCycleStarted
	// facts.
	//
	// If it is false, the predicate must only match against messages produced
	// by handlers.
	MatchDispatchCycleStartedFacts bool
}
