package engine_test

import (
	"context"

	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogma/fixtures"
	. "github.com/dogmatiq/enginekit/enginetest/stubs"
	. "github.com/dogmatiq/testkit/engine"
	g "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = g.Describe("type CommandExecutor", func() {
	var (
		aggregate *AggregateMessageHandlerStub
		app       *ApplicationStub
		engine    *Engine
		executor  *CommandExecutor
	)

	g.BeforeEach(func() {
		aggregate = &AggregateMessageHandlerStub{
			ConfigureFunc: func(c dogma.AggregateConfigurer) {
				c.Identity("<aggregate>", "4acf3050-8d02-4052-a9af-abb9e67add78")
				c.Routes(
					dogma.HandlesCommand[MessageC](),
					dogma.RecordsEvent[MessageE](),
				)
			},
			RouteCommandToInstanceFunc: func(dogma.Command) string {
				return "<instance>"
			},
		}

		app = &ApplicationStub{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "d905114d-b026-4f1a-9bc6-3abd86058e2d")
				c.RegisterAggregate(aggregate)
			},
		}

		engine = MustNew(
			configkit.FromApplication(app),
		)

		executor = &CommandExecutor{
			Engine: engine,
		}
	})

	g.Describe("func ExecuteCommand()", func() {
		g.It("dispatches to the engine", func() {
			called := false
			aggregate.HandleCommandFunc = func(
				_ dogma.AggregateRoot,
				_ dogma.AggregateCommandScope,
				m dogma.Command,
			) {
				called = true
				Expect(m).To(Equal(MessageC1))
			}

			err := executor.ExecuteCommand(context.Background(), MessageC1)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(called).To(BeTrue())
		})

		g.It("panics if the message is not a command", func() {
			Expect(func() {
				executor.ExecuteCommand(context.Background(), MessageE1)
			}).To(PanicWith("cannot execute command, fixtures.MessageE is configured as an event"))
		})

		g.It("panics if the message is unrecognized", func() {
			Expect(func() {
				executor.ExecuteCommand(context.Background(), MessageX1)
			}).To(PanicWith("cannot execute command, fixtures.MessageX is a not a recognized message type"))
		})
	})
})
