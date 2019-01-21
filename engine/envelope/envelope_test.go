package envelope_test

import (
	"time"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogmatest/engine/envelope"
	"github.com/dogmatiq/dogmatest/enginekit/message"
	"github.com/dogmatiq/dogmatest/internal/fixtures"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("type Envelope", func() {
	Describe("func New", func() {
		It("returns the expected envelope", func() {
			m := fixtures.MessageA{Value: "<value>"}
			env := New(m, message.CommandRole, time.Time{})

			Expect(env).To(Equal(
				&Envelope{
					Message: m,
					Type:    message.TypeOf(m),
					Role:    message.CommandRole,
					IsRoot:  true,
				},
			))
		})

		It("sets the timeout time", func() {
			m := fixtures.MessageA{Value: "<value>"}
			now := time.Now()
			env := New(m, message.TimeoutRole, now)

			Expect(env).To(Equal(
				&Envelope{
					Message:     m,
					Type:        message.TypeOf(m),
					Role:        message.TimeoutRole,
					IsRoot:      true,
					TimeoutTime: now,
				},
			))
		})

		It("panics if t is zero for a timeout", func() {
			Expect(func() {
				New(
					fixtures.MessageA{Value: "<value>"},
					message.TimeoutRole,
					time.Time{},
				)
			}).To(Panic())
		})

		It("panics if t is non-zero a non-timeout", func() {
			Expect(func() {
				New(
					fixtures.MessageA{Value: "<value>"},
					message.CommandRole,
					time.Now(),
				)
			}).To(Panic())
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
				time.Time{},
			)

			m = fixtures.MessageA{Value: "<parent>"}
		})

		It("returns the expected envelope", func() {
			env := parent.NewChild(m, message.EventRole, time.Time{})
			Expect(env).To(Equal(
				&Envelope{
					Message: m,
					Type:    message.TypeOf(m),
					Role:    message.EventRole,
					IsRoot:  false,
				},
			))
		})

		It("panics if t is zero for a timeout", func() {
			Expect(func() {
				parent.NewChild(
					fixtures.MessageA{Value: "<value>"},
					message.TimeoutRole,
					time.Time{},
				)
			}).To(Panic())
		})

		It("panics if t is non-zero a non-timeout", func() {
			Expect(func() {
				parent.NewChild(
					fixtures.MessageA{Value: "<value>"},
					message.CommandRole,
					time.Now(),
				)
			}).To(Panic())
		})
	})
})
