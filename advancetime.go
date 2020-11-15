package testkit

import (
	"context"
	"fmt"
	"time"

	"github.com/dogmatiq/testkit/engine"
)

// AdvanceTime returns an Action that simulates the passage of time by advancing
// the test's virtual clock.
//
// This allows testing of application logic that depends on time, such as
// processes that use timeout messages and projections that use the "recorded
// at" time of events.
//
// It accepts a TimeAdjustment which calculates the amount of time that the
// clock is advanced.
//
// There are two built-in adjustment types; ToTime() and ByDuration(). Users may
// provide their own TimeAdjustment implementations that model time-related
// concepts within the application's business domain.
func AdvanceTime(adj TimeAdjustment) Action {
	return advanceTime{adj}
}

// A TimeAdjustment describes a change to the test's virtual clock.
type TimeAdjustment interface {
	// Description returns a human-readable string that describes how the clock
	// will be advanced.
	//
	// It should complete the sentence "The clock is being advanced...". For
	// example, "by 10 seconds".
	Description() string

	// Step returns the time that the virtual clock should be set to as a result
	// of the adjustment.
	//
	// t is the virtual clock's current time.
	Step(t time.Time) time.Time
}

// ToTime returns a TimeAdjustment that advances the virtual clock to a specific
// time.
func ToTime(t time.Time) TimeAdjustment {
	return toTime(t)
}

// ByDuration returns a TimeAdjustment that advances the virtual clock by a
// fixed duration.
func ByDuration(d time.Duration) TimeAdjustment {
	return byDuration(d)
}

// advanceTime is an implementation of Action that advances the virtual clock.
type advanceTime struct {
	adj TimeAdjustment
}

// Heading returns a human-readable description of the action, used as a
// heading within the test report.
//
// Any engine activity as a result of this action is logged beneath this
// heading.
func (a advanceTime) Heading() string {
	return fmt.Sprintf(
		"ADVANCING TIME (%s)",
		a.adj.Description(),
	)
}

// ExpectOptions returns the options to use by default when this action is
// used with Test.Expect().
func (a advanceTime) ExpectOptions() []ExpectOption {
	return nil
}

// Apply performs the action within the context of a specific test.
func (a advanceTime) Apply(ctx context.Context, s ActionScope) error {
	now := a.adj.Step(s.Test.now)

	if now.Before(s.Test.now) {
		return fmt.Errorf(
			"adjusting the clock %s would reverse time",
			a.adj.Description(),
		)
	}

	s.Test.now = now

	// There is already an engine.WithCurrentTime() based on t.now in the
	// options slice. Because we have just updated s.Test.now we need to
	// override it for this one engine tick.
	s.OperationOptions = append(
		s.OperationOptions,
		engine.WithCurrentTime(now),
	)

	return s.Engine.Tick(ctx, s.OperationOptions...)
}

// toTime is a ClockMutation that advances the clock to a specific time.
type toTime time.Time

// Description returns a human-readable string that describes how the clock
// will be advanced.
//
// It should complete the sentence "The clock is being advanced...". For
// example, "by 10 seconds".
func (t toTime) Description() string {
	return fmt.Sprintf(
		"to %s",
		time.Time(t).Format(time.RFC3339),
	)
}

// Step returns the time that the virtual clock should be set to as a result
// of the adjustment.
//
// t is the virtual clock's current time.
func (t toTime) Step(time.Time) time.Time {
	return time.Time(t)
}

// ByDuration is a ClockMutation that advances the clock by a fixed duration.
type byDuration time.Duration

// Description returns a human-readable string that describes how the clock
// will be advanced.
//
// It should complete the sentence "The clock is being advanced...". For
// example, "by 10 seconds".
func (d byDuration) Description() string {
	return fmt.Sprintf(
		"by %s",
		time.Duration(d),
	)
}

// Step returns the time that the virtual clock should be set to as a result
// of the adjustment.
//
// t is the virtual clock's current time.
func (d byDuration) Step(before time.Time) time.Time {
	return before.Add(time.Duration(d))
}
