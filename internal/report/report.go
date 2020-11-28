package report

import "github.com/dogmatiq/testkit/fact"

// Report is a report on the behavior and result of a test or a portion thereof.
type Report struct {
	// TestResult is the final result of the test.
	TestResult TestResult

	// FailureMode is a brief description of the way that the test failed.
	//
	// It may be empty if the test passed.
	//
	// It should be given in lower case without a trailing period, exclamation
	// or question mark, similar to how Go error messages are formatted.
	//
	// It should describe the result of the test as succinctly and directly as
	// possible. Assume that the reader will not see any further details of the
	// report.
	FailureMode FormatString

	// Caption is a brief description of what resulted in this activity.
	//
	// It must not be empty. It should be given in lower case without a trailing
	// period, exclamation or question mark, similar to how Go error messages
	// are formatted.
	Caption FormatString

	// Transcript is a history of what occurred during this portion of the test.
	Transcript Transcript

	// Content is arbitrary additional content.
	Content []Content

	// Findings is the set of discoveries made by analysing the transcript.
	Findings []Finding
}

// TestResult is an enumeration of possible results for a test.
type TestResult int

const (
	// Failed indicates that the test failed.
	Failed TestResult = iota

	// Passed indicates that a test passed.
	Passed
)

// Builder builds a Report.
type Builder struct {
	report Report
}

// New returns a Builder for a new report.
func New(caption string, args ...interface{}) *Builder {
	return &Builder{
		Report{
			Caption: FormatString{caption, args},
		},
	}
}

// BuildTranscriptFact returns a TranscriptFactBuilder that adds a fact to the
// report's transcript.
func (b *Builder) BuildTranscriptFact(f fact.Fact) *TranscriptFactBuilder {
	return &TranscriptFactBuilder{
		&b.report,
		TranscriptFact{
			Fact:         f,
			ResultBefore: b.report.TestResult,
		},
	}
}

// AddTranscriptLog adds an arbitrary log message to the reports's transcript.
func (b *Builder) AddTranscriptLog(
	format string,
	args ...interface{},
) {
	b.report.Transcript = append(
		b.report.Transcript,
		TranscriptLog{
			FormatString{format, args},
		},
	)
}

// AddContent adds arbitrary content to the report.
func (b *Builder) AddContent(c Content) {
	b.report.Content = append(b.report.Content, c)
}

// Done completes and returns the report.
func (b *Builder) Done() Report {
	return b.report
}
