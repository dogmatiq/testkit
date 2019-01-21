package config_test

import (
	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogmatest/internal/enginekit/config"
	"github.com/dogmatiq/dogmatest/internal/enginekit/fixtures"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ Config = &AppConfig{}

var _ = Describe("type Error", func() {
	Describe("func Error", func() {
		It("returns the error message", func() {
			err := Error("<message>")
			Expect(err.Error()).To(Equal("<message>"))
		})
	})
})

var _ = Describe("func catch", func() {
	It("does not catch non-config panics", func() {
		Expect(func() {
			NewAggregateConfig(&fixtures.AggregateMessageHandler{
				ConfigureFunc: func(c dogma.AggregateConfigurer) {
					panic("<panic>")
				},
			})
		}).To(Panic())
	})
})
