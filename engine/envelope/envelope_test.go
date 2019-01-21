package envelope_test

import (
	"time"

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

		It("panics if called with the timeout role", func() {
			Expect(func() {
				New(
					fixtures.MessageA{Value: "<value>"},
					message.TimeoutRole,
				)
			}).To(Panic())
		})
	})

	Describe("func NewCommand", func() {
		It("returns the expected envelope", func() {
			parent := New(
				fixtures.MessageA{Value: "<parent>"},
				message.EventRole,
			)
			m := fixtures.MessageA{Value: "<child>"}
			env := parent.NewCommand(m)

			Expect(env).To(Equal(
				&Envelope{
					Message: m,
					Type:    message.TypeOf(m),
					Role:    message.CommandRole,
					IsRoot:  false,
				},
			))
		})
	})

	Describe("func NewEvent", func() {
		It("returns the expected envelope", func() {
			parent := New(
				fixtures.MessageA{Value: "<parent>"},
				message.CommandRole,
			)
			m := fixtures.MessageA{Value: "<child>"}
			env := parent.NewEvent(m)

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

	Describe("func NewTimeout", func() {
		It("returns the expected envelope", func() {
			parent := New(
				fixtures.MessageA{Value: "<parent>"},
				message.CommandRole,
			)
			m := fixtures.MessageA{Value: "<child>"}
			t := time.Now()
			env := parent.NewTimeout(m, t)

			Expect(env).To(Equal(
				&Envelope{
					Message:     m,
					Type:        message.TypeOf(m),
					Role:        message.TimeoutRole,
					IsRoot:      false,
					TimeoutTime: &t,
				},
			))
		})
	})
})
