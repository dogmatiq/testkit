package envelope_test

import (
	. "github.com/dogmatiq/dogmatest/engine/envelope"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("type MessageRole", func() {
	Describe("func MustValidate", func() {
		It("does not panic when the role is valid", func() {
			CommandRole.MustValidate()
			EventRole.MustValidate()
		})

		It("panics when the role is not valid", func() {
			Expect(func() {
				MessageRole("").MustValidate()
			}).To(Panic())

		})
	})
})
