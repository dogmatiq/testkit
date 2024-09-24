package testkit_test

import (
	"context"
	"errors"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/enginekit/enginetest/stubs"
	. "github.com/dogmatiq/testkit"
	"github.com/dogmatiq/testkit/internal/testingmock"
	g "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = g.Describe("type Test", func() {
	g.Describe("func Prepare()", func() {
		g.It("fails the test if the action returns an error", func() {
			app := &ApplicationStub{
				ConfigureFunc: func(c dogma.ApplicationConfigurer) {
					c.Identity("<app>", "c654acb2-3e87-493a-8b9b-f662cd5e0f55")
				},
			}

			t := &testingmock.T{FailSilently: true}

			Begin(t, app).
				Prepare(
					noopAction{errors.New("<error>")},
				)

			Expect(t.Failed()).To(BeTrue())
		})
	})

	g.Describe("func Expect()", func() {
		g.It("fails the test if the action returns an error", func() {
			app := &ApplicationStub{
				ConfigureFunc: func(c dogma.ApplicationConfigurer) {
					c.Identity("<app>", "c691b2ca-4c07-4473-bc42-060266cc7a56")
				},
			}

			t := &testingmock.T{FailSilently: true}

			Begin(t, app).
				Expect(
					noopAction{errors.New("<error>")},
					pass,
				)

			Expect(t.Failed()).To(BeTrue())
		})
	})

	g.Describe("func EnableHandlers()", func() {
		g.It("enables the handler", func() {
			called := false
			app := &ApplicationStub{
				ConfigureFunc: func(c dogma.ApplicationConfigurer) {
					c.Identity("<app>", "7d5b218d-d69b-48d5-8831-2af77561ee62")
					c.RegisterProjection(&ProjectionMessageHandlerStub{
						ConfigureFunc: func(c dogma.ProjectionConfigurer) {
							c.Identity("<projection>", "fb5f05c0-589c-4d64-9599-a4875b5a3569")
							c.Routes(
								dogma.HandlesEvent[EventStub[TypeA]](),
							)
						},
						HandleEventFunc: func(
							_ context.Context,
							_, _, _ []byte,
							_ dogma.ProjectionEventScope,
							_ dogma.Event,
						) (bool, error) {
							called = true
							return true, nil
						},
					})
				},
			}

			Begin(&testingmock.T{}, app).
				EnableHandlers("<projection>").
				Prepare(RecordEvent(EventA1))

			Expect(called).To(BeTrue())
		})

		g.It("panics if the handler is not recognized", func() {
			app := &ApplicationStub{
				ConfigureFunc: func(c dogma.ApplicationConfigurer) {
					c.Identity("<app>", "7d5b218d-d69b-48d5-8831-2af77561ee62")
				},
			}

			Expect(func() {
				Begin(&testingmock.T{}, app).
					EnableHandlers("<projection>")
			}).To(PanicWith(`the "<app>" application does not have a handler named "<projection>"`))
		})

		g.It("panics if the handler is disabled by its own configuration", func() {
			app := &ApplicationStub{
				ConfigureFunc: func(c dogma.ApplicationConfigurer) {
					c.Identity("<app>", "7d5b218d-d69b-48d5-8831-2af77561ee62")
					c.RegisterProjection(&ProjectionMessageHandlerStub{
						ConfigureFunc: func(c dogma.ProjectionConfigurer) {
							c.Identity("<projection>", "fb5f05c0-589c-4d64-9599-a4875b5a3569")
							c.Routes(
								dogma.HandlesEvent[EventStub[TypeA]](),
							)
							c.Disable()
						},
					})
				},
			}

			Expect(func() {
				Begin(&testingmock.T{}, app).
					EnableHandlers("<projection>")
			}).To(PanicWith(`cannot enable the "<projection>" handler, it has been disabled by a call to ProjectionConfigurer.Disable()`))
		})
	})

	g.Describe("func EnableHandlersLike()", func() {
		g.It("enables handlers with matching names", func() {
			called := false
			app := &ApplicationStub{
				ConfigureFunc: func(c dogma.ApplicationConfigurer) {
					c.Identity("<app>", "7d5b218d-d69b-48d5-8831-2af77561ee62")
					c.RegisterProjection(&ProjectionMessageHandlerStub{
						ConfigureFunc: func(c dogma.ProjectionConfigurer) {
							c.Identity("<projection>", "fb5f05c0-589c-4d64-9599-a4875b5a3569")
							c.Routes(
								dogma.HandlesEvent[EventStub[TypeA]](),
							)
						},
						HandleEventFunc: func(
							_ context.Context,
							_, _, _ []byte,
							_ dogma.ProjectionEventScope,
							_ dogma.Event,
						) (bool, error) {
							called = true
							return true, nil
						},
					})
				},
			}

			Begin(&testingmock.T{}, app).
				EnableHandlersLike(`^\<proj`).
				Prepare(RecordEvent(EventA1))

			Expect(called).To(BeTrue())
		})

		g.It("panics if there are no matching handlers", func() {
			app := &ApplicationStub{
				ConfigureFunc: func(c dogma.ApplicationConfigurer) {
					c.Identity("<app>", "7d5b218d-d69b-48d5-8831-2af77561ee62")
				},
			}

			Expect(func() {
				Begin(&testingmock.T{}, app).
					EnableHandlersLike(`^\<proj`)
			}).To(PanicWith(`the "<app>" application does not have any handlers with names that match the regular expression (^\<proj), or all such handlers have been disabled by a call to ProjectionConfigurer.Disable()`))
		})

		g.It("does not enable handlers that are disabled by their own configuration", func() {
			app := &ApplicationStub{
				ConfigureFunc: func(c dogma.ApplicationConfigurer) {
					c.Identity("<app>", "7d5b218d-d69b-48d5-8831-2af77561ee62")
					c.RegisterProjection(&ProjectionMessageHandlerStub{
						ConfigureFunc: func(c dogma.ProjectionConfigurer) {
							c.Identity("<projection>", "fb5f05c0-589c-4d64-9599-a4875b5a3569")
							c.Routes(
								dogma.HandlesEvent[EventStub[TypeA]](),
							)
							c.Disable()
						},
					})
				},
			}

			Expect(func() {
				Begin(&testingmock.T{}, app).
					EnableHandlersLike(`^\<proj`)
			}).To(PanicWith(`the "<app>" application does not have any handlers with names that match the regular expression (^\<proj), or all such handlers have been disabled by a call to ProjectionConfigurer.Disable()`))
		})
	})

	g.Describe("func DisableHandlers()", func() {
		g.It("disables the handler", func() {
			app := &ApplicationStub{
				ConfigureFunc: func(c dogma.ApplicationConfigurer) {
					c.Identity("<app>", "e79bcae1-8b9a-4755-a15a-dd56f2bb2fdb")
					c.RegisterAggregate(&AggregateMessageHandlerStub{
						ConfigureFunc: func(c dogma.AggregateConfigurer) {
							c.Identity("<aggregate>", "524f7944-a252-48e0-864b-503a903067c2")
							c.Routes(
								dogma.HandlesCommand[CommandStub[TypeA]](),
								dogma.RecordsEvent[EventStub[TypeA]](),
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
							g.Fail("unexpected call")
						},
					})
				},
			}

			Begin(&testingmock.T{}, app).
				DisableHandlers("<aggregate>").
				Prepare(ExecuteCommand(CommandA1))
		})

		g.It("panics if the handler is not recognized", func() {
			app := &ApplicationStub{
				ConfigureFunc: func(c dogma.ApplicationConfigurer) {
					c.Identity("<app>", "7d5b218d-d69b-48d5-8831-2af77561ee62")
				},
			}

			Expect(func() {
				Begin(&testingmock.T{}, app).
					DisableHandlers("<projection>")
			}).To(PanicWith(`the "<app>" application does not have a handler named "<projection>"`))
		})
	})

	g.Describe("func DisableHandlersLike()", func() {
		g.It("disables the handlers with matching names", func() {
			app := &ApplicationStub{
				ConfigureFunc: func(c dogma.ApplicationConfigurer) {
					c.Identity("<app>", "e79bcae1-8b9a-4755-a15a-dd56f2bb2fdb")
					c.RegisterAggregate(&AggregateMessageHandlerStub{
						ConfigureFunc: func(c dogma.AggregateConfigurer) {
							c.Identity("<aggregate>", "524f7944-a252-48e0-864b-503a903067c2")
							c.Routes(
								dogma.HandlesCommand[CommandStub[TypeA]](),
								dogma.RecordsEvent[EventStub[TypeA]](),
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
							g.Fail("unexpected call")
						},
					})
				},
			}

			Begin(&testingmock.T{}, app).
				DisableHandlersLike(`^\<agg`).
				Prepare(ExecuteCommand(CommandA1))
		})

		g.It("panics if there are no matching handlers", func() {
			app := &ApplicationStub{
				ConfigureFunc: func(c dogma.ApplicationConfigurer) {
					c.Identity("<app>", "7d5b218d-d69b-48d5-8831-2af77561ee62")
				},
			}

			Expect(func() {
				Begin(&testingmock.T{}, app).
					DisableHandlersLike(`^\<proj`)
			}).To(PanicWith(`the "<app>" application does not have any handlers with names that match the regular expression (^\<proj), or all such handlers have been disabled by a call to ProjectionConfigurer.Disable()`))
		})
	})

	g.Describe("func Annotate()", func() {
		g.It("includes annotations in diffs", func() {
			app := &ApplicationStub{
				ConfigureFunc: func(c dogma.ApplicationConfigurer) {
					c.Identity("<app>", "8ec6465c-d4e3-411c-a05b-898a4b608284")

					c.RegisterAggregate(&AggregateMessageHandlerStub{
						ConfigureFunc: func(c dogma.AggregateConfigurer) {
							c.Identity("<aggregate>", "a9cdc28d-ec85-4130-af86-4a2ae86a43dd")
							c.Routes(
								dogma.HandlesCommand[CommandStub[TypeA]](),
								dogma.RecordsEvent[EventStub[TypeA]](),
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
					})
				},
			}

			t := &testingmock.T{FailSilently: true}

			Begin(t, app).
				Annotate(TypeA("A1"), "anna's customer ID").
				Annotate(TypeA("A2"), "bob's customer ID").
				Expect(
					ExecuteCommand(CommandA1),
					ToRecordEvent(EventA2),
				)

			expectReport(
				`✗ record a specific 'stubs.EventStub[TypeA]' event`,
				``,
				`  | EXPLANATION`,
				`  |     a similar event was recorded by the '<aggregate>' aggregate message handler`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • check the content of the message`,
				`  | `,
				`  | MESSAGE DIFF`,
				`  |     stubs.EventStub[github.com/dogmatiq/enginekit/enginetest/stubs.TypeA]{`,
				`  |         Content:         "A[-2-]{+1+}" <<[-bob-]{+anna+}'s customer ID>>`,
				`  |         ValidationError: ""`,
				`  |     }`,
			)(t)
		})
	})
})
