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

var _ HandlerConfig = &ProjectionConfig{}

var _ = Describe("type ProjectionConfig", func() {
	Describe("func NewProjectionConfig", func() {
		var handler *fixtures.ProjectionMessageHandler

		BeforeEach(func() {
			handler = &fixtures.ProjectionMessageHandler{
				ConfigureFunc: func(c dogma.ProjectionConfigurer) {
					c.Name("<name>")
					c.RouteEventType(fixtures.MessageA{})
					c.RouteEventType(fixtures.MessageB{})
				},
			}
		})

		When("the configuration is valid", func() {
			var cfg *ProjectionConfig

			BeforeEach(func() {
				var err error
				cfg, err = NewProjectionConfig(handler)
				Expect(err).ShouldNot(HaveOccurred())
			})

			It("the handler name is set", func() {
				Expect(cfg.HandlerName).To(Equal("<name>"))
			})

			It("the message types are in the set", func() {
				Expect(cfg.EventTypes).To(Equal(
					map[message.Type]struct{}{
						message.TypeOf(fixtures.MessageA{}): struct{}{},
						message.TypeOf(fixtures.MessageB{}): struct{}{},
					},
				))
			})

			Describe("func Name()", func() {
				It("returns the handler name", func() {
					Expect(cfg.Name()).To(Equal("<name>"))
				})
			})

			Describe("func HandlerType()", func() {
				It("returns handler.ProjectionType", func() {
					Expect(cfg.HandlerType()).To(Equal(handlerkit.ProjectionType))
				})
			})
		})

		When("the handler does not configure anything", func() {
			BeforeEach(func() {
				handler.ConfigureFunc = nil
			})

			It("returns an error", func() {
				_, err := NewProjectionConfig(handler)
				Expect(err).Should(HaveOccurred())
			})
		})

		When("the handler does not configure a name", func() {
			BeforeEach(func() {
				handler.ConfigureFunc = func(c dogma.ProjectionConfigurer) {
					c.RouteEventType(fixtures.MessageA{})
				}
			})

			It("returns a descriptive error", func() {
				_, err := NewProjectionConfig(handler)

				Expect(err).To(Equal(
					Error(
						"*fixtures.ProjectionMessageHandler.Configure() did not call ProjectionConfigurer.Name()",
					),
				))
			})
		})

		When("the handler configures multiple names", func() {
			BeforeEach(func() {
				handler.ConfigureFunc = func(c dogma.ProjectionConfigurer) {
					c.Name("<name>")
					c.Name("<other>")
					c.RouteEventType(fixtures.MessageA{})
				}
			})

			It("returns a descriptive error", func() {
				_, err := NewProjectionConfig(handler)

				Expect(err).To(Equal(
					Error(
						`*fixtures.ProjectionMessageHandler.Configure() has already called ProjectionConfigurer.Name("<name>")`,
					),
				))
			})
		})

		When("the handler configures an invalid name", func() {
			BeforeEach(func() {
				handler.ConfigureFunc = func(c dogma.ProjectionConfigurer) {
					c.Name("\t \n")
					c.RouteEventType(fixtures.MessageA{})
				}
			})

			It("returns a descriptive error", func() {
				_, err := NewProjectionConfig(handler)

				Expect(err).To(Equal(
					Error(
						`*fixtures.ProjectionMessageHandler.Configure() called ProjectionConfigurer.Name("\t \n") with an invalid name`,
					),
				))
			})
		})

		When("the handler does not configure any routes", func() {
			BeforeEach(func() {
				handler.ConfigureFunc = func(c dogma.ProjectionConfigurer) {
					c.Name("<name>")
				}
			})

			It("returns a descriptive error", func() {
				_, err := NewProjectionConfig(handler)

				Expect(err).To(Equal(
					Error(
						"*fixtures.ProjectionMessageHandler.Configure() did not call ProjectionConfigurer.RouteEventType()",
					),
				))
			})
		})

		When("the handler does configures multiple routes for the same message type", func() {
			BeforeEach(func() {
				handler.ConfigureFunc = func(c dogma.ProjectionConfigurer) {
					c.Name("<name>")
					c.RouteEventType(fixtures.MessageA{})
					c.RouteEventType(fixtures.MessageA{})
				}
			})

			It("returns a descriptive error", func() {
				_, err := NewProjectionConfig(handler)

				Expect(err).To(Equal(
					Error(
						"*fixtures.ProjectionMessageHandler.Configure() has already called ProjectionConfigurer.RouteEventType(fixtures.MessageA)",
					),
				))
			})
		})
	})
})
