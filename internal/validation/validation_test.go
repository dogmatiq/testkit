package validation_test

import (
	"testing"
	"time"

	. "github.com/dogmatiq/testkit/internal/validation"
)

func TestScopes(t *testing.T) {
	at := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	scheduledFor := at.Add(time.Hour)

	t.Run("CommandValidationScope", func(t *testing.T) {
		s := CommandValidationScope(true, at)

		if !s.IsNew() {
			t.Error("IsNew() should return true")
		}
		if got := s.ExecutedAt(); !got.Equal(at) {
			t.Errorf("unexpected ExecutedAt(): got %s, want %s", got, at)
		}
	})

	t.Run("EventValidationScope", func(t *testing.T) {
		s := EventValidationScope(true, at)

		if !s.IsNew() {
			t.Error("IsNew() should return true")
		}
		if got := s.RecordedAt(); !got.Equal(at) {
			t.Errorf("unexpected RecordedAt(): got %s, want %s", got, at)
		}
	})

	t.Run("DeadlineValidationScope", func(t *testing.T) {
		s := DeadlineValidationScope(true, at, scheduledFor)

		if !s.IsNew() {
			t.Error("IsNew() should return true")
		}
		if got := s.ScheduledAt(); !got.Equal(at) {
			t.Errorf("unexpected ScheduledAt(): got %s, want %s", got, at)
		}
		if got := s.ScheduledFor(); !got.Equal(scheduledFor) {
			t.Errorf("unexpected ScheduledFor(): got %s, want %s", got, scheduledFor)
		}
	})

	t.Run("IsNew", func(t *testing.T) {
		for _, isNew := range []bool{true, false} {
			if got := CommandValidationScope(isNew, at).IsNew(); got != isNew {
				t.Errorf("CommandValidationScope(%t).IsNew() returned %t", isNew, got)
			}
			if got := EventValidationScope(isNew, at).IsNew(); got != isNew {
				t.Errorf("EventValidationScope(%t).IsNew() returned %t", isNew, got)
			}
			if got := DeadlineValidationScope(isNew, at, at).IsNew(); got != isNew {
				t.Errorf("DeadlineValidationScope(%t).IsNew() returned %t", isNew, got)
			}
		}
	})
}
