package testkit

import (
	"reflect"

	"github.com/dogmatiq/dogma"
	"google.golang.org/protobuf/proto"
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
	if pa, ok := a.(proto.Message); ok {
		if pb, ok := b.(proto.Message); ok {
			return proto.Equal(pa, pb)
		}
	}

	return reflect.DeepEqual(a, b)
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
