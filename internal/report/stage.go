package report

import "github.com/dogmatiq/testkit/fact"

// Stage encapsulates the activity and findings that occurs within one part of
// the test.
type Stage struct {
	// Caption is a brief description of what resulted in this activity.
	//
	// It must not be empty. It should be given in lower case without a trailing
	// period, exclamation or question mark, similar to how Go error messages
	// are formatted.
	Caption string

	// Transcript is a history of what occurred during this stage.
	Transcript Transcript

	// Content is arbitrary additional content.
	Content []Content

	// Findings is the set of discoveries made by analysing the transcript.
	Findings []Finding
}

// StageBuilder builds a report stage, which encapsulates
// the activity and findings that occurs within one part of the test.
type StageBuilder struct{}

// TranscriptFact add a fact entry to the stage's transcript.
func (b *StageBuilder) TranscriptFact(f fact.Fact) {
	panic("not implemented")
}

// TranscriptLog adds an arbitrary log message to the stage's transcript.
func (b *StageBuilder) TranscriptLog(format string, args ...interface{}) {
	panic("not implemented")
}

// Content adds arbitrary content to the stage.
func (b *StageBuilder) Content(heading, body string) {
	panic("not implemented")
}

// Finding adds a new finding to the stage. A finding is some discovery made
// by observing the engine throughout the lifetime of a test.
//
// c is the finding's caption, a brief description of what was discovered.
// It must not be empty. It should be lower case without a trailing period,
// exclamation or question mark, similar to how Go error messages are
// formatted.
//
// If the Finding is a result of an Expectation the caption should be
// phrased in terms of the expectation and not the cause of the
// expectation's passing or failing.
//
// For example, use "the expected DepositSubmitted event was not recorded"
// in preference to "no events were recorded at all".
func (b *StageBuilder) Finding(p FindingPolarity, c string) FindingBuilder {
	panic("not implemented")
}

// Done marks the stage as complete.
func (b *StageBuilder) Done() Stage {
	panic("not implemented")
}
