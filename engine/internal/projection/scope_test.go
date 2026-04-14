package projection_test

import (
	"context"
	"testing"
	"time"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/testkit/fact"
	"github.com/dogmatiq/testkit/x/xtesting"
)

func TestScopeRecordedAt(t *testing.T) {
	fx := newControllerTestFixture()

	fx.handler.HandleEventFunc = func(
		_ context.Context,
		s dogma.ProjectionEventScope,
		_ dogma.Event,
	) (uint64, error) {
		if !s.RecordedAt().Equal(fx.event.CreatedAt) {
			t.Fatalf("unexpected recorded time: %v", s.RecordedAt())
		}
		return s.Offset() + 1, nil
	}

	_, err := fx.ctrl.Handle(
		context.Background(),
		fact.Ignore,
		time.Now(),
		fx.event,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestScopeLog(t *testing.T) {
	fx := newControllerTestFixture()

	fx.handler.HandleEventFunc = func(
		_ context.Context,
		s dogma.ProjectionEventScope,
		_ dogma.Event,
	) (uint64, error) {
		s.Log("<format>", "<arg-1>", "<arg-2>")
		return s.Offset() + 1, nil
	}

	buf := &fact.Buffer{}
	_, err := fx.ctrl.Handle(
		context.Background(),
		buf,
		time.Now(),
		fx.event,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	xtesting.Expect(
		t,
		"unexpected logged facts",
		buf.Facts(),
		[]fact.Fact{
			fact.MessageLoggedByProjection{
				Handler:   fx.cfg,
				Envelope:  fx.event,
				LogFormat: "<format>",
				LogArguments: []any{
					"<arg-1>",
					"<arg-2>",
				},
			},
		},
	)
}

func TestScopeNow(t *testing.T) {
	fx := newControllerTestFixture()
	now := time.Now()

	fx.handler.CompactFunc = func(
		_ context.Context,
		s dogma.ProjectionCompactScope,
	) error {
		if got := s.Now(); got != now {
			t.Fatalf("unexpected current time: %v", got)
		}
		return nil
	}

	_, err := fx.ctrl.Tick(
		context.Background(),
		fact.Ignore,
		now,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
