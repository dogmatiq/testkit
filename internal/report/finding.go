package report

// A Finding is a piece of information discovered by observing the engine
// throughout the lifetime of a test.
type Finding struct {
	// Caption is a brief description of the finding.
	//
	// It must not be empty. It should be given in lower case without a trailing
	// period, exclamation or question mark, similar to how Go error messages
	// are formatted.
	//
	// If the Finding is a result of an Expectation the caption should be
	// phrased in terms of the expectation and not the cause of the
	// expectation's passing or failing.
	//
	// For example, use "the expected DepositSubmitted event was not recorded"
	// in preference to "no events were recorded at all".
	Caption string

	// Summary is a brief paragraph describing the finding.
	//
	// It may be empty. If provided, it should be in sentence case, punctionated
	// normally.
	//
	// If the Finding is a result of a failed Expectation the summary should
	// give the best explanation as to why the failure occurred.
	//
	// For example, use "The handler that records this event has been disabled."
	// in preference to "The expected event was not recorded.".
	Summary string

	// Polarity indicates how the finding influenced test result, if at all.
	Polarity FindingPolarity

	// Evidence contains other findings that led to this finding.
	Evidence []Finding
}

// FindingPolarity is an numerations of the "polarity" of a finding, which
// describes the effect of the finding on the result of a test.
type FindingPolarity int

const (
	// Negative indicates that the finding caused the test to fail. Typically
	// this would take precedence over any Positive finding.
	Negative FindingPolarity = -1

	// Neutral indicates the the finding did not effect the test result.
	Neutral FindingPolarity = -0

	// Positive indicates that the finding caused the test to pass, assuming no
	// Negative findings were discovered.
	Positive FindingPolarity = +1
)
