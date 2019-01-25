package config_test

import (
	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogmatest/internal/enginekit/config"
	"github.com/dogmatiq/dogmatest/internal/enginekit/fixtures"
	"github.com/dogmatiq/dogmatest/internal/enginekit/message"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ Config = &ApplicationConfig{}

var _ = Describe("type ApplicationConfig", func() {
	Describe("func NewApplicationConfig", func() {
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
					c.Name("<aggregate>")
					c.RouteCommandType(fixtures.MessageA{})
				},
			}

			process = &fixtures.ProcessMessageHandler{
				ConfigureFunc: func(c dogma.ProcessConfigurer) {
					c.Name("<process>")
					c.RouteEventType(fixtures.MessageB{})
					c.RouteEventType(fixtures.MessageE{}) // shared with <projection>
				},
			}

			integration = &fixtures.IntegrationMessageHandler{
				ConfigureFunc: func(c dogma.IntegrationConfigurer) {
					c.Name("<integration>")
					c.RouteCommandType(fixtures.MessageC{})
				},
			}

			projection = &fixtures.ProjectionMessageHandler{
				ConfigureFunc: func(c dogma.ProjectionConfigurer) {
					c.Name("<projection>")
					c.RouteEventType(fixtures.MessageD{})
					c.RouteEventType(fixtures.MessageE{}) // shared with <process>
				},
			}

			app = &fixtures.Application{
				ConfigureFunc: func(c dogma.ApplicationConfigurer) {
					c.Name("<app>")
					c.RegisterAggregate(aggregate)
					c.RegisterProcess(process)
					c.RegisterIntegration(integration)
					c.RegisterProjection(projection)
				},
			}
		})

		When("the configuration is valid", func() {
			var cfg *ApplicationConfig

			BeforeEach(func() {
				var err error
				cfg, err = NewApplicationConfig(app)
				Expect(err).ShouldNot(HaveOccurred())
			})

			It("the app name is set", func() {
				Expect(cfg.ApplicationName).To(Equal("<app>"))
			})

			It("the routes are present", func() {
				Expect(cfg.Routes).To(Equal(
					map[message.Type][]string{
						fixtures.MessageAType: {"<aggregate>"},
						fixtures.MessageBType: {"<process>"},
						fixtures.MessageCType: {"<integration>"},
						fixtures.MessageDType: {"<projection>"},
						fixtures.MessageEType: {"<process>", "<projection>"},
					},
				))
			})

			It("the command routes are present", func() {
				Expect(cfg.CommandRoutes).To(Equal(
					map[message.Type]string{
						fixtures.MessageAType: "<aggregate>",
						fixtures.MessageCType: "<integration>",
					},
				))
			})

			It("the event routes are present", func() {
				Expect(cfg.EventRoutes).To(Equal(
					map[message.Type][]string{
						fixtures.MessageBType: {"<process>"},
						fixtures.MessageDType: {"<projection>"},
						fixtures.MessageEType: {"<process>", "<projection>"},
					},
				))
			})

			Describe("func Name()", func() {
				It("returns the app name", func() {
					Expect(cfg.Name()).To(Equal("<app>"))
				})
			})
		})

		When("the app does not configure a name", func() {
			BeforeEach(func() {
				app.ConfigureFunc = nil
			})

			It("returns a descriptive error", func() {
				_, err := NewApplicationConfig(app)

				Expect(err).To(Equal(
					Error(
						"*fixtures.Application.Configure() did not call ApplicationConfigurer.Name()",
					),
				))
			})
		})

		When("the app configures multiple names", func() {
			BeforeEach(func() {
				app.ConfigureFunc = func(c dogma.ApplicationConfigurer) {
					c.Name("<name>")
					c.Name("<other>")
				}
			})

			It("returns a descriptive error", func() {
				_, err := NewApplicationConfig(app)

				Expect(err).To(Equal(
					Error(
						`*fixtures.Application.Configure() has already called ApplicationConfigurer.Name("<name>")`,
					),
				))
			})
		})

		When("the app configures an invalid name", func() {
			BeforeEach(func() {
				app.ConfigureFunc = func(c dogma.ApplicationConfigurer) {
					c.Name("\t \n")
				}
			})

			It("returns a descriptive error", func() {
				_, err := NewApplicationConfig(app)

				Expect(err).To(Equal(
					Error(
						`*fixtures.Application.Configure() called ApplicationConfigurer.Name("\t \n") with an invalid name`,
					),
				))
			})
		})

		When("the app contains an invalid handler configurations", func() {
			It("returns an error when an aggregate is misconfigured", func() {
				aggregate.ConfigureFunc = nil

				_, err := NewApplicationConfig(app)

				Expect(err).Should(HaveOccurred())
			})

			It("returns an error when a process is misconfigured", func() {
				process.ConfigureFunc = nil

				_, err := NewApplicationConfig(app)

				Expect(err).Should(HaveOccurred())
			})

			It("returns an error when an integration is misconfigured", func() {
				integration.ConfigureFunc = nil

				_, err := NewApplicationConfig(app)

				Expect(err).Should(HaveOccurred())
			})

			It("returns an error when a projection is misconfigured", func() {
				projection.ConfigureFunc = nil

				_, err := NewApplicationConfig(app)

				Expect(err).Should(HaveOccurred())
			})
		})

		When("the app contains conflicting handler names", func() {
			It("returns an error when an aggregate name is in conflict", func() {
				aggregate.ConfigureFunc = func(c dogma.AggregateConfigurer) {
					c.Name("<process>") // conflict!
					c.RouteCommandType(fixtures.MessageA{})
				}

				app.ConfigureFunc = func(c dogma.ApplicationConfigurer) {
					c.Name("<app>")
					c.RegisterProcess(process)
					c.RegisterAggregate(aggregate) // register the conflicting aggregate last
				}

				_, err := NewApplicationConfig(app)

				Expect(err).To(Equal(
					Error(
						`*fixtures.AggregateMessageHandler can not use the handler name "<process>", because it is already used by *fixtures.ProcessMessageHandler`,
					),
				))
			})

			It("returns an error when a process name is in conflict", func() {
				process.ConfigureFunc = func(c dogma.ProcessConfigurer) {
					c.Name("<aggregate>") // conflict!
					c.RouteEventType(fixtures.MessageB{})
				}

				_, err := NewApplicationConfig(app)

				Expect(err).To(Equal(
					Error(
						`*fixtures.ProcessMessageHandler can not use the handler name "<aggregate>", because it is already used by *fixtures.AggregateMessageHandler`,
					),
				))
			})

			It("returns an error when an integration name is in conflict", func() {
				integration.ConfigureFunc = func(c dogma.IntegrationConfigurer) {
					c.Name("<process>") // conflict!
					c.RouteCommandType(fixtures.MessageC{})
				}

				_, err := NewApplicationConfig(app)

				Expect(err).To(Equal(
					Error(
						`*fixtures.IntegrationMessageHandler can not use the handler name "<process>", because it is already used by *fixtures.ProcessMessageHandler`,
					),
				))
			})

			It("returns an error when a projection name is in conflict", func() {
				projection.ConfigureFunc = func(c dogma.ProjectionConfigurer) {
					c.Name("<integration>") // conflict!
					c.RouteEventType(fixtures.MessageD{})
				}

				_, err := NewApplicationConfig(app)

				Expect(err).To(Equal(
					Error(
						`*fixtures.ProjectionMessageHandler can not use the handler name "<integration>", because it is already used by *fixtures.IntegrationMessageHandler`,
					),
				))
			})
		})

		When("the app contains conflicting routes", func() {
			It("returns an error when a route is in conflict because two handlers intend to receive the same command", func() {
				integration.ConfigureFunc = func(c dogma.IntegrationConfigurer) {
					c.Name("<integration>")
					c.RouteCommandType(fixtures.MessageA{}) // conflict with <aggregate>
				}

				_, err := NewApplicationConfig(app)

				Expect(err).To(Equal(
					Error(
						`can not route commands of type fixtures.MessageA to "<integration>" because they are already routed to "<aggregate>"`,
					),
				))
			})

			It("returns an error when a process route is in conflict because a command is reclassified as an event", func() {
				process.ConfigureFunc = func(c dogma.ProcessConfigurer) {
					c.Name("<process>")
					c.RouteEventType(fixtures.MessageA{}) // conflict with <aggregate>
				}

				_, err := NewApplicationConfig(app)

				Expect(err).To(Equal(
					Error(
						`can not route messages of type fixtures.MessageA to "<process>" as events because they are already routed to "<aggregate>" as commands`,
					),
				))
			})

			It("returns an error when a process route is in conflict because an event with a single handler is reclassified as a command", func() {
				integration.ConfigureFunc = func(c dogma.IntegrationConfigurer) {
					c.Name("<integration>")
					c.RouteCommandType(fixtures.MessageB{}) // conflict with <process>
				}

				_, err := NewApplicationConfig(app)

				Expect(err).To(Equal(
					Error(
						`can not route messages of type fixtures.MessageB to "<integration>" as commands because they are already routed to "<process>" as events`,
					),
				))
			})

			It("returns an error when a process route is in conflict because an event with multiple handlers is reclassified as a command", func() {
				integration.ConfigureFunc = func(c dogma.IntegrationConfigurer) {
					c.Name("<integration>")
					c.RouteCommandType(fixtures.MessageE{}) // conflict with <process> and <projection>
				}

				app.ConfigureFunc = func(c dogma.ApplicationConfigurer) {
					c.Name("<app>")
					c.RegisterProcess(process)
					c.RegisterProjection(projection)
					c.RegisterIntegration(integration) // register the conflicting integration last
				}

				_, err := NewApplicationConfig(app)

				Expect(err).To(Equal(
					Error(
						`can not route messages of type fixtures.MessageE to "<integration>" as commands because they are already routed to "<process>" and 1 other handler(s) as events`,
					),
				))
			})
		})
	})
})
