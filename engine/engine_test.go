package engine_test

import (
	"context"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogma/fixtures"
	. "github.com/dogmatiq/testkit/engine"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("type Engine", func() {
	var (
		aggregate   *AggregateMessageHandler
		process     *ProcessMessageHandler
		integration *IntegrationMessageHandler
		projection  *ProjectionMessageHandler
		app         *Application
		engine      *Engine
	)

	BeforeEach(func() {
		aggregate = &AggregateMessageHandler{
			ConfigureFunc: func(c dogma.AggregateConfigurer) {
				c.Identity("<aggregate>", "<aggregate-key>")
				c.ConsumesCommandType(MessageA{})
				c.ProducesEventType(MessageE{})
			},
			RouteCommandToInstanceFunc: func(dogma.Message) string {
				return "<instance>"
			},
		}

		process = &ProcessMessageHandler{
			ConfigureFunc: func(c dogma.ProcessConfigurer) {
				c.Identity("<process>", "<process-key>")
				c.ConsumesEventType(MessageB{})
				c.ConsumesEventType(MessageE{}) // shared with <projection>
				c.ProducesCommandType(MessageC{})
			},
			RouteEventToInstanceFunc: func(context.Context, dogma.Message) (string, bool, error) {
				return "<instance>", true, nil
			},
		}

		integration = &IntegrationMessageHandler{
			ConfigureFunc: func(c dogma.IntegrationConfigurer) {
				c.Identity("<integration>", "<integration-key>")
				c.ConsumesCommandType(MessageC{})
				c.ProducesEventType(MessageF{})
			},
		}

		projection = &ProjectionMessageHandler{
			ConfigureFunc: func(c dogma.ProjectionConfigurer) {
				c.Identity("<projection>", "<projection-key>")
				c.ConsumesEventType(MessageD{})
				c.ConsumesEventType(MessageE{}) // shared with <process>
			},
		}

		app = &Application{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "<app-key>")
				c.RegisterAggregate(aggregate)
				c.RegisterProcess(process)
				c.RegisterIntegration(integration)
				c.RegisterProjection(projection)
			},
		}

		var err error
		engine, err = New(app)
		Expect(err).ShouldNot(HaveOccurred())
	})

	Describe("func Dispatch()", func() {
		It("panics if the message type is unrecognized", func() {
			Expect(func() {
				engine.Dispatch(context.Background(), MessageX1)
			}).To(PanicWith("the fixtures.MessageX message type is not consumed by any handlers"))
		})
	})
})
