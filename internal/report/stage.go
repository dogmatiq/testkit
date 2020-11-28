package report

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
