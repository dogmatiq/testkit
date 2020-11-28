package report

import (
	"github.com/dogmatiq/testkit/fact"
)

// Stage encapsulates the activity and findings that occurs within one part of
// the test.
type Stage struct {
	// Caption is a brief description of what resulted in this activity.
	//
	// It must not be empty. It should be given in lower case without a trailing
	// period, exclamation or question mark, similar to how Go error messages
	// are formatted.
	Caption string

	// Result is the result of the stage.
	Result TestResult

	// Transcript is a history of what occurred during this stage.
	Transcript Transcript

	// Content is arbitrary additional content.
	Content []Content

	// Findings is the set of discoveries made by analysing the transcript.
	Findings []Finding
}

// StageBuilder builds a Stage.
type StageBuilder struct {
	report *Report
	stage  Stage
}

// BuildTranscriptFact returns a TranscriptFactBuilder that adds a fact to the
// stage's transcript.
func (b *StageBuilder) BuildTranscriptFact(f fact.Fact) *TranscriptFactBuilder {
	return &TranscriptFactBuilder{
		&b.stage,
		TranscriptFact{
			Fact:         f,
			ResultBefore: b.report.TestResult,
		},
	}
}

// AddTranscriptLog adds an arbitrary log message to the stage's transcript.
func (b *StageBuilder) AddTranscriptLog(
	format string,
	args ...interface{},
) {
	b.stage.Transcript = append(
		b.stage.Transcript,
		TranscriptLog{
			format,
			args,
		},
	)
}

// AddContent adds arbitrary content to the stage.
func (b *StageBuilder) AddContent(heading, body string) {
	b.stage.Content = append(
		b.stage.Content,
		Content{
			heading,
			body,
		})
}

// Done adds the stage to the report.
func (b *StageBuilder) Done() Stage {
	b.report.Stages = append(b.report.Stages, b.stage)
	return b.stage
}
