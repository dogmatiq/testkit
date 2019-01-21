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

var _ HandlerConfig = &ProcessConfig{}

var _ = Describe("type ProcessConfig", func() {
	Describe("func NewProcessConfig", func() {
		var handler *fixtures.ProcessMessageHandler

		BeforeEach(func() {
			handler = &fixtures.ProcessMessageHandler{
				ConfigureFunc: func(c dogma.ProcessConfigurer) {
					c.Name("<name>")
					c.RouteEventType(fixtures.MessageA{})
					c.RouteEventType(fixtures.MessageB{})
				},
			}
		})

		When("the configuration is valid", func() {
			var cfg *ProcessConfig

			BeforeEach(func() {
				var err error
				cfg, err = NewProcessConfig(handler)
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
				It("returns handler.ProcessType", func() {
					Expect(cfg.HandlerType()).To(Equal(handlerkit.ProcessType))
				})
			})
		})

		When("the handler does not configure anything", func() {
			BeforeEach(func() {
				handler.ConfigureFunc = nil
			})

			It("returns an error", func() {
				_, err := NewProcessConfig(handler)
				Expect(err).Should(HaveOccurred())
			})
		})

		When("the handler does not configure a name", func() {
			BeforeEach(func() {
				handler.ConfigureFunc = func(c dogma.ProcessConfigurer) {
					c.RouteEventType(fixtures.MessageA{})
				}
			})

			It("returns a descriptive error", func() {
				_, err := NewProcessConfig(handler)

				Expect(err).To(Equal(
					Error(
						"*fixtures.ProcessMessageHandler.Configure() did not call ProcessConfigurer.Name()",
					),
				))
			})
		})

		When("the handler configures multiple names", func() {
			BeforeEach(func() {
				handler.ConfigureFunc = func(c dogma.ProcessConfigurer) {
					c.Name("<name>")
					c.Name("<other>")
					c.RouteEventType(fixtures.MessageA{})
				}
			})

			It("returns a descriptive error", func() {
				_, err := NewProcessConfig(handler)

				Expect(err).To(Equal(
					Error(
						`*fixtures.ProcessMessageHandler.Configure() has already called ProcessConfigurer.Name("<name>")`,
					),
				))
			})
		})

		When("the handler configures an invalid name", func() {
			BeforeEach(func() {
				handler.ConfigureFunc = func(c dogma.ProcessConfigurer) {
					c.Name("\t \n")
					c.RouteEventType(fixtures.MessageA{})
				}
			})

			It("returns a descriptive error", func() {
				_, err := NewProcessConfig(handler)

				Expect(err).To(Equal(
					Error(
						`*fixtures.ProcessMessageHandler.Configure() called ProcessConfigurer.Name("\t \n") with an invalid name`,
					),
				))
			})
		})

		When("the handler does not configure any routes", func() {
			BeforeEach(func() {
				handler.ConfigureFunc = func(c dogma.ProcessConfigurer) {
					c.Name("<name>")
				}
			})

			It("returns a descriptive error", func() {
				_, err := NewProcessConfig(handler)

				Expect(err).To(Equal(
					Error(
						"*fixtures.ProcessMessageHandler.Configure() did not call ProcessConfigurer.RouteEventType()",
					),
				))
			})
		})

		When("the handler does configures multiple routes for the same message type", func() {
			BeforeEach(func() {
				handler.ConfigureFunc = func(c dogma.ProcessConfigurer) {
					c.Name("<name>")
					c.RouteEventType(fixtures.MessageA{})
					c.RouteEventType(fixtures.MessageA{})
				}
			})

			It("returns a descriptive error", func() {
				_, err := NewProcessConfig(handler)

				Expect(err).To(Equal(
					Error(
						"*fixtures.ProcessMessageHandler.Configure() has already called ProcessConfigurer.RouteEventType(fixtures.MessageA)",
					),
				))
			})
		})
	})
})
