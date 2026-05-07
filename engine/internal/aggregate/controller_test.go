package aggregate_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/enginekit/config"
	"github.com/dogmatiq/enginekit/config/runtimeconfig"
	"github.com/dogmatiq/enginekit/enginetest/stubs"
	"github.com/dogmatiq/testkit/engine/internal/aggregate"
	"github.com/dogmatiq/testkit/engine/internal/panicx"
	"github.com/dogmatiq/testkit/envelope"
	"github.com/dogmatiq/testkit/fact"
	"github.com/dogmatiq/testkit/internal/x/xtesting"
	"github.com/dogmatiq/testkit/location"
)

func TestControllerHandlerConfig(t *testing.T) {
	f := newControllerTestFixture()

	if got := f.ctrl.HandlerConfig(); got != f.cfg {
		t.Fatalf("unexpected handler config: got %p, want %p", got, f.cfg)
	}
}

func TestControllerTick(t *testing.T) {
	t.Run("returns no envelopes", func(t *testing.T) {
		f := newControllerTestFixture()

		envelopes, err := f.ctrl.Tick(
			context.Background(),
			fact.Ignore,
			time.Now(),
		)
		if err != nil {
			t.Fatal(err)
		}

		xtesting.Expect(t, "unexpected envelopes", envelopes, []*envelope.Envelope(nil))
	})

	t.Run("records no facts", func(t *testing.T) {
		f := newControllerTestFixture()
		buf := &fact.Buffer{}

		_, err := f.ctrl.Tick(
			context.Background(),
			buf,
			time.Now(),
		)
		if err != nil {
			t.Fatal(err)
		}

		xtesting.Expect(t, "unexpected facts", buf.Facts(), []fact.Fact(nil))
	})
}

