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

// WithOperationOptions returns a TestOption that applies optional per-operation
// settings when performing assertions.
func WithOperationOptions(options ...engine.OperationOption) TestOption {
	return func(t *Test) {
		t.operationOptions = append(t.operationOptions, options...)
	}
}
