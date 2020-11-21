package testkit

import (
	"time"

	"github.com/dogmatiq/testkit/engine"
)

// TestOption applies optional settings to a test.
type TestOption func(*Test)

// StartVirtualClockAt returns a test option that sets the initial time of the
// test's virtual clock.
//
// By default, the current system time is used.
func StartVirtualClockAt(st time.Time) TestOption {
	return func(t *Test) {
		t.virtualClock = st
	}
}

// WithUnsafeOperationOptions returns a TestOption that applies a set of engine
// operation options when performing any action.
//
// This function is provided for forward-compatibility with engine operations
// and for low level control of the engine's behavior.
//
// The provided options may override options that the Test sets during its
// normal operation and should be used with caution.
func WithUnsafeOperationOptions(options ...engine.OperationOption) TestOption {
	return func(t *Test) {
		t.operationOptions = append(t.operationOptions, options...)
	}
}
