package report

import (
	"math"
)

// A Suggestion describes some recommended action that improve's the Dogma
// application or otherwise fixes a problem encountered during a test.
type Suggestion struct {
	// Confidence is a subjective rating of how effective this suggestion is
	// likely to be, relative to other suggestions.
	//
	// The AbsoluteConfidence constant is used within negative findings to
	// indicate that a suggestion is a definitive and appropriate solution to
	// the problem.
	Confidence Confidence

	// Caption is a brief description of the suggestion.
	//
	// It must not be empty. It should be given in lower case without a trailing
	// period, exclamation or question mark, similar to how Go error messages
	// are formatted.
	Caption string
}

// Confidence is a subjective rating of a suggestion's likely effectiveness.
type Confidence uint8

// AbsoluteConfidence indicates that a suggestion is a definitive and
// appropriate solution the problem described by a negative finding.
const AbsoluteConfidence Confidence = math.MaxUint8
