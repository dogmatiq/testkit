package engine_test

import (
	"context"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogma/fixtures"
	. "github.com/dogmatiq/testkit/engine"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("type CommandExecutor", func() {
	var (
		aggregate *AggregateMessageHandler
		app       *Application
		engine    *Engine
		executor  *CommandExecutor
	)

	BeforeEach(func() {
		aggregate = &AggregateMessageHandler{
			ConfigureFunc: func(c dogma.AggregateConfigurer) {
				c.Identity("<aggregate>", "<aggregate-key>")
				c.ConsumesCommandType(MessageC{})
				c.ProducesEventType(MessageE{})
			},
			RouteCommandToInstanceFunc: func(dogma.Message) string {
				return "<instance>"
			},
		}

		app = &Application{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "<app-key>")
				c.RegisterAggregate(aggregate)
			},
		}

		var err error
		engine, err = New(app)
		Expect(err).ShouldNot(HaveOccurred())

		executor = &CommandExecutor{
			Engine: engine,
		}
	})

	Describe("func ExecuteCommand()", func() {
		It("dispatches to the engine", func() {
			called := false
			aggregate.HandleCommandFunc = func(
				_ dogma.AggregateRoot,
				_ dogma.AggregateCommandScope,
				m dogma.Message,
			) {
				called = true
				Expect(m).To(Equal(MessageC1))
			}

			err := executor.ExecuteCommand(context.Background(), MessageC1)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(called).To(BeTrue())
		})
	})
})
