package testkit_test

import (
	"testing"
	"time"

	. "github.com/dogmatiq/testkit"
)

func TestValidationScopeCreation(t *testing.T) {
	t.Run("CommandValidationScope", func(t *testing.T) {
		before := time.Now()
		s := CommandValidationScope()
		after := time.Now()

		if !s.IsNew() {
			t.Error("IsNew() should return true")
		}
		expectBetween(t, "ExecutedAt()", s.ExecutedAt(), before, after)
	})

	t.Run("EventValidationScope", func(t *testing.T) {
		before := time.Now()
		s := EventValidationScope()
		after := time.Now()

		if !s.IsNew() {
			t.Error("IsNew() should return true")
		}
		expectBetween(t, "RecordedAt()", s.RecordedAt(), before, after)
	})

	t.Run("DeadlineValidationScope", func(t *testing.T) {
		before := time.Now()
		s := DeadlineValidationScope()
		after := time.Now()

		if !s.IsNew() {
			t.Error("IsNew() should return true")
		}
		expectBetween(t, "ScheduledAt()", s.ScheduledAt(), before, after)
		expectBetween(t, "ScheduledFor()", s.ScheduledFor(), before, after)
	})
}

func expectBetween(t *testing.T, name string, got, before, after time.Time) {
	t.Helper()

	if got.Before(before) || got.After(after) {
		t.Errorf("unexpected %s: got %s, want a time between %s and %s", name, got, before, after)
	}
}

func TestValidationScopeOptions(t *testing.T) {
	at := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	scheduledFor := at.Add(time.Hour)

	t.Run("CommandValidationScope", func(t *testing.T) {
		s := CommandValidationScope(
			WithIsNew(false),
			WithExecutedAt(at),
		)

		if s.IsNew() {
			t.Error("IsNew() should return false")
		}
		if got := s.ExecutedAt(); !got.Equal(at) {
			t.Errorf("unexpected ExecutedAt(): got %s, want %s", got, at)
		}
	})

	t.Run("EventValidationScope", func(t *testing.T) {
		s := EventValidationScope(
			WithIsNew(false),
			WithRecordedAt(at),
		)

		if s.IsNew() {
			t.Error("IsNew() should return false")
		}
		if got := s.RecordedAt(); !got.Equal(at) {
			t.Errorf("unexpected RecordedAt(): got %s, want %s", got, at)
		}
	})

	t.Run("DeadlineValidationScope", func(t *testing.T) {
		s := DeadlineValidationScope(
			WithIsNew(false),
			WithScheduledAt(at),
			WithScheduledFor(scheduledFor),
		)

		if s.IsNew() {
			t.Error("IsNew() should return false")
		}
		if got := s.ScheduledAt(); !got.Equal(at) {
			t.Errorf("unexpected ScheduledAt(): got %s, want %s", got, at)
		}
		if got := s.ScheduledFor(); !got.Equal(scheduledFor) {
			t.Errorf("unexpected ScheduledFor(): got %s, want %s", got, scheduledFor)
		}
	})
}
