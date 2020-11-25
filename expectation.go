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
