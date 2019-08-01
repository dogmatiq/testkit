package engine_test

import (
	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/enginekit/fixtures"
	. "github.com/dogmatiq/testkit/engine"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("type Engine", func() {
	Describe("func New", func() {
		var (
			aggregate   *fixtures.AggregateMessageHandler
			process     *fixtures.ProcessMessageHandler
			integration *fixtures.IntegrationMessageHandler
			projection  *fixtures.ProjectionMessageHandler
			app         *fixtures.Application
		)

		BeforeEach(func() {
			aggregate = &fixtures.AggregateMessageHandler{
				ConfigureFunc: func(c dogma.AggregateConfigurer) {
					c.Identity("<aggregate>", "<aggregate-key>")
					c.ConsumesCommandType(fixtures.MessageA{})
					c.ProducesEventType(fixtures.MessageE{})
				},
			}

			process = &fixtures.ProcessMessageHandler{
				ConfigureFunc: func(c dogma.ProcessConfigurer) {
					c.Identity("<process>", "<process-key>")
					c.ConsumesEventType(fixtures.MessageB{})
					c.ConsumesEventType(fixtures.MessageE{}) // shared with <projection>
					c.ProducesCommandType(fixtures.MessageC{})
				},
			}

			integration = &fixtures.IntegrationMessageHandler{
				ConfigureFunc: func(c dogma.IntegrationConfigurer) {
					c.Identity("<integration>", "<integration-key>")
					c.ConsumesCommandType(fixtures.MessageC{})
					c.ProducesEventType(fixtures.MessageF{})
				},
			}

			projection = &fixtures.ProjectionMessageHandler{
				ConfigureFunc: func(c dogma.ProjectionConfigurer) {
					c.Identity("<projection>", "<projection-key>")
					c.ConsumesEventType(fixtures.MessageD{})
					c.ConsumesEventType(fixtures.MessageE{}) // shared with <process>
				},
			}

			app = &fixtures.Application{
				ConfigureFunc: func(c dogma.ApplicationConfigurer) {
					c.Identity("<app>", "<app-key>")
					c.RegisterAggregate(aggregate)
					c.RegisterProcess(process)
					c.RegisterIntegration(integration)
					c.RegisterProjection(projection)
				},
			}
		})

		When("the configuration is valid", func() {
			It("does not return an error", func() {
				_, err := New(app)
				Expect(err).ShouldNot(HaveOccurred())
			})
		})
	})
})
