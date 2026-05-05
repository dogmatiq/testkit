package validation

import "github.com/dogmatiq/dogma"

// CommandValidationScope returns the validation scope for command messages.
func CommandValidationScope() dogma.CommandValidationScope {
	return struct{ dogma.CommandValidationScope }{}
}

// EventValidationScope returns the validation scope for event messages.
func EventValidationScope() dogma.EventValidationScope {
	return struct{ dogma.EventValidationScope }{}
}

// DeadlineValidationScope returns the validation scope for deadline messages.
func DeadlineValidationScope() dogma.DeadlineValidationScope {
	return struct{ dogma.DeadlineValidationScope }{}
}
