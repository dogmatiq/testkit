package report

import (
	"context"

	"github.com/dogmatiq/testkit/fact"
)

// Transcript is a history of what occurred during a Stage.
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
func (e TranscriptFact) AcceptVisitor(ctx context.Context, v TranscriptEntryVisitor) error {
	return v.VisitTranscriptFact(ctx, e)
}

// TranscriptLog is a TranscriptEntry for an arbitrary log message.
type TranscriptLog struct {
	Message string
}

// AcceptVisitor calls the appropriate method on v, based on the entry type.
func (e TranscriptLog) AcceptVisitor(ctx context.Context, v TranscriptEntryVisitor) error {
	return v.VisitTranscriptLog(ctx, e)
}