func TestControllerHandle(t *testing.T) {
	t.Run("forwards the message to the handler", func(t *testing.T) {
		f := newControllerTestFixture()
		called := false

		f.handler.HandleCommandFunc = func(
			_ *stubs.AggregateRootStub,
			_ dogma.AggregateCommandScope[*stubs.AggregateRootStub],
			m dogma.Command,
		) {
			called = true
			xtesting.Expect(t, "unexpected message", m, stubs.CommandA1)
		}

		_, err := f.ctrl.Handle(
			context.Background(),
			fact.Ignore,
			time.Now(),
			f.command,
		)
		if err != nil {
			t.Fatal(err)
		}

		xtesting.Expect(t, "expected handler to be called", called, true)
	})

	t.Run("returns the recorded events", func(t *testing.T) {
		f := newControllerTestFixture()

		f.handler.HandleCommandFunc = func(
			_ *stubs.AggregateRootStub,
			s dogma.AggregateCommandScope[*stubs.AggregateRootStub],
			_ dogma.Command,
		) {
			s.RecordEvent(stubs.EventA1)
			s.RecordEvent(stubs.EventA2)
		}

		now := time.Now()
		events, err := f.ctrl.Handle(
			context.Background(),
			fact.Ignore,
			now,
			f.command,
		)
		if err != nil {
			t.Fatal(err)
		}

		xtesting.Expect(
			t,
			"unexpected events",
			events,
			[]*envelope.Envelope{
				f.command.NewEvent(
					"1",
					stubs.EventA1,
					now,
					envelope.Origin{
						Handler:     f.cfg,
						HandlerType: config.AggregateHandlerType,
						InstanceID:  "<instance-A1>",
					},
					"78e27a08-0ae8-52cf-8f46-79e448ed5bf6",
					0,
				),
				f.command.NewEvent(
					"2",
					stubs.EventA2,
					now,
					envelope.Origin{
						Handler:     f.cfg,
						HandlerType: config.AggregateHandlerType,
						InstanceID:  "<instance-A1>",
					},
					"78e27a08-0ae8-52cf-8f46-79e448ed5bf6",
					1,
				),
			},
		)
	})

	t.Run("panics when the handler routes to an empty instance ID", func(t *testing.T) {
		f := newControllerTestFixture()
		f.handler.RouteCommandToInstanceFunc = func(dogma.Command) string {
			return ""
		}

		x := mustPanicUnexpectedBehavior(t, func() {
			_, _ = f.ctrl.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				f.command,
			)
		})

		xtesting.Expect(t, "unexpected handler", x.Handler, f.cfg)
		xtesting.Expect(t, "unexpected interface", x.Interface, "AggregateMessageHandler")
		xtesting.Expect(t, "unexpected method", x.Method, "RouteCommandToInstance")
		xtesting.Expect(t, "unexpected implementation", x.Implementation, f.cfg.Implementation())
		xtesting.Expect(t, "unexpected message", x.Message, f.command.Message)
		xtesting.Expect(
			t,
			"unexpected description",
			x.Description,
			"routed a command of type *stubs.CommandStub[TypeA] to an empty ID",
		)
		expectLocation(t, x.Location, "/stubs/aggregate.go")
	})

	t.Run("records AggregateInstanceNotFound when the instance does not exist", func(t *testing.T) {
		f := newControllerTestFixture()
		buf := &fact.Buffer{}

		_, err := f.ctrl.Handle(
			context.Background(),
			buf,
			time.Now(),
			f.command,
		)
		if err != nil {
			t.Fatal(err)
		}

		got, ok := findFact[fact.AggregateInstanceNotFound](buf.Facts())
		if !ok {
			t.Fatal("expected AggregateInstanceNotFound fact")
		}

		xtesting.Expect(
			t,
			"unexpected fact",
			got,
			fact.AggregateInstanceNotFound{
				Handler:    f.cfg,
				InstanceID: "<instance-A1>",
				Envelope:   f.command,
			},
		)
	})

	t.Run("passes a new aggregate root when the instance does not exist", func(t *testing.T) {
		f := newControllerTestFixture()
		called := false

		f.handler.HandleCommandFunc = func(
			r *stubs.AggregateRootStub,
			_ dogma.AggregateCommandScope[*stubs.AggregateRootStub],
			_ dogma.Command,
		) {
			called = true
			xtesting.Expect(t, "unexpected aggregate root", r, &stubs.AggregateRootStub{})
		}

		_, err := f.ctrl.Handle(
			context.Background(),
			fact.Ignore,
			time.Now(),
			f.command,
		)
		if err != nil {
			t.Fatal(err)
		}

		xtesting.Expect(t, "expected handler to be called", called, true)
	})

	t.Run("panics if New returns nil when the instance does not exist", func(t *testing.T) {
		f := newControllerTestFixture()
		f.handler.NewFunc = func() *stubs.AggregateRootStub {
			return nil
		}

		x := mustPanicUnexpectedBehavior(t, func() {
			_, _ = f.ctrl.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				f.command,
			)
		})

		xtesting.Expect(t, "unexpected handler", x.Handler, f.cfg)
		xtesting.Expect(t, "unexpected interface", x.Interface, "AggregateMessageHandler")
		xtesting.Expect(t, "unexpected method", x.Method, "New")
		xtesting.Expect(t, "unexpected implementation", x.Implementation, f.cfg.Implementation())
		xtesting.Expect(t, "unexpected message", x.Message, f.command.Message)
		xtesting.Expect(t, "unexpected description", x.Description, "returned a nil aggregate root")
		expectLocation(t, x.Location, "/stubs/aggregate.go")
	})

	t.Run("panics if New returns nil when the instance exists", func(t *testing.T) {
		f := newControllerTestFixture()
		seedControllerInstance(t, f)

		f.handler.NewFunc = func() *stubs.AggregateRootStub {
			return nil
		}

		x := mustPanicUnexpectedBehavior(t, func() {
			_, _ = f.ctrl.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				f.command,
			)
		})

		xtesting.Expect(t, "unexpected handler", x.Handler, f.cfg)
		xtesting.Expect(t, "unexpected interface", x.Interface, "AggregateMessageHandler")
		xtesting.Expect(t, "unexpected method", x.Method, "New")
		xtesting.Expect(t, "unexpected implementation", x.Implementation, f.cfg.Implementation())
		xtesting.Expect(t, "unexpected message", x.Message, f.command.Message)
		xtesting.Expect(t, "unexpected description", x.Description, "returned a nil aggregate root")
		expectLocation(t, x.Location, "/stubs/aggregate.go")
	})

	t.Run("records AggregateInstanceLoaded when the instance exists", func(t *testing.T) {
		f := newControllerTestFixture()
		seedControllerInstance(t, f)

		buf := &fact.Buffer{}
		_, err := f.ctrl.Handle(
			context.Background(),
			buf,
			time.Now(),
			f.command,
		)
		if err != nil {
			t.Fatal(err)
		}

		got, ok := findFact[fact.AggregateInstanceLoaded](buf.Facts())
		if !ok {
			t.Fatal("expected AggregateInstanceLoaded fact")
		}

		xtesting.Expect(
			t,
			"unexpected fact",
			got,
			fact.AggregateInstanceLoaded{
				Handler:    f.cfg,
				InstanceID: "<instance-A1>",
				Root: &stubs.AggregateRootStub{
					AppliedEvents: []dogma.Event{stubs.EventA1},
				},
				Envelope: f.command,
			},
		)
	})

	t.Run("passes an aggregate root with historical events applied when the instance exists", func(t *testing.T) {
		f := newControllerTestFixture()
		seedControllerInstance(t, f)
		called := false

		f.handler.HandleCommandFunc = func(
			r *stubs.AggregateRootStub,
			_ dogma.AggregateCommandScope[*stubs.AggregateRootStub],
			_ dogma.Command,
		) {
			called = true
			xtesting.Expect(
				t,
				"unexpected aggregate root",
				r,
				&stubs.AggregateRootStub{
					AppliedEvents: []dogma.Event{stubs.EventA1},
				},
			)
		}

		_, err := f.ctrl.Handle(
			context.Background(),
			fact.Ignore,
			time.Now(),
			f.command,
		)
		if err != nil {
			t.Fatal(err)
		}

		xtesting.Expect(t, "expected handler to be called", called, true)
	})

	t.Run("provides more context to UnexpectedMessage panics from RouteCommandToInstance", func(t *testing.T) {
		f := newControllerTestFixture()
		f.handler.RouteCommandToInstanceFunc = func(dogma.Command) string {
			panic(dogma.UnexpectedMessage)
		}

		x := mustPanicUnexpectedMessage(t, func() {
			_, _ = f.ctrl.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				f.command,
			)
		})

		xtesting.Expect(t, "unexpected handler", x.Handler, f.cfg)
		xtesting.Expect(t, "unexpected interface", x.Interface, "AggregateMessageHandler")
		xtesting.Expect(t, "unexpected method", x.Method, "RouteCommandToInstance")
		xtesting.Expect(t, "unexpected message", x.Message, f.command.Message)
	})

	t.Run("provides more context to UnexpectedMessage panics from HandleCommand", func(t *testing.T) {
		f := newControllerTestFixture()
		f.handler.HandleCommandFunc = func(
			*stubs.AggregateRootStub,
			dogma.AggregateCommandScope[*stubs.AggregateRootStub],
			dogma.Command,
		) {
			panic(dogma.UnexpectedMessage)
		}

		x := mustPanicUnexpectedMessage(t, func() {
			_, _ = f.ctrl.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				f.command,
			)
		})

		xtesting.Expect(t, "unexpected handler", x.Handler, f.cfg)
		xtesting.Expect(t, "unexpected interface", x.Interface, "AggregateMessageHandler")
		xtesting.Expect(t, "unexpected method", x.Method, "HandleCommand")
		xtesting.Expect(t, "unexpected message", x.Message, f.command.Message)
	})

	t.Run("provides more context to UnexpectedMessage panics from ApplyEvent when called with new events", func(t *testing.T) {
		f := newControllerTestFixture()
		f.handler.HandleCommandFunc = func(
			_ *stubs.AggregateRootStub,
			s dogma.AggregateCommandScope[*stubs.AggregateRootStub],
			_ dogma.Command,
		) {
			s.RecordEvent(stubs.EventA1)
		}

		f.handler.NewFunc = func() *stubs.AggregateRootStub {
			return &stubs.AggregateRootStub{
				ApplyEventFunc: func(dogma.Event) {
					panic(dogma.UnexpectedMessage)
				},
			}
		}

		x := mustPanicUnexpectedMessage(t, func() {
			_, _ = f.ctrl.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				f.command,
			)
		})

		xtesting.Expect(t, "unexpected handler", x.Handler, f.cfg)
		xtesting.Expect(t, "unexpected interface", x.Interface, "AggregateRoot")
		xtesting.Expect(t, "unexpected method", x.Method, "ApplyEvent")
		xtesting.Expect(t, "unexpected message", x.Message, stubs.EventA1)
	})

	t.Run("provides more context to UnexpectedMessage panics from ApplyEvent when called with historical events", func(t *testing.T) {
		f := newControllerTestFixture()

		// Prevent snapshot so that historical events are replayed.
		f.handler.NewFunc = func() *stubs.AggregateRootStub {
			return &stubs.AggregateRootStub{
				MarshalBinaryFunc: func() ([]byte, error) {
					return nil, dogma.ErrNotSupported
				},
			}
		}

		seedControllerInstance(t, f)

		f.handler.NewFunc = func() *stubs.AggregateRootStub {
			return &stubs.AggregateRootStub{
				ApplyEventFunc: func(dogma.Event) {
					panic(dogma.UnexpectedMessage)
				},
			}
		}

		x := mustPanicUnexpectedMessage(t, func() {
			_, _ = f.ctrl.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				f.command,
			)
		})

		xtesting.Expect(t, "unexpected handler", x.Handler, f.cfg)
		xtesting.Expect(t, "unexpected interface", x.Interface, "AggregateRoot")
		xtesting.Expect(t, "unexpected method", x.Method, "ApplyEvent")
		xtesting.Expect(t, "unexpected message", x.Message, stubs.EventA1)
	})

	t.Run("panics if MarshalBinary fails with a non-ErrNotSupported error", func(t *testing.T) {
		f := newControllerTestFixture()
		f.handler.NewFunc = func() *stubs.AggregateRootStub {
			return &stubs.AggregateRootStub{
				MarshalBinaryFunc: func() ([]byte, error) {
					return nil, errors.New("<marshal error>")
				},
			}
		}
		f.handler.HandleCommandFunc = func(
			_ *stubs.AggregateRootStub,
			s dogma.AggregateCommandScope[*stubs.AggregateRootStub],
			_ dogma.Command,
		) {
			s.RecordEvent(stubs.EventA1)
		}

		xtesting.ExpectPanic(
			t,
			"the '<name>' aggregate message handler behaved unexpectedly in *stubs.AggregateRootStub.MarshalBinary(): unable to marshal the aggregate root: <marshal error>",
			func() {
				_, _ = f.ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					f.command,
				)
			},
		)
	})

	t.Run("does not panic if MarshalBinary returns ErrNotSupported", func(t *testing.T) {
		f := newControllerTestFixture()
		f.handler.NewFunc = func() *stubs.AggregateRootStub {
			return &stubs.AggregateRootStub{
				MarshalBinaryFunc: func() ([]byte, error) {
					return nil, dogma.ErrNotSupported
				},
			}
		}
		f.handler.HandleCommandFunc = func(
			_ *stubs.AggregateRootStub,
			s dogma.AggregateCommandScope[*stubs.AggregateRootStub],
			_ dogma.Command,
		) {
			s.RecordEvent(stubs.EventA1)
		}

		_, err := f.ctrl.Handle(
			context.Background(),
			fact.Ignore,
			time.Now(),
			f.command,
		)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("panics if UnmarshalBinary fails when loading a snapshot", func(t *testing.T) {
		f := newControllerTestFixture()
		seedControllerInstance(t, f)

		f.handler.NewFunc = func() *stubs.AggregateRootStub {
			return &stubs.AggregateRootStub{
				UnmarshalBinaryFunc: func([]byte) error {
					return errors.New("<unmarshal error>")
				},
			}
		}

		xtesting.ExpectPanic(
			t,
			"the '<name>' aggregate message handler behaved unexpectedly in *stubs.AggregateRootStub.UnmarshalBinary(): unable to unmarshal the aggregate root: <unmarshal error>",
			func() {
				_, _ = f.ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					f.command,
				)
			},
		)
	})

	t.Run("calls UnmarshalBinary when MarshalBinary returns nil", func(t *testing.T) {
		f := newControllerTestFixture()

		f.handler.NewFunc = func() *stubs.AggregateRootStub {
			return &stubs.AggregateRootStub{
				MarshalBinaryFunc: func() ([]byte, error) {
					return nil, nil
				},
			}
		}
		f.handler.HandleCommandFunc = func(
			_ *stubs.AggregateRootStub,
			s dogma.AggregateCommandScope[*stubs.AggregateRootStub],
			_ dogma.Command,
		) {
			s.RecordEvent(stubs.EventA1)
		}

		_, err := f.ctrl.Handle(
			context.Background(),
			fact.Ignore,
			time.Now(),
			f.command,
		)
		if err != nil {
			t.Fatal(err)
		}

		called := false
		f.handler.NewFunc = func() *stubs.AggregateRootStub {
			return &stubs.AggregateRootStub{
				UnmarshalBinaryFunc: func([]byte) error {
					called = true
					return nil
				},
			}
		}

		_, err = f.ctrl.Handle(
			context.Background(),
			fact.Ignore,
			time.Now(),
			f.command,
		)
		if err != nil {
			t.Fatal(err)
		}

		xtesting.Expect(t, "expected UnmarshalBinary to be called", called, true)
	})

	t.Run("calls UnmarshalBinary when MarshalBinary returns an empty slice", func(t *testing.T) {
		f := newControllerTestFixture()

		f.handler.NewFunc = func() *stubs.AggregateRootStub {
			return &stubs.AggregateRootStub{
				MarshalBinaryFunc: func() ([]byte, error) {
					return []byte{}, nil
				},
			}
		}
		f.handler.HandleCommandFunc = func(
			_ *stubs.AggregateRootStub,
			s dogma.AggregateCommandScope[*stubs.AggregateRootStub],
			_ dogma.Command,
		) {
			s.RecordEvent(stubs.EventA1)
		}

		_, err := f.ctrl.Handle(
			context.Background(),
			fact.Ignore,
			time.Now(),
			f.command,
		)
		if err != nil {
			t.Fatal(err)
		}

		called := false
		f.handler.NewFunc = func() *stubs.AggregateRootStub {
			return &stubs.AggregateRootStub{
				UnmarshalBinaryFunc: func([]byte) error {
					called = true
					return nil
				},
			}
		}

		_, err = f.ctrl.Handle(
			context.Background(),
			fact.Ignore,
			time.Now(),
			f.command,
		)
		if err != nil {
			t.Fatal(err)
		}

		xtesting.Expect(t, "expected UnmarshalBinary to be called", called, true)
	})
}

