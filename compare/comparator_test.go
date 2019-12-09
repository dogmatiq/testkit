package compare_test

import (
	. "github.com/dogmatiq/dogma/fixtures"
	. "github.com/dogmatiq/testkit/compare"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ Comparator = DefaultComparator{}

var _ = Describe("type DefaultComparator", func() {
	comparator := DefaultComparator{}

	Describe("func MessageIsEqual()", func() {
		It("returns true if the values have the same type and value", func() {
			Expect(comparator.MessageIsEqual(
				MessageA1,
				MessageA1,
			)).To(BeTrue())
		})

		It("returns false if the values have the same type with different values", func() {
			Expect(comparator.MessageIsEqual(
				MessageA1,
				MessageA2,
			)).To(BeFalse())
		})

		It("returns false if the values are different types", func() {
			Expect(comparator.MessageIsEqual(
				MessageA1,
				MessageA2,
			)).To(BeFalse())
		})
	})

	Describe("func AggregateRootIsEqual()", func() {
		It("returns true if tlues have the same type and value", func() {
			Expect(comparator.AggregateRootIsEqual(
				&AggregateRoot{Value: "<foo>"},
				&AggregateRoot{Value: "<foo>"},
			)).To(BeTrue())
		})

		It("returns false if the values have the same type with different values", func() {
			Expect(comparator.AggregateRootIsEqual(
				&AggregateRoot{Value: "<foo>"},
				&AggregateRoot{Value: "<bar>"},
			)).To(BeFalse())
		})

		It("returns false if the values are different types", func() {
			Expect(comparator.AggregateRootIsEqual(
				&AggregateRoot{Value: "<foo>"},
				&struct{ AggregateRoot }{},
			)).To(BeFalse())
		})
	})

	Describe("func ProcessRootIsEqual()", func() {
		It("returns true if tlues have the same type and value", func() {
			Expect(comparator.ProcessRootIsEqual(
				&ProcessRoot{Value: "<foo>"},
				&ProcessRoot{Value: "<foo>"},
			)).To(BeTrue())
		})

		It("returns false if the values have the same type with different values", func() {
			Expect(comparator.ProcessRootIsEqual(
				&ProcessRoot{Value: "<foo>"},
				&ProcessRoot{Value: "<bar>"},
			)).To(BeFalse())
		})

		It("returns false if the values are different types", func() {
			Expect(comparator.ProcessRootIsEqual(
				&ProcessRoot{Value: "<foo>"},
				&struct{ ProcessRoot }{},
			)).To(BeFalse())
		})
	})
})
