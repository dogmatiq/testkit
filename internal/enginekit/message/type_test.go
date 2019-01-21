package message_test

import (
	"reflect"

	. "github.com/dogmatiq/dogmatest/internal/enginekit/message"
	"github.com/dogmatiq/dogmatest/internal/fixtures"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("type MessageType", func() {
	Describe("func TypeOf", func() {
		It("returns values that compare as equal for messages of the same type", func() {
			ta := TypeOf(fixtures.MessageA1)
			tb := TypeOf(fixtures.MessageA1)

			Expect(ta).To(Equal(tb))
			Expect(ta == tb).To(BeTrue()) // explicitly check the pointers for standard equality comparability
		})

		It("returns values that do not compare as equal for messages of different types", func() {
			ta := TypeOf(fixtures.MessageA1)
			tb := TypeOf(fixtures.MessageB1)

			Expect(ta).NotTo(Equal(tb))
			Expect(ta != tb).To(BeTrue()) // explicitly check the pointers for standard equality comparability
		})
	})

	Describe("func ReflectType", func() {
		It("returns the reflect.Type for the message", func() {
			mt := TypeOf(fixtures.MessageA1)
			rt := reflect.TypeOf(fixtures.MessageA1)

			Expect(mt.ReflectType()).To(BeIdenticalTo(rt))
		})
	})

	Describe("func String", func() {
		It("returns the package-qualified type name and the pointer address", func() {
			t := TypeOf(fixtures.MessageA1)

			Expect(t.String()).To(Equal(
				"fixtures.MessageA",
			))
		})

		It("supports anonymous types", func() {
			t := TypeOf(struct{ fixtures.MessageA }{})

			Expect(t.String()).To(Equal("<anonymous>"))
		})
	})
})
