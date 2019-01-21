package message_test

import (
	"reflect"

	"github.com/dogmatiq/dogmatest/internal/enginekit/fixtures"
	. "github.com/dogmatiq/dogmatest/internal/enginekit/message"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("type Type", func() {
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

var _ = Describe("type TypeSet", func() {
	Describe("func TypesOf", func() {
		It("returns a set containing the types of the given messages", func() {
			Expect(TypesOf(
				fixtures.MessageA1,
				fixtures.MessageB1,
			)).To(Equal(TypeSet{
				fixtures.MessageAType: struct{}{},
				fixtures.MessageBType: struct{}{},
			}))
		})
	})

	Describe("func Has", func() {
		set := TypesOf(
			fixtures.MessageA1,
			fixtures.MessageB1,
		)

		It("returns true if the type is in the set", func() {
			Expect(
				set.Has(fixtures.MessageAType),
			).To(BeTrue())
		})

		It("returns false if the type is not in the set", func() {
			Expect(
				set.Has(fixtures.MessageCType),
			).To(BeFalse())
		})
	})

	Describe("func Add", func() {
		It("adds the type to the set", func() {
			s := TypesOf()
			s.Add(fixtures.MessageAType)

			Expect(
				s.Has(fixtures.MessageAType),
			).To(BeTrue())
		})
	})

	Describe("func Remove", func() {
		It("removes the type from the set", func() {
			s := TypesOf(fixtures.MessageA1)
			s.Remove(fixtures.MessageAType)

			Expect(
				s.Has(fixtures.MessageAType),
			).To(BeFalse())
		})
	})

	Describe("func AddM", func() {
		It("adds the type of the message to the set", func() {
			s := TypesOf()
			s.AddM(fixtures.MessageA1)

			Expect(
				s.Has(fixtures.MessageAType),
			).To(BeTrue())
		})
	})

	Describe("func RemoveM", func() {
		It("removes the type of the message from the set", func() {
			s := TypesOf(fixtures.MessageA1)
			s.RemoveM(fixtures.MessageA1)

			Expect(
				s.Has(fixtures.MessageAType),
			).To(BeFalse())
		})
	})
})
