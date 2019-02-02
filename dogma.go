package dogmatest

// This file (dogma.go) is used to write log messages to the *testing.T struct
// (via the T interface), so that the test output more clearly indicates that
// the log message originated within Dogma's tooling, as the Go test framework
// includes the filename in log messages.

// log logs a message during a test.
func log(t T, args ...interface{}) {
	t.Log(args...)
}

// logf logs a formatted message during a test.
func logf(t T, f string, args ...interface{}) {
	t.Logf(f, args...)
}
