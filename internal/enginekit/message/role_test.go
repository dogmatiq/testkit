package message_test

import (
	. "github.com/dogmatiq/dogmatest/internal/enginekit/message"
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

	Describe("func MustBe", func() {
		It("does not panic when the role is in the given set", func() {
			CommandRole.MustBe(CommandRole, EventRole)
		})

		It("panics when the role is not in the given set", func() {
			Expect(func() {
				TimeoutRole.MustBe(CommandRole, EventRole)
			}).To(Panic())
		})
	})

	Describe("func MustNotBe", func() {
		It("does not panic when the role is not in the given set", func() {
			TimeoutRole.MustNotBe(CommandRole, EventRole)
		})

		It("panics when the role is in the given set", func() {
			Expect(func() {
				CommandRole.MustNotBe(CommandRole, EventRole)
			}).To(Panic())
		})
	})
})
