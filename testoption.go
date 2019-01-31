package dogmatest

import (
	"github.com/dogmatiq/dogmatest/engine"
)

// TestOption applies optional settings to a test.
type TestOption func(*testOptions)

// testOptions is a container for the options set via TestOption values.
type testOptions struct {
	operationOptions []engine.OperationOption
}

// newTestOptions returns a new testOptions with the given options.
func newTestOptions(options []TestOption) *testOptions {
	ro := &testOptions{
		operationOptions: []engine.OperationOption{
			engine.EnableIntegrations(false),
			engine.EnableProjections(false),
		},
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
