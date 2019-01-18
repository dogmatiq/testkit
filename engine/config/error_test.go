package config_test

import (
	. "github.com/dogmatiq/dogmatest/engine/config"
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
