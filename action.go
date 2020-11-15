package testkit

import (
	"context"

	"github.com/dogmatiq/testkit/engine"
)

// Action is an interface for any action that can be performed within a test.
//
// Actions always attempt to cause some state change within the engine or
// application.
type Action interface {
	// Heading returns a human-readable description of the action, used as a
	// heading within the test report.
	//
	// Any engine activity as a result of this action is logged beneath this
	// heading.
	Heading() string

	// ExpectOptions returns the options to use by default when this action is
	// used with Test.Expect().
	ExpectOptions() []ExpectOption

	// Apply performs the action within the context of a specific test.
	Apply(
		ctx context.Context,
		t *Test,
		options []engine.OperationOption,
	) error
}
