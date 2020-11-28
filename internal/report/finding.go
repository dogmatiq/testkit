package report

// A Finding is a piece of information discovered by observing the engine
// throughout the lifetime of a test.
type Finding struct {
	// Polarity indicates how the finding influenced test result, if at all.
	Polarity FindingPolarity

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
	Caption FormatString

	// Summary is a brief paragraph describing the finding.
	//
	// It may be empty. If provided, it should be in sentence case, punctuated
	// normally.
	//
	// If the Finding is a result of a failed Expectation the summary should
	// give the best explanation as to why the failure occurred.
	//
	// For example, use "The handler that records this event has been disabled."
	// in preference to "The expected event was not recorded.".
	Summary string

	// Evidence contains other findings that led to this finding.
	Evidence []Finding

	// Content is arbitrary additional content.
	Content []Content

	// Suggestions is a collection of recommended actions.
	//
	// For negative findings the suggestions would typically describe how to
	// change the application (or the test, in some cases) to prevent the same
	// negative finding in the future.
	Suggestions []Suggestion
}

// FindingPolarity is an enumeration of the "polarity" of a finding, which
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

// FindingBuilder builds a Finding.
type FindingBuilder struct {
	done    func(Finding)
	finding Finding
}

// Summary adds an optional summary to the finding.
func (b *FindingBuilder) Summary(s string) {
	b.finding.Summary = s
}

// BuildEvidence returns a FindingBuilder that adds a finding to this finding as
// "supporting evidence".
func (b *FindingBuilder) BuildEvidence(
	p FindingPolarity,
	caption string,
	args ...interface{},
) *FindingBuilder {
	return &FindingBuilder{
		b.addEvidence,
		Finding{
			Polarity: p,
			Caption:  FormatString{caption, args},
		},
	}
}

// AddContent adds arbitrary content to the finding.
func (b *FindingBuilder) AddContent(c Content) {
	b.finding.Content = append(b.finding.Content, c)
}

// Suggestion adds a suggestion to the finding.
//
// A suggestion describes some recommended action that improves the Dogma
// application or otherwise fixes a problem encountered during a test.
//
// c is the suggestion caption, a brief description of what resulted in this
// activity. It must not be empty. It should be lower case without a
// trailing period, exclamation or question mark, similar to how Go error
// messages are formatted.
func (b *FindingBuilder) Suggestion(
	con SuggestionConfidence,
	caption string,
	args ...interface{},
) {
	b.finding.Suggestions = append(
		b.finding.Suggestions,
		Suggestion{
			con,
			FormatString{caption, args},
		},
	)
}

// Done marks the finding as complete.
func (b *FindingBuilder) Done() Finding {
	b.done(b.finding)
	return b.finding
}

func (b FindingBuilder) addEvidence(f Finding) {
	b.finding.Evidence = append(b.finding.Evidence, f)
}
