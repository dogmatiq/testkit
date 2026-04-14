package integration_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/enginekit/config"
	"github.com/dogmatiq/enginekit/config/runtimeconfig"
	. "github.com/dogmatiq/enginekit/enginetest/stubs"
	. "github.com/dogmatiq/testkit/engine/internal/integration"
	"github.com/dogmatiq/testkit/engine/internal/panicx"
	"github.com/dogmatiq/testkit/envelope"
	"github.com/dogmatiq/testkit/fact"
	"github.com/dogmatiq/testkit/internal/test"
)

func TestScope(t *testing.T) {
	newController := func() (*IntegrationMessageHandlerStub, *config.Integration, *Controller, *envelope.Envelope) {
		var messageIDs envelope.MessageIDGenerator

		command := envelope.NewCommand(
			"1000",
			CommandA1,
			time.Now(),
		)

		handler := &IntegrationMessageHandlerStub{
			ConfigureFunc: func(c dogma.IntegrationConfigurer) {
				c.Identity("<name>", "24ec3839-5d51-4904-9b45-34b5282e7f24")
				c.Routes(
					dogma.HandlesCommand[*CommandStub[TypeA]](),
					dogma.RecordsEvent[*EventStub[TypeA]](),
				)
			},
		}

		cfg := runtimeconfig.FromIntegration(handler)
		ctrl := &Controller{
			Config:     cfg,
			MessageIDs: &messageIDs,
		}

		messageIDs.Reset()

		return handler, cfg, ctrl, command
	}

	t.Run("func RecordEvent()", func(t *testing.T) {
		t.Run("it records a fact", func(t *testing.T) {
			handler, cfg, ctrl, command := newController()
			handler.HandleCommandFunc = func(
				_ context.Context,
				s dogma.IntegrationCommandScope,
				_ dogma.Command,
			) error {
				s.RecordEvent(EventA1)
				return nil
			}

			buf := &fact.Buffer{}
			now := time.Now()
			_, err := ctrl.Handle(
				context.Background(),
				buf,
				now,
				command,
			)

			test.Expect(t, "unexpected error", err, nil)
			test.Expect(
				t,
				"unexpected facts",
				buf.Facts(),
				[]fact.Fact{
					fact.EventRecordedByIntegration{
						Handler:  cfg,
						Envelope: command,
						EventEnvelope: command.NewEvent(
							"1",
							EventA1,
							now,
							envelope.Origin{
								Handler:     cfg,
								HandlerType: config.IntegrationHandlerType,
							},
							"24ec3839-5d51-4904-9b45-34b5282e7f24",
							0,
						),
					},
				},
			)
		})

		t.Run("it panics if the event type is not configured to be produced", func(t *testing.T) {
			handler, cfg, ctrl, command := newController()
			handler.HandleCommandFunc = func(
				_ context.Context,
				s dogma.IntegrationCommandScope,
				_ dogma.Command,
			) error {
				s.RecordEvent(EventX1)
				return nil
			}

			defer func() {
				r := recover()
				x, ok := r.(panicx.UnexpectedBehavior)
				if !ok {
					t.Fatalf("expected UnexpectedBehavior panic, got %T", r)
				}

				test.Expect(t, "unexpected handler", x.Handler, cfg)
				test.Expect(t, "unexpected interface", x.Interface, "IntegrationMessageHandler")
				test.Expect(t, "unexpected method", x.Method, "HandleCommand")
				test.Expect(t, "unexpected implementation", x.Implementation, cfg.Source.Get())
				test.Expect(t, "unexpected message", x.Message, command.Message)
				test.Expect(
					t,
					"unexpected description",
					x.Description,
					"recorded an event of type *stubs.EventStub[TypeX], which is not produced by this handler",
				)
				if x.Location.Func == "" {
					t.Fatal("expected non-empty location func")
				}
				if !strings.HasSuffix(x.Location.File, "/engine/internal/integration/scope_test.go") {
					t.Fatalf("unexpected location file: %s", x.Location.File)
				}
				if x.Location.Line == 0 {
					t.Fatal("expected non-zero location line")
				}
			}()

			ctrl.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				command,
			)
		})

		t.Run("it panics if the event is invalid", func(t *testing.T) {
			handler, cfg, ctrl, command := newController()
			handler.HandleCommandFunc = func(
				_ context.Context,
				s dogma.IntegrationCommandScope,
				_ dogma.Command,
			) error {
				s.RecordEvent(&EventStub[TypeA]{
					ValidationError: "<invalid>",
				})
				return nil
			}

			defer func() {
				r := recover()
				x, ok := r.(panicx.UnexpectedBehavior)
				if !ok {
					t.Fatalf("expected UnexpectedBehavior panic, got %T", r)
				}

				test.Expect(t, "unexpected handler", x.Handler, cfg)
				test.Expect(t, "unexpected interface", x.Interface, "IntegrationMessageHandler")
				test.Expect(t, "unexpected method", x.Method, "HandleCommand")
				test.Expect(t, "unexpected implementation", x.Implementation, cfg.Source.Get())
				test.Expect(t, "unexpected message", x.Message, command.Message)
				test.Expect(
					t,
					"unexpected description",
					x.Description,
					"recorded an invalid *stubs.EventStub[TypeA] event: <invalid>",
				)
				if x.Location.Func == "" {
					t.Fatal("expected non-empty location func")
				}
				if !strings.HasSuffix(x.Location.File, "/engine/internal/integration/scope_test.go") {
					t.Fatalf("unexpected location file: %s", x.Location.File)
				}
				if x.Location.Line == 0 {
					t.Fatal("expected non-zero location line")
				}
			}()

			ctrl.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				command,
			)
		})
	})

	t.Run("func Log()", func(t *testing.T) {
		t.Run("it records a fact", func(t *testing.T) {
			handler, cfg, ctrl, command := newController()
			handler.HandleCommandFunc = func(
				_ context.Context,
				s dogma.IntegrationCommandScope,
				_ dogma.Command,
			) error {
				s.Log("<format>", "<arg-1>", "<arg-2>")
				return nil
			}

			buf := &fact.Buffer{}
			_, err := ctrl.Handle(
				context.Background(),
				buf,
				time.Now(),
				command,
			)

			test.Expect(t, "unexpected error", err, nil)
			test.Expect(
				t,
				"unexpected facts",
				buf.Facts(),
				[]fact.Fact{
					fact.MessageLoggedByIntegration{
						Handler:   cfg,
						Envelope:  command,
						LogFormat: "<format>",
						LogArguments: []any{
							"<arg-1>",
							"<arg-2>",
						},
					},
				},
			)
		})
	})
}
