package compare_test

import (
	. "github.com/dogmatiq/dogmatest/compare"
	"github.com/dogmatiq/enginekit/fixtures"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ Comparator = DefaultComparator{}

var _ = Describe("type DefaultComparator", func() {
	comparator := DefaultComparator{}

	Describe("func MessageIsEqual", func() {
		It("returns true if the values have the same type and value", func() {
			Expect(comparator.MessageIsEqual(
				fixtures.MessageA1,
				fixtures.MessageA1,
			)).To(BeTrue())
		})

		It("returns false if the values have the same type with different values", func() {
			Expect(comparator.MessageIsEqual(
				fixtures.MessageA1,
				fixtures.MessageA2,
			)).To(BeFalse())
		})

		It("returns false if the values are different types", func() {
			Expect(comparator.MessageIsEqual(
				fixtures.MessageA1,
				fixtures.MessageA2,
			)).To(BeFalse())
		})
	})

	Describe("func AggregateRootIsEqual", func() {
		It("returns true if tlues have the same type and value", func() {
			Expect(comparator.AggregateRootIsEqual(
				&fixtures.AggregateRoot{Value: "<foo>"},
				&fixtures.AggregateRoot{Value: "<foo>"},
			)).To(BeTrue())
		})

		It("returns false if the values have the same type with different values", func() {
			Expect(comparator.AggregateRootIsEqual(
				&fixtures.AggregateRoot{Value: "<foo>"},
				&fixtures.AggregateRoot{Value: "<bar>"},
			)).To(BeFalse())
		})

		It("returns false if the values are different types", func() {
			Expect(comparator.AggregateRootIsEqual(
				&fixtures.AggregateRoot{Value: "<foo>"},
				&struct{ fixtures.AggregateRoot }{},
			)).To(BeFalse())
		})
	})

	Describe("func ProcessRootIsEqual", func() {
		It("returns true if tlues have the same type and value", func() {
			Expect(comparator.ProcessRootIsEqual(
				&fixtures.ProcessRoot{Value: "<foo>"},
				&fixtures.ProcessRoot{Value: "<foo>"},
			)).To(BeTrue())
		})

		It("returns false if the values have the same type with different values", func() {
			Expect(comparator.ProcessRootIsEqual(
				&fixtures.ProcessRoot{Value: "<foo>"},
				&fixtures.ProcessRoot{Value: "<bar>"},
			)).To(BeFalse())
		})

		It("returns false if the values are different types", func() {
			Expect(comparator.ProcessRootIsEqual(
				&fixtures.ProcessRoot{Value: "<foo>"},
				&struct{ fixtures.ProcessRoot }{},
			)).To(BeFalse())
		})
	})
})
