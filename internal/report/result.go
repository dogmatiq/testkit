package report

// TestResult is an enumeration of possible results for a test or one of the
// stages within it.
type TestResult int

const (
	// Failed indicates that the test failed.
	Failed TestResult = iota

	// Passed indicates that a test passed.
	Passed
)
