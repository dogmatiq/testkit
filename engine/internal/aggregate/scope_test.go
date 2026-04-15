package aggregate_test

import (
	"context"
	"testing"
	"time"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/enginekit/config"
	"github.com/dogmatiq/enginekit/config/runtimeconfig"
	stubs "github.com/dogmatiq/enginekit/enginetest/stubs"
	"github.com/dogmatiq/testkit/engine/internal/aggregate"
	"github.com/dogmatiq/testkit/envelope"
	"github.com/dogmatiq/testkit/fact"
	"github.com/dogmatiq/testkit/internal/x/xtesting"
)

func TestScopeRecordEvent(t *testing.T) {
	t.Run("records facts about instance creation and the event when the instance does not exist", func(t *testing.T) {
		f := newScopeTestFixture()
		f.handler.HandleCommandFunc = func(
			_ dogma.AggregateRoot,
			s dogma.AggregateCommandScope,
			_ dogma.Command,
		) {
			s.RecordEvent(stubs.EventA1)
		}

		now := time.Now()
		buf := &fact.Buffer{}
		_, err := f.ctrl.Handle(
			context.Background(),
			buf,
			now,
			f.command,
		)
		if err != nil {
			t.Fatal(err)
		}

		xtesting.Expect(
			t,
			"unexpected facts",
			buf.Facts(),
			[]fact.Fact{
				fact.AggregateInstanceNotFound{
					Handler:    f.cfg,
					InstanceID: "<instance>",
					Envelope:   f.command,
				},
				fact.AggregateInstanceCreated{
					Handler:    f.cfg,
					InstanceID: "<instance>",
					Root: &stubs.AggregateRootStub{
						AppliedEvents: []dogma.Event{stubs.EventA1},
					},
					Envelope: f.command,
				},
				fact.EventRecordedByAggregate{
					Handler:    f.cfg,
					InstanceID: "<instance>",
					Root: &stubs.AggregateRootStub{
						AppliedEvents: []dogma.Event{stubs.EventA1},
					},
					Envelope: f.command,
					EventEnvelope: f.command.NewEvent(
						"1",
						stubs.EventA1,
						now,
						envelope.Origin{
							Handler:     f.cfg,
							HandlerType: config.AggregateHandlerType,
							InstanceID:  "<instance>",
						},
						"aa9aa868-af3f-5dbb-a718-223782f4c77c",
						0,
					),
				},
			},
		)
	})

	t.Run("records a fact when the instance exists", func(t *testing.T) {
		f := newScopeTestFixture()
		seedScopeInstance(t, f)
		f.messageIDs.Reset()

		f.handler.HandleCommandFunc = func(
			_ dogma.AggregateRoot,
			s dogma.AggregateCommandScope,
			_ dogma.Command,
		) {
			s.RecordEvent(stubs.EventA1)
		}

		buf := &fact.Buffer{}
		now := time.Now()
		_, err := f.ctrl.Handle(
			context.Background(),
			buf,
			now,
			f.command,
		)
		if err != nil {
			t.Fatal(err)
		}

		got, ok := findFact[fact.EventRecordedByAggregate](buf.Facts())
		if !ok {
			t.Fatal("expected EventRecordedByAggregate fact")
		}

		xtesting.Expect(
			t,
			"unexpected fact",
			got,
			fact.EventRecordedByAggregate{
				Handler:    f.cfg,
				InstanceID: "<instance>",
				Root: &stubs.AggregateRootStub{
					AppliedEvents: []dogma.Event{stubs.EventA1, stubs.EventA1},
				},
				Envelope: f.command,
				EventEnvelope: f.command.NewEvent(
					"1",
					stubs.EventA1,
					now,
					envelope.Origin{
						Handler:     f.cfg,
						HandlerType: config.AggregateHandlerType,
						InstanceID:  "<instance>",
					},
					"aa9aa868-af3f-5dbb-a718-223782f4c77c",
					1,
				),
			},
		)
	})

	t.Run("does not record a fact about instance creation when the instance exists", func(t *testing.T) {
		f := newScopeTestFixture()
		seedScopeInstance(t, f)

		f.handler.HandleCommandFunc = func(
			_ dogma.AggregateRoot,
			s dogma.AggregateCommandScope,
			_ dogma.Command,
		) {
			s.RecordEvent(stubs.EventA1)
		}

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

		if _, ok := findFact[fact.AggregateInstanceCreated](buf.Facts()); ok {
			t.Fatal("did not expect AggregateInstanceCreated fact")
		}
	})

	t.Run("panics if the event type is not configured to be produced", func(t *testing.T) {
		f := newScopeTestFixture()
		seedScopeInstance(t, f)

		f.handler.HandleCommandFunc = func(
			_ dogma.AggregateRoot,
			s dogma.AggregateCommandScope,
			_ dogma.Command,
		) {
			s.RecordEvent(stubs.EventX1)
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
		xtesting.Expect(t, "unexpected method", x.Method, "HandleCommand")
		xtesting.Expect(t, "unexpected implementation", x.Implementation, f.cfg.Source.Get())
		xtesting.Expect(t, "unexpected message", x.Message, f.command.Message)
		xtesting.Expect(
			t,
			"unexpected description",
			x.Description,
			"recorded an event of type *stubs.EventStub[TypeX], which is not produced by this handler",
		)
		expectLocation(t, x.Location, "/engine/internal/aggregate/scope_test.go")
	})

	t.Run("panics if the event is invalid", func(t *testing.T) {
		f := newScopeTestFixture()
		seedScopeInstance(t, f)

		f.handler.HandleCommandFunc = func(
			_ dogma.AggregateRoot,
			s dogma.AggregateCommandScope,
			_ dogma.Command,
		) {
			s.RecordEvent(&stubs.EventStub[stubs.TypeA]{
				ValidationError: "<invalid>",
			})
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
		xtesting.Expect(t, "unexpected method", x.Method, "HandleCommand")
		xtesting.Expect(t, "unexpected implementation", x.Implementation, f.cfg.Source.Get())
		xtesting.Expect(t, "unexpected message", x.Message, f.command.Message)
		xtesting.Expect(
			t,
			"unexpected description",
			x.Description,
			"recorded an invalid *stubs.EventStub[TypeA] event: <invalid>",
		)
		expectLocation(t, x.Location, "/engine/internal/aggregate/scope_test.go")
	})
}

func TestScopeInstanceID(t *testing.T) {
	f := newScopeTestFixture()
	called := false

	f.handler.HandleCommandFunc = func(
		_ dogma.AggregateRoot,
		s dogma.AggregateCommandScope,
		_ dogma.Command,
	) {
		called = true
		xtesting.Expect(t, "unexpected instance ID", s.InstanceID(), "<instance>")
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
}

func TestScopeLog(t *testing.T) {
	f := newScopeTestFixture()
	f.handler.HandleCommandFunc = func(
		_ dogma.AggregateRoot,
		s dogma.AggregateCommandScope,
		_ dogma.Command,
	) {
		s.Log("<format>", "<arg-1>", "<arg-2>")
	}

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

	xtesting.Expect(
		t,
		"unexpected facts",
		buf.Facts(),
		[]fact.Fact{
			fact.AggregateInstanceNotFound{
				Handler:    f.cfg,
				InstanceID: "<instance>",
				Envelope:   f.command,
			},
			fact.MessageLoggedByAggregate{
				Handler:    f.cfg,
				InstanceID: "<instance>",
				Root:       &stubs.AggregateRootStub{},
				Envelope:   f.command,
				LogFormat:  "<format>",
				LogArguments: []any{
					"<arg-1>",
					"<arg-2>",
				},
			},
		},
	)
}

type scopeTestFixture struct {
	messageIDs envelope.MessageIDGenerator
	handler    *stubs.AggregateMessageHandlerStub
	cfg        *config.Aggregate
	ctrl       *aggregate.Controller
	command    *envelope.Envelope
}

func newScopeTestFixture() *scopeTestFixture {
	f := &scopeTestFixture{
		command: envelope.NewCommand(
			"1000",
			stubs.CommandA1,
			time.Now(),
		),
	}

	f.handler = &stubs.AggregateMessageHandlerStub{
		ConfigureFunc: func(c dogma.AggregateConfigurer) {
			c.Identity("<name>", "fd88e430-32fe-49a6-888f-f678dcf924ef")
			c.Routes(
				dogma.HandlesCommand[*stubs.CommandStub[stubs.TypeA]](),
				dogma.RecordsEvent[*stubs.EventStub[stubs.TypeA]](),
			)
		},
		RouteCommandToInstanceFunc: func(m dogma.Command) string {
			switch m.(type) {
			case *stubs.CommandStub[stubs.TypeA]:
				return "<instance>"
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

func seedScopeInstance(t *testing.T, f *scopeTestFixture) {
	t.Helper()

	f.handler.HandleCommandFunc = func(
		_ dogma.AggregateRoot,
		s dogma.AggregateCommandScope,
		_ dogma.Command,
	) {
		s.RecordEvent(stubs.EventA1)
	}

	_, err := f.ctrl.Handle(
		context.Background(),
		fact.Ignore,
		time.Now(),
		envelope.NewCommand(
			"2000",
			stubs.CommandA2,
			time.Now(),
		),
	)
	if err != nil {
		t.Fatal(err)
	}

	f.messageIDs.Reset()
}
