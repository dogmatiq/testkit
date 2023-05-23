package envelope_test

import (
	. "github.com/dogmatiq/testkit/envelope"
	g "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = g.Describe("type MessageIDGenerator", func() {
	var generator *MessageIDGenerator

	g.BeforeEach(func() {
		generator = &MessageIDGenerator{}
	})

	g.Describe("func Next()", func() {
		g.It("returns the next ID in the sequence", func() {
			Expect(generator.Next()).To(Equal("1"))
			Expect(generator.Next()).To(Equal("2"))
			Expect(generator.Next()).To(Equal("3"))
		})
	})

	g.Describe("func Reset()", func() {
		g.It("returns the sequence to 1", func() {
			generator.Next()
			generator.Next()
			generator.Next()
			generator.Reset()
			Expect(generator.Next()).To(Equal("1"))
		})
	})
})
