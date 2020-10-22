package testkit

import (
	"fmt"
	"time"
)

// TimeAdvancer is a function that determines the new time that the engine is
// advanced to during a call to Test.AdvanceTime().
type TimeAdvancer func(before time.Time) (after time.Time, description string)

// ToTime returns a TimeAdvance that advances the engine time to a specific time.
func ToTime(t time.Time) TimeAdvancer {
	return func(time.Time) (time.Time, string) {
		return t, fmt.Sprintf("ADVANCING TIME TO %s", t.Format(time.RFC3339))
	}
}

// ByDuration returns a TimeAdvance that advances the engine time by a fixed duration.
func ByDuration(d time.Duration) TimeAdvancer {
	return func(now time.Time) (time.Time, string) {
		return now.Add(d), fmt.Sprintf("ADVANCING TIME BY %s", d)
	}
}
