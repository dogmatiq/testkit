package testkit

import (
	"time"

	"github.com/dogmatiq/testkit/engine"
)

// TestOption applies optional settings to a test.
type TestOption interface {
	applyTestOption(*Test)
}

type testOptionFunc func(*Test)

func (f testOptionFunc) applyTestOption(t *Test) {
	f(t)
}

// StartTimeAt returns a test option that sets the initial time of the test's
// virtual clock.
//
// By default, the current system time is used.
func StartTimeAt(st time.Time) TestOption {
	return testOptionFunc(func(t *Test) {
		t.virtualClock = st
	})
}

// WithMessageComparator returns a test option that sets the comparator
// to be used when comparing messages for equality.
//
// This effects the ToExecuteCommand() and ToRecordEvent() expectations.
//
// By default, DefaultMessageComparator is used.
func WithMessageComparator(c MessageComparator) TestOption {
	return testOptionFunc(func(t *Test) {
		t.predicateOptions.MessageComparator = c
	})
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
	return testOptionFunc(func(t *Test) {
		t.operationOptions = append(t.operationOptions, options...)
	})
}
