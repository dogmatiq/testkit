package config_test

import (
	"reflect"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogmatest/engine/config"
	"github.com/dogmatiq/dogmatest/internal/fixtures"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ Config = &AppConfig{}

var _ = Describe("type AppConfig", func() {
	Describe("func NewAppConfig", func() {
		var (
			aggregate   *fixtures.AggregateMessageHandler
			process     *fixtures.ProcessMessageHandler
			integration *fixtures.IntegrationMessageHandler
			projection  *fixtures.ProjectionMessageHandler
			app         dogma.App
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
				},
			}

			app = dogma.App{
				Name:         "<app>",
				Aggregates:   []dogma.AggregateMessageHandler{aggregate},
				Processes:    []dogma.ProcessMessageHandler{process},
				Integrations: []dogma.IntegrationMessageHandler{integration},
				Projections:  []dogma.ProjectionMessageHandler{projection},
			}
		})

		When("the configuration is valid", func() {
			var cfg *AppConfig

			BeforeEach(func() {
				var err error
				cfg, err = NewAppConfig(app)
				Expect(err).ShouldNot(HaveOccurred())
			})

			It("the app name is set", func() {
				Expect(cfg.AppName).To(Equal("<app>"))
			})

			It("the routes are present", func() {
				Expect(cfg.Routes).To(Equal(
					map[reflect.Type][]string{
						reflect.TypeOf(fixtures.MessageA{}): {"<aggregate>"},
						reflect.TypeOf(fixtures.MessageB{}): {"<process>"},
						reflect.TypeOf(fixtures.MessageC{}): {"<integration>"},
						reflect.TypeOf(fixtures.MessageD{}): {"<projection>"},
					},
				))
			})

			It("the command routes are present", func() {
				Expect(cfg.CommandRoutes).To(Equal(
					map[reflect.Type]string{
						reflect.TypeOf(fixtures.MessageA{}): "<aggregate>",
						reflect.TypeOf(fixtures.MessageC{}): "<integration>",
					},
				))
			})

			It("the event routes are present", func() {
				Expect(cfg.EventRoutes).To(Equal(
					map[reflect.Type][]string{
						reflect.TypeOf(fixtures.MessageB{}): {"<process>"},
						reflect.TypeOf(fixtures.MessageD{}): {"<projection>"},
					},
				))
			})

			Describe("func Name()", func() {
				It("returns the app name", func() {
					Expect(cfg.Name()).To(Equal("<app>"))
				})
			})
		})

		When("the app name is invalid", func() {
			BeforeEach(func() {
				app.Name = "\t \n"
			})

			It("returns a descriptive error", func() {
				_, err := NewAppConfig(app)

				Expect(err).To(Equal(
					ConfigurationError(
						`application name "\t \n" is invalid`,
					),
				))
			})
		})

		When("the app contains an invalid handler configurations", func() {
			It("returns an error when an aggregate is misconfigured", func() {
				aggregate.ConfigureFunc = nil

				_, err := NewAppConfig(app)

				Expect(err).Should(HaveOccurred())
			})

			It("returns an error when a process is misconfigured", func() {
				process.ConfigureFunc = nil

				_, err := NewAppConfig(app)

				Expect(err).Should(HaveOccurred())
			})

			It("returns an error when an integration is misconfigured", func() {
				integration.ConfigureFunc = nil

				_, err := NewAppConfig(app)

				Expect(err).Should(HaveOccurred())
			})

			It("returns an error when a projection is misconfigured", func() {
				projection.ConfigureFunc = nil

				_, err := NewAppConfig(app)

				Expect(err).Should(HaveOccurred())
			})
		})

		When("the app contains conflicting handler names", func() {
			It("returns an error when an aggregate name is in conflict", func() {
				// Note that aggregates are processed before everything else, so in order to
				// induce a conflict we need to have two aggregate names in conflict. For
				// the other tests, we will test conflicts across differing handler types.
				app.Aggregates = append(app.Aggregates, aggregate)

				_, err := NewAppConfig(app)

				Expect(err).To(Equal(
					ConfigurationError(
						`*fixtures.AggregateMessageHandler can not use the handler name "<aggregate>", because it is already used by *fixtures.AggregateMessageHandler`,
					),
				))
			})

			It("returns an error when a process name is in conflict", func() {
				process.ConfigureFunc = func(c dogma.ProcessConfigurer) {
					c.Name("<aggregate>") // conflict!
					c.RouteEventType(fixtures.MessageB{})
				}

				_, err := NewAppConfig(app)

				Expect(err).To(Equal(
					ConfigurationError(
						`*fixtures.ProcessMessageHandler can not use the handler name "<aggregate>", because it is already used by *fixtures.AggregateMessageHandler`,
					),
				))
			})

			It("returns an error when an integration name is in conflict", func() {
				integration.ConfigureFunc = func(c dogma.IntegrationConfigurer) {
					c.Name("<process>") // conflict!
					c.RouteCommandType(fixtures.MessageC{})
				}

				_, err := NewAppConfig(app)

				Expect(err).To(Equal(
					ConfigurationError(
						`*fixtures.IntegrationMessageHandler can not use the handler name "<process>", because it is already used by *fixtures.ProcessMessageHandler`,
					),
				))
			})

			It("returns an error when a projection name is in conflict", func() {
				projection.ConfigureFunc = func(c dogma.ProjectionConfigurer) {
					c.Name("<integration>") // conflict!
					c.RouteEventType(fixtures.MessageD{})
				}

				_, err := NewAppConfig(app)

				Expect(err).To(Equal(
					ConfigurationError(
						`*fixtures.ProjectionMessageHandler can not use the handler name "<integration>", because it is already used by *fixtures.IntegrationMessageHandler`,
					),
				))
			})
		})
	})
})
