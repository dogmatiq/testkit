package message_test

import (
	. "github.com/dogmatiq/dogmatest/enginekit/message"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("type Role", func() {
	Describe("func MustValidate", func() {
		It("does not panic when the role is valid", func() {
			CommandRole.MustValidate()
			EventRole.MustValidate()
			TimeoutRole.MustValidate()
		})

		It("panics when the role is not valid", func() {
			Expect(func() {
				Role(-1).MustValidate()
			}).To(Panic())
		})
	})
})