func TestControllerReset(t *testing.T) {
	f := newControllerTestFixture()
	seedControllerInstance(t, f)
	f.ctrl.Reset()

	buf := &fact.Buffer{}
	_, err := f.ctrl.Handle(
		context.Background(),
		buf,
		time.Now(),
		f.command,
	)
	if err != nil {
		t.Fatal(err)
	}

	if _, ok := findFact[fact.AggregateInstanceLoaded](buf.Facts()); ok {
		t.Fatal("did not expect AggregateInstanceLoaded fact after reset")
	}
}

type controllerTestFixture struct {
	messageIDs envelope.MessageIDGenerator
	handler    *stubs.AggregateMessageHandlerStub[*stubs.AggregateRootStub]
	cfg        *config.Aggregate
	ctrl       *aggregate.Controller
	command    *envelope.Envelope
}

func newControllerTestFixture() *controllerTestFixture {
	f := &controllerTestFixture{
		command: envelope.NewCommand(
			"1000",
			stubs.CommandA1,
			time.Now(),
		),
	}

	f.handler = &stubs.AggregateMessageHandlerStub[*stubs.AggregateRootStub]{
		ConfigureFunc: func(c dogma.AggregateConfigurer) {
			c.Identity("<name>", "e8fd6bd4-c3a3-4eb4-bf0f-56862a123229")
			c.Routes(
				dogma.HandlesCommand[*stubs.CommandStub[stubs.TypeA]](),
				dogma.RecordsEvent[*stubs.EventStub[stubs.TypeA]](),
			)
		},
		RouteCommandToInstanceFunc: func(m dogma.Command) string {
			switch x := m.(type) {
			case *stubs.CommandStub[stubs.TypeA]:
				return fmt.Sprintf("<instance-%s>", x.Content)
			default:
				panic(dogma.UnexpectedMessage)
			}
		},
	}

	f.cfg = runtimeconfig.FromAggregate(f.handler)
	f.ctrl = &aggregate.Controller{
		Config:     f.cfg,
		MessageIDs: &f.messageIDs,
	}

	f.messageIDs.Reset()

	return f
}

