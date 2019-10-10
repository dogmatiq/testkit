package testkit

import (
	"testing"
	"time"

	"github.com/dogmatiq/testkit/engine"
)

// TestOption applies optional settings to a test.
type TestOption func(*testOptions)

// Verbose returns a test option that enables or disables verbose test output
// for an individual test.
//
// By default, tests produce verbose output if the -v flag is passed to "go test".
func Verbose(enabled bool) TestOption {
	return func(to *testOptions) {
		to.verbose = enabled
	}
}

// StartTime returns a test option that sets the initial time of the test clock.
//
// By default, the current system time is used.
func StartTime(t time.Time) TestOption {
	return func(to *testOptions) {
		to.time = t
	}
}

// testOptions is a container for the options set via TestOption values.
type testOptions struct {
	operationOptions []engine.OperationOption
	verbose          bool
	time             time.Time
}

// newTestOptions returns a new testOptions with the given options.
func newTestOptions(options []TestOption, verbose *bool) *testOptions {
	var v bool
	if verbose == nil {
		// note: testing.Verbose() is called here instead of in New() so that New()
		// can be called during package initialization, at which time
		// testing.Verbose() will always return false.
		v = testing.Verbose()
	} else {
		v = *verbose
	}

	ro := &testOptions{
		operationOptions: []engine.OperationOption{
			engine.EnableIntegrations(false),
			engine.EnableProjections(false),
		},
		verbose: v,
		time:    time.Now(),
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
