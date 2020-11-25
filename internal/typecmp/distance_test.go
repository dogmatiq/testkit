package typecmp_test

import (
	"reflect"

	. "github.com/dogmatiq/dogma/fixtures"
	. "github.com/dogmatiq/testkit/internal/typecmp"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("func MeasureDistance()", func() {
	It("returns Identical when given two identical types", func() {
		Expect(
			MeasureDistance(
				reflect.TypeOf(MessageA1),
				reflect.TypeOf(MessageA1),
			),
		).To(Equal(
			Identical,
		))
	})

	It("returns Unrelated when given two unrelated types", func() {
		Expect(
			MeasureDistance(
				reflect.TypeOf(MessageA1),
				reflect.TypeOf(MessageB1),
			),
		).To(Equal(
			Unrelated,
		))
	})

	It("returns some intermediate value when given types that differ only by 'pointer depth'", func() {
		a := reflect.PtrTo(reflect.TypeOf(MessageA1))
		b := reflect.TypeOf(MessageA1)

		dist := MeasureDistance(a, b)

		Expect(dist).NotTo(Equal(Identical))
		Expect(dist).NotTo(Equal(Unrelated))
	})

	It("returns the same value regardless of parameter order", func() {
		a := reflect.PtrTo(reflect.TypeOf(MessageA1))
		b := reflect.TypeOf(MessageA1)

		Expect(
			MeasureDistance(a, b),
		).To(Equal(
			MeasureDistance(b, a),
		))
	})

	It("returns a lower value for more similar types", func() {
		a := reflect.TypeOf(MessageA1)
		b := reflect.PtrTo(a)
		c := reflect.PtrTo(b)

		distA := MeasureDistance(a, b)
		distB := MeasureDistance(a, c)

		Expect(distA).To(BeNumerically("<", distB))
	})
})
