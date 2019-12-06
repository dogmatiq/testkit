package compare_test

import (
	"reflect"

	. "github.com/dogmatiq/dogma/fixtures"
	. "github.com/dogmatiq/testkit/compare"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("func FuzzyTypeComparison", func() {
	It("returns SameTypes when given two identical types", func() {
		Expect(
			FuzzyTypeComparison(
				reflect.TypeOf(MessageA1),
				reflect.TypeOf(MessageA1),
			),
		).To(Equal(
			SameTypes,
		))
	})

	It("returns SameTypes when given two unrelated types", func() {
		Expect(
			FuzzyTypeComparison(
				reflect.TypeOf(MessageA1),
				reflect.TypeOf(MessageB1),
			),
		).To(Equal(
			UnrelatedTypes,
		))
	})

	It("returns some intermediate value when given types that differ only by 'pointer depth'", func() {
		a := reflect.PtrTo(reflect.TypeOf(MessageA1))
		b := reflect.TypeOf(MessageA1)

		sim := FuzzyTypeComparison(a, b)

		Expect(sim).NotTo(Equal(SameTypes))
		Expect(sim).NotTo(Equal(UnrelatedTypes))
	})

	It("returns the same value regardless of parameter order", func() {
		a := reflect.PtrTo(reflect.TypeOf(MessageA1))
		b := reflect.TypeOf(MessageA1)

		Expect(
			FuzzyTypeComparison(a, b),
		).To(Equal(
			FuzzyTypeComparison(b, a),
		))
	})

	It("returns a higher value for more similar types", func() {
		a := reflect.TypeOf(MessageA1)
		b := reflect.PtrTo(a)
		c := reflect.PtrTo(b)

		simA := FuzzyTypeComparison(a, b)
		simB := FuzzyTypeComparison(a, c)

		Expect(simA).To(BeNumerically(">", simB))
	})
})
