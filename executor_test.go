package testkit_test

import (
	"context"
	"errors"
	"testing"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/enginekit/enginetest/stubs"
	. "github.com/dogmatiq/testkit"
	"github.com/dogmatiq/testkit/internal/testingmock"
	"github.com/dogmatiq/testkit/x/xtesting"
)

func TestInterceptCommandExecutor(t *testing.T) {
	newFixture := func(t *testing.T) (*testingmock.T, dogma.Application, CommandExecutorInterceptor, CommandExecutorInterceptor) {
		t.Helper()

		testingT := &testingmock.T{}

		app := &ApplicationStub{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "b5453327-a0fa-4e94-bb46-8464e727c4fd")
				c.Routes(
					dogma.ViaIntegration(&IntegrationMessageHandlerStub{
						ConfigureFunc: func(c dogma.IntegrationConfigurer) {
							c.Identity("<handler-name>", "67c167a8-d09e-4827-beab-7c8c9817bb1a")
							c.Routes(
								dogma.HandlesCommand[*CommandStub[TypeA]](),
								dogma.RecordsEvent[*EventStub[TypeA]](),
							)
						},
						HandleCommandFunc: func(_ context.Context, s dogma.IntegrationCommandScope, _ dogma.Command) error {
							s.RecordEvent(EventA1)
							return nil
						},
					}),
				)
			},
		}

		doNothing := func(context.Context, dogma.Command, []dogma.ExecuteCommandOption, dogma.CommandExecutor) error {
			return nil
		}

		executeCommandAndReturnError := func(
			ctx context.Context,
			m dogma.Command,
			opts []dogma.ExecuteCommandOption,
			e dogma.CommandExecutor,
		) error {
			xtesting.Expect(t, "unexpected command", m, CommandA1)

			err := e.ExecuteCommand(ctx, m, opts...)
			if err != nil {
				t.Fatalf("unexpected execute error: %v", err)
			}

			return errors.New("<error>")
		}

		return testingT, app, doNothing, executeCommandAndReturnError
	}

	t.Run("it panics if the interceptor function is nil", func(t *testing.T) {
		xtesting.ExpectPanic(
			t,
			"InterceptCommandExecutor(<nil>): function must not be nil",
			func() {
				InterceptCommandExecutor(nil)
			},
		)
	})

	t.Run("used as a TestOption", func(t *testing.T) {
		t.Run("it intercepts calls to ExecuteCommand()", func(t *testing.T) {
			testingT, app, _, executeCommandAndReturnError := newFixture(t)

			tc := Begin(
				testingT,
				app,
				InterceptCommandExecutor(executeCommandAndReturnError),
			)

			tc.EnableHandlers("<handler-name>")

			tc.Expect(
				Call(func() {
					err := tc.CommandExecutor().ExecuteCommand(
						context.Background(),
						CommandA1,
					)
					if err == nil || err.Error() != "<error>" {
						t.Fatalf("unexpected error: %v", err)
					}
				}),
				ToRecordEvent(EventA1),
			)
		})
	})

	t.Run("used as a CallOption", func(t *testing.T) {
		t.Run("it intercepts calls to ExecuteCommand()", func(t *testing.T) {
			_, app, _, executeCommandAndReturnError := newFixture(t)

			tc := Begin(
				&testingmock.T{},
				app,
			)

			tc.EnableHandlers("<handler-name>")

			tc.Expect(
				Call(
					func() {
						err := tc.CommandExecutor().ExecuteCommand(
							context.Background(),
							CommandA1,
						)
						if err == nil || err.Error() != "<error>" {
							t.Fatalf("unexpected error: %v", err)
						}
					},
					InterceptCommandExecutor(executeCommandAndReturnError),
				),
				ToRecordEvent(EventA1),
			)
		})

		t.Run("it uninstalls the interceptor upon completion of the Call() action", func(t *testing.T) {
			_, app, _, executeCommandAndReturnError := newFixture(t)

			tc := Begin(
				&testingmock.T{},
				app,
			)

			tc.Prepare(
				Call(
					func() {
						err := tc.CommandExecutor().ExecuteCommand(
							context.Background(),
							CommandA1,
						)
						if err == nil || err.Error() != "<error>" {
							t.Fatalf("unexpected error: %v", err)
						}
					},
					InterceptCommandExecutor(executeCommandAndReturnError),
				),
				Call(
					func() {
						err := tc.CommandExecutor().ExecuteCommand(
							context.Background(),
							CommandA1,
						)
						if err != nil {
							t.Fatalf("unexpected error: %v", err)
						}
					},
				),
			)
		})

		t.Run("it re-installs the test-level interceptor upon completion of the Call() action", func(t *testing.T) {
			_, app, doNothing, executeCommandAndReturnError := newFixture(t)

			tc := Begin(
				&testingmock.T{},
				app,
				InterceptCommandExecutor(executeCommandAndReturnError),
			)

			tc.Prepare(
				Call(
					func() {
						err := tc.CommandExecutor().ExecuteCommand(
							context.Background(),
							CommandA1,
						)
						if err != nil {
							t.Fatalf("unexpected error: %v", err)
						}
					},
					InterceptCommandExecutor(doNothing),
				),
				Call(
					func() {
						err := tc.CommandExecutor().ExecuteCommand(
							context.Background(),
							CommandA1,
						)
						if err == nil || err.Error() != "<error>" {
							t.Fatalf("unexpected error: %v", err)
						}
					},
				),
			)
		})
	})
}
