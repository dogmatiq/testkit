package envelope_test

import (
	"time"

	. "github.com/dogmatiq/dogmatest/engine/envelope"
	"github.com/dogmatiq/dogmatest/internal/enginekit/fixtures"
	"github.com/dogmatiq/dogmatest/internal/enginekit/message"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("type Envelope", func() {
	Describe("func New", func() {
		It("returns the expected envelope", func() {
			env := New(fixtures.MessageC1, message.CommandRole)

			Expect(env).To(Equal(
				&Envelope{
					Message: fixtures.MessageC1,
					Type:    fixtures.MessageCType,
					Role:    message.CommandRole,
					IsRoot:  true,
				},
			))
		})

		It("panics if called with the timeout role", func() {
			Expect(func() {
				New(
					fixtures.MessageA1,
					message.TimeoutRole,
				)
			}).To(Panic())
		})
	})

	Describe("func NewCommand", func() {
		It("returns the expected envelope", func() {
			parent := New(
				fixtures.MessageP1,
				message.EventRole,
			)
			child := parent.NewCommand(
				fixtures.MessageC1,
			)

			Expect(child).To(Equal(
				&Envelope{
					Message: fixtures.MessageC1,
					Type:    fixtures.MessageCType,
					Role:    message.CommandRole,
					IsRoot:  false,
				},
			))
		})
	})

	Describe("func NewEvent", func() {
		It("returns the expected envelope", func() {
			parent := New(
				fixtures.MessageP1,
				message.CommandRole,
			)
			child := parent.NewEvent(
				fixtures.MessageE1,
			)

			Expect(child).To(Equal(
				&Envelope{
					Message: fixtures.MessageE1,
					Type:    fixtures.MessageEType,
					Role:    message.EventRole,
					IsRoot:  false,
				},
			))
		})
	})

	Describe("func NewTimeout", func() {
		It("returns the expected envelope", func() {
			t := time.Now()
			parent := New(
				fixtures.MessageP1,
				message.CommandRole,
			)
			child := parent.NewTimeout(
				fixtures.MessageT1,
				t,
			)

			Expect(child).To(Equal(
				&Envelope{
					Message:     fixtures.MessageT1,
					Type:        fixtures.MessageTType,
					Role:        message.TimeoutRole,
					IsRoot:      false,
					TimeoutTime: &t,
				},
			))
		})
	})
})
