package report

// Report is a report on the behavior and result of a test.
type Report struct {
	// TestResult is the final result of the test.
	TestResult bool

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
	FailureMode string

	// Stages contains details of each step performed within the test.
	Stages []Stage
}

// TestResult is an enumeration of possible test results.
type TestResult int

const (
	// Failed indicates that the test failed.
	Failed TestResult = iota

	// Passed indicates that a test passed.
	Passed
)
