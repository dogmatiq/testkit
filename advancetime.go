package testkit

import (
	"fmt"
	"time"

	"github.com/dogmatiq/testkit/assert"
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

func (act advanceTime) Apply(
	t *Test,
	e Expectation,
	options []ExpectOption,
) {
	now := act.adj.Step(t.now)

	if now.Before(t.now) {
		panic(fmt.Sprintf(
			"changing the clock %s results in a time earlier than the current time",
			act.adj.Description(),
		))
	}

	t.now = now

	t.begin(assert.AdvanceTimeOperation, e)
	t.tick(options, e)
	t.end(e)
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
