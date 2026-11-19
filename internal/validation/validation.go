package validation

import (
	"time"

	"github.com/dogmatiq/dogma"
)

// CommandValidationScope returns the validation scope for a command message
// that was executed at executedAt.
//
// isNew is the value that the scope's IsNew() method returns.
func CommandValidationScope(isNew bool, executedAt time.Time) dogma.CommandValidationScope {
	return commandScope{messageScope{isNew}, executedAt}
}

// EventValidationScope returns the validation scope for an event message that
// was recorded at recordedAt.
//
// isNew is the value that the scope's IsNew() method returns.
func EventValidationScope(isNew bool, recordedAt time.Time) dogma.EventValidationScope {
	return eventScope{messageScope{isNew}, recordedAt}
}

// DeadlineValidationScope returns the validation scope for a deadline message
// that was scheduled at scheduledAt, to be delivered at scheduledFor.
//
// isNew is the value that the scope's IsNew() method returns.
func DeadlineValidationScope(isNew bool, scheduledAt, scheduledFor time.Time) dogma.DeadlineValidationScope {
	return deadlineScope{messageScope{isNew}, scheduledAt, scheduledFor}
}

// messageScope is embedded by each scope type to implement
// [dogma.MessageValidationScope].
type messageScope struct {
	isNew bool
}

func (s messageScope) IsNew() bool { return s.isNew }

type commandScope struct {
	messageScope
	executedAt time.Time
}

func (s commandScope) ExecutedAt() time.Time { return s.executedAt }

type eventScope struct {
	messageScope
	recordedAt time.Time
}

func (s eventScope) RecordedAt() time.Time { return s.recordedAt }

type deadlineScope struct {
	messageScope
	scheduledAt  time.Time
	scheduledFor time.Time
}

func (s deadlineScope) ScheduledAt() time.Time  { return s.scheduledAt }
func (s deadlineScope) ScheduledFor() time.Time { return s.scheduledFor }
