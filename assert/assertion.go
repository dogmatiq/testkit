package assert

import (
	"github.com/dogmatiq/testkit/compare"
	"github.com/dogmatiq/testkit/engine/fact"
	"github.com/dogmatiq/testkit/report"
)

// ExpectOptionSet is a set of options that dictate the behavior of the
// Test.Expect() method.
type ExpectOptionSet struct {
	// MessageComparator compares two messages for equality.
	MessageComparator compare.Comparator

	// MatchMessagesInDispatchCycle controls whether expectations should match
	// messages from the start of a dispatch cycle.
	//
	// If it is false, only messages produced by handlers within the application
	// are matched.
	MatchMessagesInDispatchCycle bool
}

// An Assertion is a predicate for determining whether some specific criteria
// was met during a test.
type Assertion interface {
	fact.Observer

	// Banner returns a human-readable banner to display in the logs when this
	// expectation is used.
	//
	// The banner text should be in uppercase, and complete the sentence "The
	// application is expected ...". For example, "TO DO A THING".
	Banner() string

	// Begin is called to prepare the assertion for a new test.
	Begin(o ExpectOptionSet)

	// End is called once the test is complete.
	End()

	// Ok returns true if the assertion passed.
	Ok() bool

	// BuildReport generates a report about the assertion.
	//
	// ok is true if the assertion is considered to have passed. This may not be
	// the same value as returned from Ok() when this assertion is used as a
	// sub-assertion inside a composite.
	BuildReport(ok bool, r report.Renderer) *Report
}
