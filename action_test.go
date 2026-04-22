package testkit_test

import (
	"context"

	. "github.com/dogmatiq/testkit"
	"github.com/dogmatiq/testkit/location"
)

// noop is an Action that does nothing.
var noop noopAction

// noopAction is an Action that does nothing. It is intended to be used for
// testing the test system itself.
type noopAction struct {
	err error
}

func (a noopAction) Caption() string                             { return "[no-op]" }
func (a noopAction) Location() location.Location                 { return location.Location{Func: "<noop>"} }
func (a noopAction) ConfigurePredicate(*PredicateOptions)        {}
func (a noopAction) Do(ctx context.Context, s ActionScope) error { return a.err }
