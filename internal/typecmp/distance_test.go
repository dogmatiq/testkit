package typecmp_test

import (
	"reflect"

	. "github.com/dogmatiq/enginekit/enginetest/stubs"
	. "github.com/dogmatiq/testkit/internal/typecmp"
	g "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = g.Describe("func MeasureDistance()", func() {
	g.It("returns Identical when given two identical types", func() {
		Expect(
			MeasureDistance(
				reflect.TypeOf(CommandA1),
				reflect.TypeOf(CommandA1),
			),
		).To(Equal(
			Identical,
		))
	})

	g.It("returns Unrelated when given two unrelated types", func() {
		Expect(
			MeasureDistance(
				reflect.TypeOf(CommandA1),
				reflect.TypeOf(EventA1),
			),
		).To(Equal(
			Unrelated,
		))
	})

	g.It("returns some intermediate value when given types that differ only by 'pointer depth'", func() {
		a := reflect.PointerTo(reflect.TypeOf(CommandA1))
		b := reflect.TypeOf(CommandA1)

		dist := MeasureDistance(a, b)

		Expect(dist).NotTo(Equal(Identical))
		Expect(dist).NotTo(Equal(Unrelated))
	})

	g.It("returns the same value regardless of parameter order", func() {
		a := reflect.PointerTo(reflect.TypeOf(CommandA1))
		b := reflect.TypeOf(CommandA1)

		Expect(
			MeasureDistance(a, b),
		).To(Equal(
			MeasureDistance(b, a),
		))
	})

	g.It("returns a lower value for more similar types", func() {
		a := reflect.TypeOf(CommandA1)
		b := reflect.PointerTo(a)
		c := reflect.PointerTo(b)

		distA := MeasureDistance(a, b)
		distB := MeasureDistance(a, c)

		Expect(distA).To(BeNumerically("<", distB))
	})
})
