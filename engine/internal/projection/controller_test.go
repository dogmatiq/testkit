package projection_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/enginekit/config"
	"github.com/dogmatiq/enginekit/config/runtimeconfig"
	. "github.com/dogmatiq/enginekit/enginetest/stubs"
	"github.com/dogmatiq/testkit/engine/internal/panicx"
	"github.com/dogmatiq/testkit/engine/internal/projection"
	"github.com/dogmatiq/testkit/envelope"
	"github.com/dogmatiq/testkit/fact"
	"github.com/dogmatiq/testkit/internal/test"
)

type controllerTestFixture struct {
	handler *ProjectionMessageHandlerStub
	cfg     *config.Projection
	ctrl    *projection.Controller
	event   *envelope.Envelope
}

func newControllerTestFixture() controllerTestFixture {
	event := envelope.NewEvent(
		"1000",
		EventA1,
		time.Now(),
	)

	handler := &ProjectionMessageHandlerStub{
		ConfigureFunc: func(c dogma.ProjectionConfigurer) {
			c.Identity("<name>", "fcbe8fe1-1085-497d-ba8e-09bedb031db2")
			c.Routes(
				dogma.HandlesEvent[*EventStub[TypeA]](),
			)
		},
	}

	cfg := runtimeconfig.FromProjection(handler)

	return controllerTestFixture{
		handler: handler,
		cfg:     cfg,
		ctrl: &projection.Controller{
			Config: cfg,
		},
		event: event,
	}
}

func TestControllerHandlerConfig(t *testing.T) {
	fx := newControllerTestFixture()

	if got := fx.ctrl.HandlerConfig(); got != fx.cfg {
		t.Fatalf("expected handler config %p, got %p", fx.cfg, got)
	}
}

