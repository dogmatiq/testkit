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
	})
})
