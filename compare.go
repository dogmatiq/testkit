package testkit

import (
	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/testkit/internal/compare"
)

// A MessageComparator is a function that returns true if two messages are
// considered equal.
type MessageComparator func(a, b dogma.Message) bool

// DefaultMessageComparator returns true if a and b are considered equal.
//
// It is the default implementation of the MessageComparator type.
//
// It supports comparison of protocol buffers messages using the proto.Equal()
// function. All other types are compared using reflect.DeepEqual().
func DefaultMessageComparator(a, b dogma.Message) bool {
	return compare.Equal(a, b)
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
