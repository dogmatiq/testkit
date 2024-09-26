package testkit

import (
	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/testkit/internal/validation"
)

// CommandValidationScope returns a [dogma.CommandValidationScope] that can be
// used when testing command validation logic.
func CommandValidationScope(...CommandValidationScopeOption) dogma.CommandValidationScope {
	return validation.CommandValidationScope()
}

// EventValidationScope returns a [dogma.EventValidationScope] that can be used
// when testing event validation logic.
func EventValidationScope(...EventValidationScopeOption) dogma.EventValidationScope {
	return validation.EventValidationScope()
}

// TimeoutValidationScope returns a [dogma.TimeoutValidationScope] that can be
// used when testing timeout validation logic.
func TimeoutValidationScope(...TimeoutValidationScopeOption) dogma.TimeoutValidationScope {
	return validation.TimeoutValidationScope()
}

// CommandValidationScopeOption is an option that changes the behavior of a
// [dogma.CommandValidationScope] created by calling [CommandValidationScope].
type CommandValidationScopeOption interface {
	reservedCommandValidationScopeOption()
}

// EventValidationScopeOption is an option that changes the behavior of a
// [dogma.EventValidationScope] created by calling [EventValidationScope].
type EventValidationScopeOption interface {
	reservedEventValidationScopeOption()
}

// TimeoutValidationScopeOption is an option that changes the behavior of a
// [dogma.TimeoutValidationScope] created by calling [TimeoutValidationScope].
type TimeoutValidationScopeOption interface {
	reservedTimeoutValidationScopeOption()
}
