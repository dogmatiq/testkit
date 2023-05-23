package typecmp_test

import (
	"reflect"

	. "github.com/dogmatiq/dogma/fixtures"
	. "github.com/dogmatiq/testkit/internal/typecmp"
	g "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = g.Describe("func MeasureDistance()", func() {
	g.It("returns Identical when given two identical types", func() {
		Expect(
			MeasureDistance(
				reflect.TypeOf(MessageA1),
				reflect.TypeOf(MessageA1),
			),
		).To(Equal(
			Identical,
		))
	})

	g.It("returns Unrelated when given two unrelated types", func() {
		Expect(
			MeasureDistance(
				reflect.TypeOf(MessageA1),
				reflect.TypeOf(MessageB1),
			),
		).To(Equal(
			Unrelated,
		))
	})

	g.It("returns some intermediate value when given types that differ only by 'pointer depth'", func() {
		a := reflect.PtrTo(reflect.TypeOf(MessageA1))
		b := reflect.TypeOf(MessageA1)

		dist := MeasureDistance(a, b)

		Expect(dist).NotTo(Equal(Identical))
		Expect(dist).NotTo(Equal(Unrelated))
	})

	g.It("returns the same value regardless of parameter order", func() {
		a := reflect.PtrTo(reflect.TypeOf(MessageA1))
		b := reflect.TypeOf(MessageA1)

		Expect(
			MeasureDistance(a, b),
		).To(Equal(
			MeasureDistance(b, a),
		))
	})

	g.It("returns a lower value for more similar types", func() {
		a := reflect.TypeOf(MessageA1)
		b := reflect.PtrTo(a)
		c := reflect.PtrTo(b)

		distA := MeasureDistance(a, b)
		distB := MeasureDistance(a, c)

		Expect(distA).To(BeNumerically("<", distB))
	})
})
