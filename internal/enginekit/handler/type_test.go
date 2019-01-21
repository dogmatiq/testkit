package handler_test

import (
	. "github.com/dogmatiq/dogmatest/internal/enginekit/handler"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("type Type", func() {
	Describe("func MustValidate", func() {
		It("does not panic when the type is valid", func() {
			AggregateType.MustValidate()
			ProcessType.MustValidate()
			IntegrationType.MustValidate()
			ProjectionType.MustValidate()
		})

		It("panics when the type is not valid", func() {
			Expect(func() {
				Type(-1).MustValidate()
			}).To(Panic())
		})
	})

	Describe("func MustBe", func() {
		It("does not panic when the type is in the given set", func() {
			AggregateType.MustBe(AggregateType, ProcessType)
		})

		It("panics when the type is not in the given set", func() {
			Expect(func() {
				IntegrationType.MustBe(AggregateType, ProcessType)
			}).To(Panic())
		})
	})

	Describe("func MustNotBe", func() {
		It("does not panic when the type is not in the given set", func() {
			IntegrationType.MustNotBe(AggregateType, ProcessType)
		})

		It("panics when the type is in the given set", func() {
			Expect(func() {
				AggregateType.MustNotBe(AggregateType, ProcessType)
			}).To(Panic())
		})
	})
})
