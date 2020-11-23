package report

// Report is a report on the behavior and result of a test.
type Report struct {
	// Passed is true if the test passed.
	Passed bool

	// FailureMode is a brief description of the way that the test failed.
	//
	// It should be given in lower case without a trailing period, exclamation
	// or question mark, similar to how Go error messages are formatted.
	//
	// It should describe the result of the test as succinctly and directly as
	// possible. Assume that the reader will not see any further details of the
	// report.
	FailureMode string

	// Transcripts describes the activity that occurred during the test.
	Transcripts []Transcript

	// Findings is the set of discoveries made by analysing the test activity.
	Findings []Finding
}
