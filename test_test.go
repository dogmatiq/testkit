package testkit_test

import (
	"context"
	"errors"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogma/fixtures"
	. "github.com/dogmatiq/testkit"
	"github.com/dogmatiq/testkit/internal/testingmock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("type Test", func() {
	Describe("func Prepare()", func() {
		It("fails the test if the action returns an error", func() {
			app := &Application{
				ConfigureFunc: func(c dogma.ApplicationConfigurer) {
					c.Identity("<app>", "<app-key>")
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

	Describe("func Expect()", func() {
		It("fails the test if the action returns an error", func() {
			app := &Application{
				ConfigureFunc: func(c dogma.ApplicationConfigurer) {
					c.Identity("<app>", "<app-key>")
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

	Describe("func EnableHandlers()", func() {
		It("enables the handler", func() {
			called := false
			app := &Application{
				ConfigureFunc: func(c dogma.ApplicationConfigurer) {
					c.Identity("<app>", "<app-key>")
					c.RegisterProjection(&ProjectionMessageHandler{
						ConfigureFunc: func(c dogma.ProjectionConfigurer) {
							c.Identity("<projection>", "<projection-key>")
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

	Describe("func DisableHandlers()", func() {
		It("disables the handler", func() {
			app := &Application{
				ConfigureFunc: func(c dogma.ApplicationConfigurer) {
					c.Identity("<app>", "<app-key>")
					c.RegisterAggregate(&AggregateMessageHandler{
						ConfigureFunc: func(c dogma.AggregateConfigurer) {
							c.Identity("<aggregate>", "<aggregate-key>")
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
							Fail("unexpected call")
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
