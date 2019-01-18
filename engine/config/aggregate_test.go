package config_test

import (
	"reflect"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogmatest/engine/config"
	"github.com/dogmatiq/dogmatest/internal/fixtures"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ Config = &AggregateConfig{}

var _ = Describe("type AggregateConfig", func() {
	Describe("func NewAggregateConfig", func() {
		var handler *fixtures.AggregateMessageHandler

		BeforeEach(func() {
			handler = &fixtures.AggregateMessageHandler{
				ConfigureFunc: func(c dogma.AggregateConfigurer) {
					c.Name("<name>")
					c.RouteCommandType(fixtures.Message{})
				},
			}
		})

		When("the configuration is successfully created", func() {
			var cfg *AggregateConfig

			BeforeEach(func() {
				var err error
				cfg, err = NewAggregateConfig(handler)
				Expect(err).ShouldNot(HaveOccurred())
			})

			It("the handler name is set", func() {
				Expect(cfg.HandlerName).To(Equal("<name>"))
			})

			It("the message types are in the set", func() {
				Expect(cfg.CommandTypes).To(HaveKey(
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
				_, err := NewAggregateConfig(handler)
				Expect(err).Should(HaveOccurred())
			})
		})

		When("the handler does not configure a name", func() {
			BeforeEach(func() {
				handler.ConfigureFunc = func(c dogma.AggregateConfigurer) {
					c.RouteCommandType(fixtures.Message{})
				}
			})

			It("returns a descriptive error", func() {
				_, err := NewAggregateConfig(handler)

				Expect(err).To(Equal(
					ConfigurationError(
						"*fixtures.AggregateMessageHandler.Configure() did not call AggregateConfigurer.Name()",
					),
				))
			})
		})

		When("the handler configures multiple names", func() {
			BeforeEach(func() {
				handler.ConfigureFunc = func(c dogma.AggregateConfigurer) {
					c.Name("<name>")
					c.Name("<other>")
					c.RouteCommandType(fixtures.Message{})
				}
			})

			It("returns a descriptive error", func() {
				_, err := NewAggregateConfig(handler)

				Expect(err).To(Equal(
					ConfigurationError(
						`*fixtures.AggregateMessageHandler.Configure() has already called AggregateConfigurer.Name("<name>")`,
					),
				))
			})
		})

		When("the handler configures an invalid name", func() {
			BeforeEach(func() {
				handler.ConfigureFunc = func(c dogma.AggregateConfigurer) {
					c.Name("\t \n")
					c.RouteCommandType(fixtures.Message{})
				}
			})

			It("returns a descriptive error", func() {
				_, err := NewAggregateConfig(handler)

				Expect(err).To(Equal(
					ConfigurationError(
						`*fixtures.AggregateMessageHandler.Configure() called AggregateConfigurer.Name("\t \n") with an invalid name`,
					),
				))
			})
		})

		When("the handler does not configure any routes", func() {
			BeforeEach(func() {
				handler.ConfigureFunc = func(c dogma.AggregateConfigurer) {
					c.Name("<name>")
				}
			})

			It("returns a descriptive error", func() {
				_, err := NewAggregateConfig(handler)

				Expect(err).To(Equal(
					ConfigurationError(
						"*fixtures.AggregateMessageHandler.Configure() did not call AggregateConfigurer.RouteCommandType()",
					),
				))
			})
		})

		When("the handler does configures multiple routes for the same message type", func() {
			BeforeEach(func() {
				handler.ConfigureFunc = func(c dogma.AggregateConfigurer) {
					c.Name("<name>")
					c.RouteCommandType(fixtures.Message{})
					c.RouteCommandType(fixtures.Message{})
				}
			})

			It("returns a descriptive error", func() {
				_, err := NewAggregateConfig(handler)

				Expect(err).To(Equal(
					ConfigurationError(
						"*fixtures.AggregateMessageHandler.Configure() has already called AggregateConfigurer.RouteCommandType(fixtures.Message)",
					),
				))
			})
		})
	})
})
