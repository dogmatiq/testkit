package integration_test

import (
	"context"
	"errors"
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

func TestController(t *testing.T) {
	newController := func() (*IntegrationMessageHandlerStub, *config.Integration, *Controller, *envelope.Envelope) {
		var messageIDs envelope.MessageIDGenerator

		command := envelope.NewCommand(
			"1000",
			CommandA1,
			time.Now(),
		)

		handler := &IntegrationMessageHandlerStub{
			ConfigureFunc: func(c dogma.IntegrationConfigurer) {
				c.Identity("<name>", "8cbb8bca-b5eb-4c94-a877-dfc8dc9968ca")
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

	t.Run("func HandlerConfig()", func(t *testing.T) {
		_, cfg, ctrl, _ := newController()

		test.Expect(t, "unexpected handler config", ctrl.HandlerConfig(), cfg)
	})

	t.Run("func Tick()", func(t *testing.T) {
		t.Run("it does not return any envelopes", func(t *testing.T) {
			_, _, ctrl, _ := newController()

			envelopes, err := ctrl.Tick(
				context.Background(),
				fact.Ignore,
				time.Now(),
			)

			test.Expect(t, "unexpected error", err, nil)
			test.Expect(t, "unexpected envelopes", envelopes, []*envelope.Envelope(nil))
		})

		t.Run("it does not record any facts", func(t *testing.T) {
			_, _, ctrl, _ := newController()
			buf := &fact.Buffer{}

			_, err := ctrl.Tick(
				context.Background(),
				buf,
				time.Now(),
			)

			test.Expect(t, "unexpected error", err, nil)
			test.Expect(t, "unexpected facts", buf.Facts(), []fact.Fact(nil))
		})
	})

	t.Run("func Handle()", func(t *testing.T) {
		t.Run("it forwards the message to the handler", func(t *testing.T) {
			handler, _, ctrl, command := newController()
			called := false

			handler.HandleCommandFunc = func(
				_ context.Context,
				_ dogma.IntegrationCommandScope,
				m dogma.Command,
			) error {
				called = true
				test.Expect(t, "unexpected command", m, CommandA1)
				return nil
			}

			_, err := ctrl.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				command,
			)

			test.Expect(t, "unexpected error", err, nil)
			test.Expect(t, "expected handler to be called", called, true)
		})

		t.Run("it returns the recorded events", func(t *testing.T) {
			handler, cfg, ctrl, command := newController()

			handler.HandleCommandFunc = func(
				_ context.Context,
				s dogma.IntegrationCommandScope,
				_ dogma.Command,
			) error {
				s.RecordEvent(EventA1)
				s.RecordEvent(EventA2)
				return nil
			}

			now := time.Now()
			events, err := ctrl.Handle(
				context.Background(),
				fact.Ignore,
				now,
				command,
			)

			test.Expect(t, "unexpected error", err, nil)
			test.Expect(
				t,
				"unexpected events",
				events,
				[]*envelope.Envelope{
					command.NewEvent(
						"1",
						EventA1,
						now,
						envelope.Origin{
							Handler:     cfg,
							HandlerType: config.IntegrationHandlerType,
						},
						"8cbb8bca-b5eb-4c94-a877-dfc8dc9968ca",
						0,
					),
					command.NewEvent(
						"2",
						EventA2,
						now,
						envelope.Origin{
							Handler:     cfg,
							HandlerType: config.IntegrationHandlerType,
						},
						"8cbb8bca-b5eb-4c94-a877-dfc8dc9968ca",
						1,
					),
				},
			)
		})

		t.Run("it propagates handler errors", func(t *testing.T) {
			handler, _, ctrl, command := newController()
			expected := errors.New("<error>")

			handler.HandleCommandFunc = func(
				_ context.Context,
				_ dogma.IntegrationCommandScope,
				_ dogma.Command,
			) error {
				return expected
			}

			_, err := ctrl.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				command,
			)

			test.Expect(t, "unexpected error", err, expected)
		})

		t.Run("it provides more context to UnexpectedMessage panics from HandleCommand()", func(t *testing.T) {
			handler, cfg, ctrl, command := newController()

			handler.HandleCommandFunc = func(
				context.Context,
				dogma.IntegrationCommandScope,
				dogma.Command,
			) error {
				panic(dogma.UnexpectedMessage)
			}

			defer func() {
				r := recover()
				x, ok := r.(panicx.UnexpectedMessage)
				if !ok {
					t.Fatalf("expected UnexpectedMessage panic, got %T", r)
				}

				test.Expect(t, "unexpected handler", x.Handler, cfg)
				test.Expect(t, "unexpected interface", x.Interface, "IntegrationMessageHandler")
				test.Expect(t, "unexpected method", x.Method, "HandleCommand")
				test.Expect(t, "unexpected message", x.Message, command.Message)
			}()

			ctrl.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				command,
			)
		})
	})

	t.Run("func Reset()", func(t *testing.T) {
		_, _, ctrl, _ := newController()
		ctrl.Reset()
	})
}
