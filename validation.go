package testkit

import (
	"time"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/testkit/internal/validation"
)

// CommandValidationScope returns a [dogma.CommandValidationScope] that can be
// used when testing command validation logic.
//
// By default, the scope describes a new command that was executed at the
// current time.
func CommandValidationScope(options ...CommandValidationScopeOption) dogma.CommandValidationScope {
	o := commandValidationScopeOptions{
		isNew:      true,
		executedAt: time.Now(),
	}

	for _, opt := range options {
		opt.applyToCommandValidationScope(&o)
	}

	return validation.CommandValidationScope(o.isNew, o.executedAt)
}

// EventValidationScope returns a [dogma.EventValidationScope] that can be used
// when testing event validation logic.
//
// By default, the scope describes a new event that was recorded at the current
// time.
func EventValidationScope(options ...EventValidationScopeOption) dogma.EventValidationScope {
	o := eventValidationScopeOptions{
		isNew:      true,
		recordedAt: time.Now(),
	}

	for _, opt := range options {
		opt.applyToEventValidationScope(&o)
	}

	return validation.EventValidationScope(o.isNew, o.recordedAt)
}

// DeadlineValidationScope returns a [dogma.DeadlineValidationScope] that can be
// used when testing deadline validation logic.
//
// By default, the scope describes a new deadline that was scheduled at the
// current time, to be delivered at the current time.
func DeadlineValidationScope(options ...DeadlineValidationScopeOption) dogma.DeadlineValidationScope {
	now := time.Now()
	o := deadlineValidationScopeOptions{
		isNew:        true,
		scheduledAt:  now,
		scheduledFor: now,
	}

	for _, opt := range options {
		opt.applyToDeadlineValidationScope(&o)
	}

	return validation.DeadlineValidationScope(o.isNew, o.scheduledAt, o.scheduledFor)
}

// CommandValidationScopeOption is an option that changes the behavior of a
// [dogma.CommandValidationScope] created by calling [CommandValidationScope].
type CommandValidationScopeOption interface {
	applyToCommandValidationScope(*commandValidationScopeOptions)
}

// EventValidationScopeOption is an option that changes the behavior of a
// [dogma.EventValidationScope] created by calling [EventValidationScope].
type EventValidationScopeOption interface {
	applyToEventValidationScope(*eventValidationScopeOptions)
}

// DeadlineValidationScopeOption is an option that changes the behavior of a
// [dogma.DeadlineValidationScope] created by calling [DeadlineValidationScope].
type DeadlineValidationScopeOption interface {
	applyToDeadlineValidationScope(*deadlineValidationScopeOptions)
}

// ValidationScopeOption is an option that can be used with any of
// [CommandValidationScope], [EventValidationScope] and [DeadlineValidationScope].
type ValidationScopeOption interface {
	CommandValidationScopeOption
	EventValidationScopeOption
	DeadlineValidationScopeOption
}

// WithIsNew returns an option that sets the value returned by the scope's
// IsNew() method.
//
// IsNew() returns true by default.
func WithIsNew(isNew bool) ValidationScopeOption {
	return isNewOption(isNew)
}

// WithExecutedAt returns an option that sets the value returned by a command
// validation scope's ExecutedAt() method.
func WithExecutedAt(t time.Time) CommandValidationScopeOption {
	return commandValidationScopeOptionFunc(func(o *commandValidationScopeOptions) {
		o.executedAt = t
	})
}

// WithRecordedAt returns an option that sets the value returned by an event
// validation scope's RecordedAt() method.
func WithRecordedAt(t time.Time) EventValidationScopeOption {
	return eventValidationScopeOptionFunc(func(o *eventValidationScopeOptions) {
		o.recordedAt = t
	})
}

// WithScheduledAt returns an option that sets the value returned by a deadline
// validation scope's ScheduledAt() method.
func WithScheduledAt(t time.Time) DeadlineValidationScopeOption {
	return deadlineValidationScopeOptionFunc(func(o *deadlineValidationScopeOptions) {
		o.scheduledAt = t
	})
}

// WithScheduledFor returns an option that sets the value returned by a deadline
// validation scope's ScheduledFor() method.
func WithScheduledFor(t time.Time) DeadlineValidationScopeOption {
	return deadlineValidationScopeOptionFunc(func(o *deadlineValidationScopeOptions) {
		o.scheduledFor = t
	})
}

type (
	commandValidationScopeOptions struct {
		isNew      bool
		executedAt time.Time
	}

	eventValidationScopeOptions struct {
		isNew      bool
		recordedAt time.Time
	}

	deadlineValidationScopeOptions struct {
		isNew        bool
		scheduledAt  time.Time
		scheduledFor time.Time
	}

	commandValidationScopeOptionFunc  func(*commandValidationScopeOptions)
	eventValidationScopeOptionFunc    func(*eventValidationScopeOptions)
	deadlineValidationScopeOptionFunc func(*deadlineValidationScopeOptions)

	// isNewOption is the implementation of the option returned by WithIsNew().
	isNewOption bool
)

func (f commandValidationScopeOptionFunc) applyToCommandValidationScope(o *commandValidationScopeOptions) {
	f(o)
}

func (f eventValidationScopeOptionFunc) applyToEventValidationScope(o *eventValidationScopeOptions) {
	f(o)
}

func (f deadlineValidationScopeOptionFunc) applyToDeadlineValidationScope(o *deadlineValidationScopeOptions) {
	f(o)
}

func (o isNewOption) applyToCommandValidationScope(opts *commandValidationScopeOptions) {
	opts.isNew = bool(o)
}

func (o isNewOption) applyToEventValidationScope(opts *eventValidationScopeOptions) {
	opts.isNew = bool(o)
}

func (o isNewOption) applyToDeadlineValidationScope(opts *deadlineValidationScopeOptions) {
	opts.isNew = bool(o)
}
