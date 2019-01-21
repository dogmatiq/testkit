package config_test

import (
	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogmatest/internal/enginekit/config"
	"github.com/dogmatiq/dogmatest/internal/enginekit/fixtures"
	handlerkit "github.com/dogmatiq/dogmatest/internal/enginekit/handler"
	"github.com/dogmatiq/dogmatest/internal/enginekit/message"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ HandlerConfig = &IntegrationConfig{}

var _ = Describe("type IntegrationConfig", func() {
	Describe("func NewIntegrationConfig", func() {
		var handler *fixtures.IntegrationMessageHandler

		BeforeEach(func() {
			handler = &fixtures.IntegrationMessageHandler{
				ConfigureFunc: func(c dogma.IntegrationConfigurer) {
					c.Name("<name>")
					c.RouteCommandType(fixtures.MessageA{})
					c.RouteCommandType(fixtures.MessageB{})
				},
			}
		})

		When("the configuration is valid", func() {
			var cfg *IntegrationConfig

			BeforeEach(func() {
				var err error
				cfg, err = NewIntegrationConfig(handler)
				Expect(err).ShouldNot(HaveOccurred())
			})

			It("the handler name is set", func() {
				Expect(cfg.HandlerName).To(Equal("<name>"))
			})

			It("the message types are in the set", func() {
				Expect(cfg.MessageTypes).To(Equal(
					message.NewTypeSet(
						fixtures.MessageAType,
						fixtures.MessageBType,
					),
				))
			})

			Describe("func Name()", func() {
				It("returns the handler name", func() {
					Expect(cfg.Name()).To(Equal("<name>"))
				})
			})

			Describe("func HandlerType()", func() {
				It("returns handler.IntegrationType", func() {
					Expect(cfg.HandlerType()).To(Equal(handlerkit.IntegrationType))
				})
			})
		})

		When("the handler does not configure anything", func() {
			BeforeEach(func() {
				handler.ConfigureFunc = nil
			})

			It("returns an error", func() {
				_, err := NewIntegrationConfig(handler)
				Expect(err).Should(HaveOccurred())
			})
		})

		When("the handler does not configure a name", func() {
			BeforeEach(func() {
				handler.ConfigureFunc = func(c dogma.IntegrationConfigurer) {
					c.RouteCommandType(fixtures.MessageA{})
				}
			})

			It("returns a descriptive error", func() {
				_, err := NewIntegrationConfig(handler)

				Expect(err).To(Equal(
					Error(
						"*fixtures.IntegrationMessageHandler.Configure() did not call IntegrationConfigurer.Name()",
					),
				))
			})
		})

		When("the handler configures multiple names", func() {
			BeforeEach(func() {
				handler.ConfigureFunc = func(c dogma.IntegrationConfigurer) {
					c.Name("<name>")
					c.Name("<other>")
					c.RouteCommandType(fixtures.MessageA{})
				}
			})

			It("returns a descriptive error", func() {
				_, err := NewIntegrationConfig(handler)

				Expect(err).To(Equal(
					Error(
						`*fixtures.IntegrationMessageHandler.Configure() has already called IntegrationConfigurer.Name("<name>")`,
					),
				))
			})
		})

		When("the handler configures an invalid name", func() {
			BeforeEach(func() {
				handler.ConfigureFunc = func(c dogma.IntegrationConfigurer) {
					c.Name("\t \n")
					c.RouteCommandType(fixtures.MessageA{})
				}
			})

			It("returns a descriptive error", func() {
				_, err := NewIntegrationConfig(handler)

				Expect(err).To(Equal(
					Error(
						`*fixtures.IntegrationMessageHandler.Configure() called IntegrationConfigurer.Name("\t \n") with an invalid name`,
					),
				))
			})
		})

		When("the handler does not configure any routes", func() {
			BeforeEach(func() {
				handler.ConfigureFunc = func(c dogma.IntegrationConfigurer) {
					c.Name("<name>")
				}
			})

			It("returns a descriptive error", func() {
				_, err := NewIntegrationConfig(handler)

				Expect(err).To(Equal(
					Error(
						"*fixtures.IntegrationMessageHandler.Configure() did not call IntegrationConfigurer.RouteCommandType()",
					),
				))
			})
		})

		When("the handler does configures multiple routes for the same message type", func() {
			BeforeEach(func() {
				handler.ConfigureFunc = func(c dogma.IntegrationConfigurer) {
					c.Name("<name>")
					c.RouteCommandType(fixtures.MessageA{})
					c.RouteCommandType(fixtures.MessageA{})
				}
			})

			It("returns a descriptive error", func() {
				_, err := NewIntegrationConfig(handler)

				Expect(err).To(Equal(
					Error(
						"*fixtures.IntegrationMessageHandler.Configure() has already called IntegrationConfigurer.RouteCommandType(fixtures.MessageA)",
					),
				))
			})
		})
	})
})
