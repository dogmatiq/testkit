package testkit

import (
	"context"
	"fmt"
	"time"

	"github.com/dogmatiq/testkit/engine"
	"github.com/dogmatiq/testkit/location"
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
	if adj == nil {
		panic("AdvanceTime(<nil>): adjustment must not be nil")
	}

	return advanceTimeAction{
		adj,
		location.OfCall(),
	}
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
	if d < 0 {
		panic(fmt.Sprintf("ByDuration(%s): duration must not be negative", d))
	}

	return byDuration(d)
}

// advanceTimeAction is an implementation of Action that advances the virtual
// clock.
type advanceTimeAction struct {
	adj TimeAdjustment
	loc location.Location
}

func (a advanceTimeAction) Caption() string {
	return fmt.Sprintf(
		"advancing time %s",
		a.adj.Description(),
	)
}

func (a advanceTimeAction) Location() location.Location {
	return a.loc
}

func (a advanceTimeAction) ConfigurePredicate(*PredicateOptions) {
}

func (a advanceTimeAction) Do(ctx context.Context, s ActionScope) error {
	now := a.adj.Step(*s.VirtualClock)

	if now.Before(*s.VirtualClock) {
		return fmt.Errorf(
			"adjusting the clock %s would reverse time",
			a.adj.Description(),
		)
	}

	*s.VirtualClock = now

	// There is already an engine.WithCurrentTime() based on the virtual clock
	// in options slice. Because we have just updated the clock we need to
	// override it for this one engine tick.
	s.OperationOptions = append(
		s.OperationOptions,
		engine.WithCurrentTime(now),
	)

	return s.Engine.Tick(ctx, s.OperationOptions...)
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