func TestControllerTick(t *testing.T) {
	t.Run("it does not return any envelopes", func(t *testing.T) {
		fx := newControllerTestFixture()

		envelopes, err := fx.ctrl.Tick(
			context.Background(),
			fact.Ignore,
			time.Now(),
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(envelopes) != 0 {
			t.Fatalf("expected no envelopes, got %d", len(envelopes))
		}
	})

	t.Run("it performs projection compaction", func(t *testing.T) {
		fx := newControllerTestFixture()
		expected := errors.New("<error>")

		fx.handler.CompactFunc = func(
			context.Context,
			dogma.ProjectionCompactScope,
		) error {
			return expected
		}

		buf := &fact.Buffer{}
		_, err := fx.ctrl.Tick(
			context.Background(),
			buf,
			time.Now(),
		)

		test.Expect(t, "unexpected tick error", err, expected)
		test.Expect(
			t,
			"unexpected compaction facts",
			buf.Facts(),
			[]fact.Fact{
				fact.ProjectionCompactionBegun{Handler: fx.cfg},
				fact.ProjectionCompactionCompleted{
					Handler: fx.cfg,
					Error:   expected,
				},
			},
		)
	})

	t.Run("it does not compact again until CompactInterval has elapsed", func(t *testing.T) {
		fx := newControllerTestFixture()

		fx.handler.CompactFunc = func(
			context.Context,
			dogma.ProjectionCompactScope,
		) error {
			return errors.New("<called>")
		}

		start := time.Now()

		_, err := fx.ctrl.Tick(
			context.Background(),
			fact.Ignore,
			start,
		)
		if err == nil || err.Error() != "<called>" {
			t.Fatalf("expected compaction error, got %v", err)
		}

		_, err = fx.ctrl.Tick(
			context.Background(),
			fact.Ignore,
			start.Add(projection.CompactInterval-1),
		)
		if err != nil {
			t.Fatalf("expected no compaction before interval, got %v", err)
		}

		_, err = fx.ctrl.Tick(
			context.Background(),
			fact.Ignore,
			start.Add(projection.CompactInterval),
		)
		if err == nil || err.Error() != "<called>" {
			t.Fatalf("expected compaction error at interval, got %v", err)
		}
	})
}

func TestControllerHandle(t *testing.T) {
	t.Run("it forwards the message to the handler", func(t *testing.T) {
		fx := newControllerTestFixture()
		called := false

		fx.handler.HandleEventFunc = func(
			_ context.Context,
			s dogma.ProjectionEventScope,
			m dogma.Event,
		) (uint64, error) {
			called = true
			test.Expect(t, "unexpected event", m, EventA1)
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
		if !called {
			t.Fatal("expected handler to be called")
		}
	})

	t.Run("it propagates handler errors", func(t *testing.T) {
		fx := newControllerTestFixture()
		expected := errors.New("<error>")

		fx.handler.HandleEventFunc = func(
			context.Context,
			dogma.ProjectionEventScope,
			dogma.Event,
		) (uint64, error) {
			return 0, expected
		}

		_, err := fx.ctrl.Handle(
			context.Background(),
			fact.Ignore,
			time.Now(),
			fx.event,
		)

		test.Expect(t, "unexpected handler error", err, expected)
	})

	t.Run("it propagates errors when loading the checkpoint offset", func(t *testing.T) {
		fx := newControllerTestFixture()
		expected := errors.New("<error>")

		fx.handler.CheckpointOffsetFunc = func(
			context.Context,
			string,
		) (uint64, error) {
			return 0, expected
		}

		_, err := fx.ctrl.Handle(
			context.Background(),
			fact.Ignore,
			time.Now(),
			fx.event,
		)

		test.Expect(t, "unexpected checkpoint error", err, expected)
	})

	t.Run("it passes the correct stream offsets", func(t *testing.T) {
		fx := newControllerTestFixture()
		var checkpoint uint64

		fx.handler.CheckpointOffsetFunc = func(
			_ context.Context,
			streamID string,
		) (uint64, error) {
			if streamID != fx.event.EventStreamID {
				t.Fatalf("unexpected stream ID: %s", streamID)
			}
			return checkpoint, nil
		}

		fx.handler.HandleEventFunc = func(
			_ context.Context,
			s dogma.ProjectionEventScope,
			_ dogma.Event,
		) (uint64, error) {
			if s.StreamID() != fx.event.EventStreamID {
				t.Fatalf("unexpected scope stream ID: %s", s.StreamID())
			}
			if s.Offset() != fx.event.EventStreamOffset {
				t.Fatalf("unexpected scope offset: %d", s.Offset())
			}
			if s.CheckpointOffset() != checkpoint {
				t.Fatalf("unexpected checkpoint offset: %d", s.CheckpointOffset())
			}

			checkpoint = s.Offset() + 1
			return checkpoint, nil
		}

		_, err := fx.ctrl.Handle(
			context.Background(),
			fact.Ignore,
			time.Now(),
			fx.event,
		)
		if err != nil {
			t.Fatalf("unexpected error on first handle: %v", err)
		}

		fx.event.EventStreamOffset++

		_, err = fx.ctrl.Handle(
			context.Background(),
			fact.Ignore,
			time.Now(),
			fx.event,
		)
		if err != nil {
			t.Fatalf("unexpected error on second handle: %v", err)
		}
	})

	t.Run("it does not handle events that have already been applied", func(t *testing.T) {
		fx := newControllerTestFixture()

		fx.handler.CheckpointOffsetFunc = func(
			_ context.Context,
			_ string,
		) (uint64, error) {
			return 1, nil
		}

		fx.handler.HandleEventFunc = func(
			_ context.Context,
			_ dogma.ProjectionEventScope,
			_ dogma.Event,
		) (uint64, error) {
			t.Fatal("unexpected call to HandleEvent()")
			return 0, nil
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
	})

	t.Run("it returns an error if there is an optimistic concurrency conflict", func(t *testing.T) {
		fx := newControllerTestFixture()

		fx.handler.HandleEventFunc = func(
			_ context.Context,
			s dogma.ProjectionEventScope,
			_ dogma.Event,
		) (uint64, error) {
			return 123, nil
		}

		_, err := fx.ctrl.Handle(
			context.Background(),
			fact.Ignore,
			time.Now(),
			fx.event,
		)

		const expected = "optimistic concurrency conflict when handling event at offset 0 of stream ea2763d3-11d4-5f81-aa9f-e666d13bff4f: expected checkpoint offset of 1, handler returned 123"
		if err == nil || err.Error() != expected {
			t.Fatalf("expected concurrency conflict %q, got %v", expected, err)
		}
	})

	t.Run("it provides more context to UnexpectedMessage panics from HandleEvent()", func(t *testing.T) {
		fx := newControllerTestFixture()

		fx.handler.HandleEventFunc = func(
			context.Context,
			dogma.ProjectionEventScope,
			dogma.Event,
		) (uint64, error) {
			panic(dogma.UnexpectedMessage)
		}

		defer func() {
			r := recover()
			x, ok := r.(panicx.UnexpectedMessage)
			if !ok {
				t.Fatalf("expected panicx.UnexpectedMessage panic, got %T", r)
			}

			if x.Handler != fx.cfg {
				t.Fatalf("unexpected handler: %v", x.Handler)
			}
			if x.Interface != "ProjectionMessageHandler" {
				t.Fatalf("unexpected interface: %s", x.Interface)
			}
			if x.Method != "HandleEvent" {
				t.Fatalf("unexpected method: %s", x.Method)
			}
			test.Expect(t, "unexpected panic message", x.Message, fx.event.Message)
		}()

		fx.ctrl.Handle(
			context.Background(),
			fact.Ignore,
			time.Now(),
			fx.event,
		)
	})

	t.Run("when compact-during-handling is disabled, it does not perform compaction", func(t *testing.T) {
		fx := newControllerTestFixture()

		fx.handler.CompactFunc = func(
			context.Context,
			dogma.ProjectionCompactScope,
		) error {
			return errors.New("<error>")
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

		test.Expect(t, "unexpected facts", buf.Facts(), []fact.Fact(nil))
	})

	t.Run("when compact-during-handling is enabled, it performs projection compaction", func(t *testing.T) {
		fx := newControllerTestFixture()
		fx.ctrl.CompactDuringHandling = true
		expected := errors.New("<error>")

		fx.handler.CompactFunc = func(
			context.Context,
			dogma.ProjectionCompactScope,
		) error {
			return expected
		}

		buf := &fact.Buffer{}
		_, err := fx.ctrl.Handle(
			context.Background(),
			buf,
			time.Now(),
			fx.event,
		)

		test.Expect(t, "unexpected handle error", err, expected)
		test.Expect(
			t,
			"unexpected compaction facts",
			buf.Facts(),
			[]fact.Fact{
				fact.ProjectionCompactionBegun{Handler: fx.cfg},
				fact.ProjectionCompactionCompleted{
					Handler: fx.cfg,
					Error:   expected,
				},
			},
		)
	})
}

func TestControllerReset(t *testing.T) {
	fx := newControllerTestFixture()
	fx.ctrl.Reset()
}
