package envelope_test

import (
	. "github.com/dogmatiq/dogmatest/engine/envelope"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("type MessageIDGenerator", func() {
	var generator *MessageIDGenerator

	BeforeEach(func() {
		generator = &MessageIDGenerator{}
	})

	Describe("func Next", func() {
		It("returns the next ID in the sequence", func() {
			Expect(generator.Next()).To(Equal(uint64(1)))
			Expect(generator.Next()).To(Equal(uint64(2)))
			Expect(generator.Next()).To(Equal(uint64(3)))
		})
	})

	Describe("func Reset", func() {
		It("returns the sequence to 1", func() {
			generator.Next()
			generator.Next()
			generator.Next()
			generator.Reset()
			Expect(generator.Next()).To(Equal(uint64(1)))
		})
	})
})
