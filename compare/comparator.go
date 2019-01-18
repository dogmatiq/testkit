package compare

import (
	"reflect"

	"github.com/dogmatiq/dogma"
)

// Comparator is an interface for comparing messages.
type Comparator interface {
	// MessageIsEqual returns true if a and b are equal.
	MessageIsEqual(a, b dogma.Message) bool

	// AggregateRootIsEqual returns true if a and b are equal.
	AggregateRootIsEqual(a, b dogma.AggregateRoot) bool

	// ProcessRootIsEqual returns true if a and b are equal.
	ProcessRootIsEqual(a, b dogma.ProcessRoot) bool
}

// DefaultComparator is the default comparator implementation.
//
// All values are using reflect.DeepEqual().
type DefaultComparator struct{}

// MessageIsEqual returns true if a and b are equal.
func (r DefaultComparator) MessageIsEqual(a, b dogma.Message) bool {
	return r.isEqual(a, b)
}

// AggregateRootIsEqual returns true if a and b are equal.
func (r DefaultComparator) AggregateRootIsEqual(a, b dogma.AggregateRoot) bool {
	return r.isEqual(a, b)
}

// ProcessRootIsEqual returns true if a and b are equal.
func (r DefaultComparator) ProcessRootIsEqual(a, b dogma.ProcessRoot) bool {
	return r.isEqual(a, b)
}

func (r DefaultComparator) isEqual(a, b interface{}) bool {
	return reflect.DeepEqual(a, b)
}