func seedControllerInstance(t *testing.T, f *controllerTestFixture) {
	t.Helper()

	f.handler.HandleCommandFunc = func(
		_ *stubs.AggregateRootStub,
		s dogma.AggregateCommandScope[*stubs.AggregateRootStub],
		_ dogma.Command,
	) {
		s.RecordEvent(stubs.EventA1)
	}

	_, err := f.ctrl.Handle(
		context.Background(),
		fact.Ignore,
		time.Now(),
		f.command,
	)
	if err != nil {
		t.Fatal(err)
	}

	f.handler.HandleCommandFunc = nil
}

func mustPanicUnexpectedBehavior(t *testing.T, fn func()) panicx.UnexpectedBehavior {
	t.Helper()

	r := recoverPanic(t, fn)
	x, ok := r.(panicx.UnexpectedBehavior)
	if !ok {
		t.Fatalf("expected UnexpectedBehavior panic, got %T", r)
	}

	return x
}

func mustPanicUnexpectedMessage(t *testing.T, fn func()) panicx.UnexpectedMessage {
	t.Helper()

	r := recoverPanic(t, fn)
	x, ok := r.(panicx.UnexpectedMessage)
	if !ok {
		t.Fatalf("expected UnexpectedMessage panic, got %T", r)
	}

	return x
}

func recoverPanic(t *testing.T, fn func()) any {
	t.Helper()

	var r any

	func() {
		defer func() {
			r = recover()
		}()

		fn()
	}()

	if r == nil {
		t.Fatal("expected panic")
	}

	return r
}

func expectLocation(t *testing.T, loc location.Location, fileSuffix string) {
	t.Helper()

	if loc.Func == "" {
		t.Fatal("expected func to be set in location")
	}

	if !strings.HasSuffix(loc.File, fileSuffix) {
		t.Fatalf("unexpected file in location: got %s, want suffix %s", loc.File, fileSuffix)
	}

	if loc.Line == 0 {
		t.Fatal("expected line to be set in location")
	}
}

func findFact[T any](facts []fact.Fact) (T, bool) {
	var zero T

	for _, f := range facts {
		x, ok := f.(T)
		if ok {
			return x, true
		}
	}

	return zero, false
}
