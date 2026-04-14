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

func TestTest_Prepare(t *testing.T) {
	t.Run("it fails the test if the action returns an error", func(t *testing.T) {
		app := &ApplicationStub{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "c654acb2-3e87-493a-8b9b-f662cd5e0f55")
			},
		}

		mt := &testingmock.T{FailSilently: true}

		Begin(mt, app).
			Prepare(
				noopAction{errors.New("<error>")},
			)

		if !mt.Failed() {
			t.Fatal("expected test to fail")
		}
	})
}

func TestTest_Expect(t *testing.T) {
	t.Run("it fails the test if the action returns an error", func(t *testing.T) {
		app := &ApplicationStub{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "c691b2ca-4c07-4473-bc42-060266cc7a56")
			},
		}

		mt := &testingmock.T{FailSilently: true}

		Begin(mt, app).
			Expect(
				noopAction{errors.New("<error>")},
				pass,
			)

		if !mt.Failed() {
			t.Fatal("expected test to fail")
		}
	})
}

func TestTest_EnableHandlers(t *testing.T) {
	t.Run("it enables the handler", func(t *testing.T) {
		called := false
		app := &ApplicationStub{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "7d5b218d-d69b-48d5-8831-2af77561ee62")
				c.Routes(
					dogma.ViaProjection(&ProjectionMessageHandlerStub{
						ConfigureFunc: func(c dogma.ProjectionConfigurer) {
							c.Identity("<projection>", "fb5f05c0-589c-4d64-9599-a4875b5a3569")
							c.Routes(
								dogma.HandlesEvent[*EventStub[TypeA]](),
							)
						},
						HandleEventFunc: func(
							_ context.Context,
							s dogma.ProjectionEventScope,
							_ dogma.Event,
						) (uint64, error) {
							called = true
							return s.Offset() + 1, nil
						},
					}),
				)
			},
		}

		Begin(&testingmock.T{}, app).
			EnableHandlers("<projection>").
			Prepare(RecordEvent(EventA1))

		if !called {
			t.Fatal("expected handler to be called")
		}
	})

	t.Run("it panics if the handler is not recognized", func(t *testing.T) {
		app := &ApplicationStub{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "7d5b218d-d69b-48d5-8831-2af77561ee62")
			},
		}

		xtesting.ExpectPanic(
			t,
			`the "<app>" application does not have a handler named "<projection>"`,
			func() {
				Begin(&testingmock.T{}, app).
					EnableHandlers("<projection>")
			},
		)
	})

	t.Run("it panics if the handler is disabled by its own configuration", func(t *testing.T) {
		app := &ApplicationStub{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "7d5b218d-d69b-48d5-8831-2af77561ee62")
				c.Routes(
					dogma.ViaProjection(&ProjectionMessageHandlerStub{
						ConfigureFunc: func(c dogma.ProjectionConfigurer) {
							c.Identity("<projection>", "fb5f05c0-589c-4d64-9599-a4875b5a3569")
							c.Routes(
								dogma.HandlesEvent[*EventStub[TypeA]](),
							)
							c.Disable()
						},
					}),
				)
			},
		}

		xtesting.ExpectPanic(
			t,
			`cannot enable the "<projection>" handler, it has been disabled by a call to ProjectionConfigurer.Disable()`,
			func() {
				Begin(&testingmock.T{}, app).
					EnableHandlers("<projection>")
			},
		)
	})
}

func TestTest_EnableHandlersLike(t *testing.T) {
	t.Run("it enables handlers with matching names", func(t *testing.T) {
		called := false
		app := &ApplicationStub{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "7d5b218d-d69b-48d5-8831-2af77561ee62")
				c.Routes(
					dogma.ViaProjection(&ProjectionMessageHandlerStub{
						ConfigureFunc: func(c dogma.ProjectionConfigurer) {
							c.Identity("<projection>", "fb5f05c0-589c-4d64-9599-a4875b5a3569")
							c.Routes(
								dogma.HandlesEvent[*EventStub[TypeA]](),
							)
						},
						HandleEventFunc: func(
							_ context.Context,
							s dogma.ProjectionEventScope,
							_ dogma.Event,
						) (uint64, error) {
							called = true
							return s.Offset() + 1, nil
						},
					}),
				)
			},
		}

		Begin(&testingmock.T{}, app).
			EnableHandlersLike(`^\<proj`).
			Prepare(RecordEvent(EventA1))

		if !called {
			t.Fatal("expected handler to be called")
		}
	})

	t.Run("it panics if there are no matching handlers", func(t *testing.T) {
		app := &ApplicationStub{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "7d5b218d-d69b-48d5-8831-2af77561ee62")
			},
		}

		xtesting.ExpectPanic(
			t,
			`the "<app>" application does not have any handlers with names that match the regular expression (^\<proj), or all such handlers have been disabled by a call to ProjectionConfigurer.Disable()`,
			func() {
				Begin(&testingmock.T{}, app).
					EnableHandlersLike(`^\<proj`)
			},
		)
	})

	t.Run("it does not enable handlers that are disabled by their own configuration", func(t *testing.T) {
		app := &ApplicationStub{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "7d5b218d-d69b-48d5-8831-2af77561ee62")
				c.Routes(
					dogma.ViaProjection(&ProjectionMessageHandlerStub{
						ConfigureFunc: func(c dogma.ProjectionConfigurer) {
							c.Identity("<projection>", "fb5f05c0-589c-4d64-9599-a4875b5a3569")
							c.Routes(
								dogma.HandlesEvent[*EventStub[TypeA]](),
							)
							c.Disable()
						},
					}),
				)
			},
		}

		xtesting.ExpectPanic(
			t,
			`the "<app>" application does not have any handlers with names that match the regular expression (^\<proj), or all such handlers have been disabled by a call to ProjectionConfigurer.Disable()`,
			func() {
				Begin(&testingmock.T{}, app).
					EnableHandlersLike(`^\<proj`)
			},
		)
	})
}

