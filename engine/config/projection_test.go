package config_test

import (
	"reflect"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogmatest/engine/config"
	"github.com/dogmatiq/dogmatest/internal/fixtures"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ Config = &ProjectionConfig{}

var _ = Describe("type ProjectionConfig", func() {
	Describe("func NewProjectionConfig", func() {
		var handler *fixtures.ProjectionMessageHandler

		BeforeEach(func() {
			handler = &fixtures.ProjectionMessageHandler{
				ConfigureFunc: func(c dogma.ProjectionConfigurer) {
					c.Name("<name>")
					c.RouteEventType(fixtures.Message{})
				},
			}
		})

		When("the configuration is successfully created", func() {
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
				Expect(cfg.EventTypes).To(HaveKey(
					reflect.TypeOf(fixtures.Message{}),
				))
			})

			Describe("func Name()", func() {
				It("returns the handler name", func() {
					Expect(cfg.Name()).To(Equal("<name>"))
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
					c.RouteEventType(fixtures.Message{})
				}
			})

			It("returns a descriptive error", func() {
				_, err := NewProjectionConfig(handler)

				Expect(err).To(Equal(
					ConfigurationError(
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
					c.RouteEventType(fixtures.Message{})
				}
			})

			It("returns a descriptive error", func() {
				_, err := NewProjectionConfig(handler)

				Expect(err).To(Equal(
					ConfigurationError(
						`*fixtures.ProjectionMessageHandler.Configure() has already called ProjectionConfigurer.Name("<name>")`,
					),
				))
			})
		})

		When("the handler configures an invalid name", func() {
			BeforeEach(func() {
				handler.ConfigureFunc = func(c dogma.ProjectionConfigurer) {
					c.Name("\t \n")
					c.RouteEventType(fixtures.Message{})
				}
			})

			It("returns a descriptive error", func() {
				_, err := NewProjectionConfig(handler)

				Expect(err).To(Equal(
					ConfigurationError(
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
					ConfigurationError(
						"*fixtures.ProjectionMessageHandler.Configure() did not call ProjectionConfigurer.RouteEventType()",
					),
				))
			})
		})

		When("the handler does configures multiple routes for the same message type", func() {
			BeforeEach(func() {
				handler.ConfigureFunc = func(c dogma.ProjectionConfigurer) {
					c.Name("<name>")
					c.RouteEventType(fixtures.Message{})
					c.RouteEventType(fixtures.Message{})
				}
			})

			It("returns a descriptive error", func() {
				_, err := NewProjectionConfig(handler)

				Expect(err).To(Equal(
					ConfigurationError(
						"*fixtures.ProjectionMessageHandler.Configure() has already called ProjectionConfigurer.RouteEventType(fixtures.Message)",
					),
				))
			})
		})
	})
})
