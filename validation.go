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

// DeadlineValidationScope returns a [dogma.DeadlineValidationScope] that can be
// used when testing deadline validation logic.
func DeadlineValidationScope(...DeadlineValidationScopeOption) dogma.DeadlineValidationScope {
	return validation.DeadlineValidationScope()
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

// DeadlineValidationScopeOption is an option that changes the behavior of a
// [dogma.DeadlineValidationScope] created by calling [DeadlineValidationScope].
type DeadlineValidationScopeOption interface {
	reservedDeadlineValidationScopeOption()
}