func TestTest_DisableHandlers(t *testing.T) {
	t.Run("it disables the handler", func(t *testing.T) {
		app := &ApplicationStub{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "e79bcae1-8b9a-4755-a15a-dd56f2bb2fdb")
				c.Routes(
					dogma.ViaAggregate(&AggregateMessageHandlerStub{
						ConfigureFunc: func(c dogma.AggregateConfigurer) {
							c.Identity("<aggregate>", "524f7944-a252-48e0-864b-503a903067c2")
							c.Routes(
								dogma.HandlesCommand[*CommandStub[TypeA]](),
								dogma.RecordsEvent[*EventStub[TypeA]](),
							)
						},
						RouteCommandToInstanceFunc: func(dogma.Command) string {
							return "<instance>"
						},
						HandleCommandFunc: func(
							dogma.AggregateRoot,
							dogma.AggregateCommandScope,
							dogma.Command,
						) {
							t.Fatal("unexpected call")
						},
					}),
				)
			},
		}

		Begin(&testingmock.T{}, app).
			DisableHandlers("<aggregate>").
			Prepare(ExecuteCommand(CommandA1))
	})

	t.Run("it panics if the handler is not recognized", func(t *testing.T) {
		app := &ApplicationStub{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "7d5b218d-d69b-48d5-8831-2af77561ee62")
			},
		}

		xtesting.ExpectPanic(
			t,
			`the "<app>" application does not have a handler named "<projection>"`,
			func() {
				Begin(&testingmock.T{}, app).
					DisableHandlers("<projection>")
			},
		)
	})
}

func TestTest_DisableHandlersLike(t *testing.T) {
	t.Run("it disables the handlers with matching names", func(t *testing.T) {
		app := &ApplicationStub{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "e79bcae1-8b9a-4755-a15a-dd56f2bb2fdb")
				c.Routes(
					dogma.ViaAggregate(&AggregateMessageHandlerStub{
						ConfigureFunc: func(c dogma.AggregateConfigurer) {
							c.Identity("<aggregate>", "524f7944-a252-48e0-864b-503a903067c2")
							c.Routes(
								dogma.HandlesCommand[*CommandStub[TypeA]](),
								dogma.RecordsEvent[*EventStub[TypeA]](),
							)
						},
						RouteCommandToInstanceFunc: func(dogma.Command) string {
							return "<instance>"
						},
						HandleCommandFunc: func(
							dogma.AggregateRoot,
							dogma.AggregateCommandScope,
							dogma.Command,
						) {
							t.Fatal("unexpected call")
						},
					}),
				)
			},
		}

		Begin(&testingmock.T{}, app).
			DisableHandlersLike(`^\<agg`).
			Prepare(ExecuteCommand(CommandA1))
	})

	t.Run("it panics if there are no matching handlers", func(t *testing.T) {
		app := &ApplicationStub{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "7d5b218d-d69b-48d5-8831-2af77561ee62")
			},
		}

		xtesting.ExpectPanic(
			t,
			`the "<app>" application does not have any handlers with names that match the regular expression (^\<proj), or all such handlers have been disabled by a call to ProjectionConfigurer.Disable()`,
			func() {
				Begin(&testingmock.T{}, app).
					DisableHandlersLike(`^\<proj`)
			},
		)
	})
}

func TestTest_Annotate(t *testing.T) {
	t.Run("it includes annotations in diffs", func(t *testing.T) {
		app := &ApplicationStub{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "8ec6465c-d4e3-411c-a05b-898a4b608284")

				c.Routes(
					dogma.ViaAggregate(&AggregateMessageHandlerStub{
						ConfigureFunc: func(c dogma.AggregateConfigurer) {
							c.Identity("<aggregate>", "a9cdc28d-ec85-4130-af86-4a2ae86a43dd")
							c.Routes(
								dogma.HandlesCommand[*CommandStub[TypeA]](),
								dogma.RecordsEvent[*EventStub[TypeA]](),
							)
						},
						RouteCommandToInstanceFunc: func(dogma.Command) string {
							return "<instance>"
						},
						HandleCommandFunc: func(
							_ dogma.AggregateRoot,
							s dogma.AggregateCommandScope,
							m dogma.Command,
						) {
							s.RecordEvent(EventA1)
						},
					}),
				)
			},
		}

		mt := &testingmock.T{FailSilently: true}

		Begin(mt, app).
			Annotate(TypeA("A1"), "anna's customer ID").
			Annotate(TypeA("A2"), "bob's customer ID").
			Expect(
				ExecuteCommand(CommandA1),
				ToRecordEvent(EventA2),
			)

		expectReport(
			`✗ record a specific '*stubs.EventStub[TypeA]' event`,
			``,
			`  | EXPLANATION`,
			`  |     a similar event was recorded by the '<aggregate>' aggregate message handler`,
			`  | `,
			`  | SUGGESTIONS`,
			`  |     • check the content of the message`,
			`  | `,
			`  | MESSAGE DIFF`,
			`  |     *stubs.EventStub[github.com/dogmatiq/enginekit/enginetest/stubs.TypeA]{`,
			`  |         Content:         "A[-2-]{+1+}" <<[-bob-]{+anna+}'s customer ID>>`,
			`  |         ValidationError: ""`,
			`  |     }`,
		)(mt)
	})
}
