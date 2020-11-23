package report

// Transcript is a record of the activity that occurred as the result of
// some specific step within a test.
type Transcript struct {
	// Caption is a brief description of what resulted in this activity.
	//
	// It must not be empty. It should be given in lower case without a trailing
	// period, exclamation or question mark, similar to how Go error messages
	// are formatted.
	Caption string

	// Entries contains the human-readable entries in the transcript.
	Entries []string
}
