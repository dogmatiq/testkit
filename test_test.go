package testkit_test

import (
	"context"
	"errors"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogma/fixtures"
	. "github.com/dogmatiq/testkit"
	"github.com/dogmatiq/testkit/internal/testingmock"
	g "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = g.Describe("type Test", func() {
	g.Describe("func Prepare()", func() {
		g.It("fails the test if the action returns an error", func() {
			app := &Application{
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
			app := &Application{
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
			app := &Application{
				ConfigureFunc: func(c dogma.ApplicationConfigurer) {
					c.Identity("<app>", "7d5b218d-d69b-48d5-8831-2af77561ee62")
					c.RegisterProjection(&ProjectionMessageHandler{
						ConfigureFunc: func(c dogma.ProjectionConfigurer) {
							c.Identity("<projection>", "fb5f05c0-589c-4d64-9599-a4875b5a3569")
							c.ConsumesEventType(MessageE{})
						},
						HandleEventFunc: func(
							_ context.Context,
							_, _, _ []byte,
							_ dogma.ProjectionEventScope,
							_ dogma.Message,
						) (bool, error) {
							called = true
							return true, nil
						},
					})
				},
			}

			Begin(&testingmock.T{}, app).
				EnableHandlers("<projection>").
				Prepare(RecordEvent(MessageE1))

			Expect(called).To(BeTrue())
		})
	})

	g.Describe("func DisableHandlers()", func() {
		g.It("disables the handler", func() {
			app := &Application{
				ConfigureFunc: func(c dogma.ApplicationConfigurer) {
					c.Identity("<app>", "e79bcae1-8b9a-4755-a15a-dd56f2bb2fdb")
					c.RegisterAggregate(&AggregateMessageHandler{
						ConfigureFunc: func(c dogma.AggregateConfigurer) {
							c.Identity("<aggregate>", "524f7944-a252-48e0-864b-503a903067c2")
							c.ConsumesCommandType(MessageC{})
							c.ProducesEventType(MessageE{})
						},
						RouteCommandToInstanceFunc: func(dogma.Message) string {
							return "<instance>"
						},
						HandleCommandFunc: func(
							dogma.AggregateRoot,
							dogma.AggregateCommandScope,
							dogma.Message,
						) {
							g.Fail("unexpected call")
						},
					})
				},
			}

			Begin(&testingmock.T{}, app).
				DisableHandlers("<aggregate>").
				Prepare(ExecuteCommand(MessageC1))
		})
	})
})
