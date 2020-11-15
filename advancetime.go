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

func (act advanceTime) Heading() string {
	return fmt.Sprintf(
		"ADVANCING TIME (%s)",
		act.adj.Description(),
	)
}

// ExpectOptions returns the options to use by default when this action is
// used with Test.Expect().
func (act advanceTime) ExpectOptions() []ExpectOption {
	return nil
}

func (act advanceTime) Apply(
	ctx context.Context,
	t *Test,
	options []engine.OperationOption,
) error {
	now := act.adj.Step(t.now)

	if now.Before(t.now) {
		return fmt.Errorf(
			"adjusting the clock %s would reverse time",
			act.adj.Description(),
		)
	}

	t.now = now

	// There is already an engine.WithCurrentTime() based on t.now in the
	// options slice. Because we have just updated t.now we need to override it
	// for this one engine tick.
	options = append(options, engine.WithCurrentTime(now))

	return t.engine.Tick(ctx, options...)
}

// toTime is a ClockMutation that advances the clock to a specific time.
type toTime time.Time

func (t toTime) Description() string {
	return fmt.Sprintf(
		"to %s",
		time.Time(t).Format(time.RFC3339),
	)
}

func (t toTime) Step(time.Time) time.Time {
	return time.Time(t)
}

// ByDuration is a ClockMutation that advances the clock by a fixed duration.
type byDuration time.Duration

func (d byDuration) Description() string {
	return fmt.Sprintf(
		"by %s",
		time.Duration(d),
	)
}

func (d byDuration) Step(before time.Time) time.Time {
	return before.Add(time.Duration(d))
}