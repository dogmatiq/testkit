package report

import (
	"context"

	"github.com/dogmatiq/testkit/fact"
)

// Transcript is a history of what occurred during a test.
type Transcript []TranscriptEntry

// TranscriptEntry is an entry within transcript.
type TranscriptEntry interface {
	// AcceptVisitor calls the appropriate method on v, based on the entry type.
	AcceptVisitor(ctx context.Context, v TranscriptEntryVisitor) error
}

// TranscriptEntryVisitor visits different kinds of transcript entries.
type TranscriptEntryVisitor interface {
	VisitTranscriptFact(context.Context, TranscriptFact) error
	VisitTranscriptLog(context.Context, TranscriptLog) error
}

// TranscriptFact is a TranscriptEntry for a Fact produced by the engine.
type TranscriptFact struct {
	Fact fact.Fact

	// ResultBefore is the result of the test as it was immediately before this
	// fact was emitted.
	ResultBefore TestResult

	// ResultAfter is the result of the test as it was immediately after this
	// fact was emitted.
	ResultAfter TestResult
}

// AcceptVisitor calls the appropriate method on v, based on the entry type.
func (e TranscriptFact) AcceptVisitor(
	ctx context.Context,
	v TranscriptEntryVisitor,
) error {
	return v.VisitTranscriptFact(ctx, e)
}

// TranscriptLog is a TranscriptEntry for an arbitrary log message.
type TranscriptLog struct {
	LogMessage FormatString
}

// AcceptVisitor calls the appropriate method on v, based on the entry type.
func (e TranscriptLog) AcceptVisitor(
	ctx context.Context,
	v TranscriptEntryVisitor,
) error {
	return v.VisitTranscriptLog(ctx, e)
}

// TranscriptFactBuilder builds a TranscriptFact.
type TranscriptFactBuilder struct {
	report *Report
	entry  TranscriptFact
}

// BuildFinding returns a FindingBuilder that adds a finding that was made as a
// result of this fact to the report.
func (b *TranscriptFactBuilder) BuildFinding(
	p FindingPolarity,
	caption string,
	args ...interface{},
) *FindingBuilder {
	return &FindingBuilder{
		b.addFinding,
		Finding{
			Polarity: p,
			Caption:  FormatString{caption, args},
		},
	}
}

// Done adds the fact to the transcript.
func (b *TranscriptFactBuilder) Done() TranscriptFact {
	b.report.Transcript = append(b.report.Transcript, b.entry)
	return b.entry
}

func (b *TranscriptFactBuilder) addFinding(f Finding) {
	if f.Polarity == Negative {
		b.report.TestResult = Failed
	}

	b.report.Findings = append(b.report.Findings, f)
}
