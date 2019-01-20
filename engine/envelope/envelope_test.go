package envelope_test

import (
	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogmatest/engine/envelope"
	"github.com/dogmatiq/dogmatest/internal/enginekit/message"
	"github.com/dogmatiq/dogmatest/internal/fixtures"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("type Envelope", func() {
	Describe("func New", func() {
		It("returns the expected envelope", func() {
			m := fixtures.MessageA{Value: "<value>"}
			env := New(m, message.CommandRole)

			Expect(env).To(Equal(
				&Envelope{
					Message: m,
					Type:    message.TypeOf(m),
					Role:    message.CommandRole,
					IsRoot:  true,
				},
			))
		})
	})

	Describe("func NewChild", func() {
		var (
			parent *Envelope
			m      dogma.Message
		)

		BeforeEach(func() {
			parent = New(
				fixtures.MessageA{Value: "<parent>"},
				message.CommandRole,
			)

			m = fixtures.MessageA{Value: "<parent>"}
		})

		It("returns the expected envelope", func() {
			env := parent.NewChild(m, message.EventRole)
			Expect(env).To(Equal(
				&Envelope{
					Message: m,
					Type:    message.TypeOf(m),
					Role:    message.EventRole,
					IsRoot:  false,
				},
			))

		})
	})
})
