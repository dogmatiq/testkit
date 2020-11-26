package testkit_test

import (
	"context"

	. "github.com/dogmatiq/testkit"
)

// noop is an Action that does nothing.
var noop noopAction

// noopAction is an Action that does nothing. It is intended to be used for
// testing the test system itself.
type noopAction struct{}

func (noopAction) Banner() string                                 { return "[NO-OP]" }
func (noopAction) ConfigurePredicate(*PredicateOptions)           {}
func (noopAction) Apply(ctx context.Context, s ActionScope) error { return nil }
