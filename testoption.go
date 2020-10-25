package testkit

import (
	"time"

	"github.com/dogmatiq/testkit/engine"
)

// TestOption applies optional settings to a test.
type TestOption func(*testOptions)

// WithStartTime returns a test option that sets the time of the test runner's
// clock at the start of the test.
//
// By default, the current system time is used.
func WithStartTime(t time.Time) TestOption {
	return func(to *testOptions) {
		to.time = t
	}
}

// testOptions is a container for the options set via TestOption values.
type testOptions struct {
	operationOptions []engine.OperationOption
	time             time.Time
}

// newTestOptions returns a new testOptions with the given options.
func newTestOptions(options []TestOption) *testOptions {
	ro := &testOptions{
		operationOptions: []engine.OperationOption{
			engine.EnableIntegrations(false),
			engine.EnableProjections(false),
		},
		time: time.Now(),
	}

	for _, opt := range options {
		opt(ro)
	}

	return ro
}

// WithOperationOptions returns a TestOption that applies optional per-operation
// settings when performing assertions.
func WithOperationOptions(options ...engine.OperationOption) TestOption {
	return func(ro *testOptions) {
		ro.operationOptions = append(ro.operationOptions, options...)
	}
}
