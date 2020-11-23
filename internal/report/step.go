package report

import (
	"github.com/dogmatiq/testkit/fact"
)

// Step encapsulates the activity and findings that occurs within one "step" of
// the test.
type Step struct {
	// Caption is a brief description of what resulted in this activity.
	//
	// It must not be empty. It should be given in lower case without a trailing
	// period, exclamation or question mark, similar to how Go error messages
	// are formatted.
	Caption string

	// Transcript is the set of facts emitted by the engine.
	Transcript []TranscriptEntry

	// Findings is the set of discoveries made by analysing the transcript.
	Findings []Finding
}

// TranscriptEntry is an entry within an XXX's transcript.
type TranscriptEntry struct {
	// Fact is the fact emitted by the engine.
	Fact fact.Fact

	// ResultBefore is the result of the test as it was immediately before this
	// fact was emitted.
	ResultBefore TestResult

	// ResultAfter is the result of the test as it was immediately after this
	// fact was emitted.
	ResultAfter TestResult
}
