package envelope_test

import (
	. "github.com/dogmatiq/testkit/envelope"
	g "github.com/onsi/ginkgo/v2"
	gm "github.com/onsi/gomega"
)

var _ = g.Describe("type MessageIDGenerator", func() {
	var generator *MessageIDGenerator

	g.BeforeEach(func() {
		generator = &MessageIDGenerator{}
	})

	g.Describe("func Next()", func() {
		g.It("returns the next ID in the sequence", func() {
			gm.Expect(generator.Next()).To(gm.Equal("1"))
			gm.Expect(generator.Next()).To(gm.Equal("2"))
			gm.Expect(generator.Next()).To(gm.Equal("3"))
		})
	})

	g.Describe("func Reset()", func() {
		g.It("returns the sequence to 1", func() {
			generator.Next()
			generator.Next()
			generator.Next()
			generator.Reset()
			gm.Expect(generator.Next()).To(gm.Equal("1"))
		})
	})
})
