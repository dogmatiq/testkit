package dogmatest

import (
	"testing"

	"github.com/dogmatiq/dogmatest/engine"
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

// testOptions is a container for the options set via TestOption values.
type testOptions struct {
	operationOptions []engine.OperationOption
	verbose          bool
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
	}

	for _, opt := range options {
		opt(ro)
	}

	return ro
}

// WithOperationOption returns a TestOption that applies optional per-operation
// settings when performing assertions.
func WithOperationOption(opt engine.OperationOption) TestOption {
	return func(ro *testOptions) {
		ro.operationOptions = append(ro.operationOptions, opt)
	}
}