package envelope_test

import (
	"time"

	"github.com/dogmatiq/enginekit/fixtures"
	"github.com/dogmatiq/enginekit/handler"
	"github.com/dogmatiq/enginekit/message"
	. "github.com/dogmatiq/testkit/engine/envelope"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("type Envelope", func() {
	Describe("func New", func() {
		It("returns the expected envelope", func() {
			now := time.Now()
			env := New(
				"100",
				fixtures.MessageC1,
				message.CommandRole,
				now,
			)

			Expect(env).To(Equal(
				&Envelope{
					Correlation: message.Correlation{
						MessageID:     "100",
						CorrelationID: "100",
						CausationID:   "100",
					},
					Message: fixtures.MessageC1,
					Type:    fixtures.MessageCType,
					Role:    message.CommandRole,
					Time:    now,
				},
			))
		})

		It("panics if called with the timeout role", func() {
			Expect(func() {
				New(
					"100",
					fixtures.MessageA1,
					message.TimeoutRole,
					time.Now(),
				)
			}).To(Panic())
		})
	})

	Describe("func NewCommand", func() {
		It("returns the expected envelope", func() {
			parent := New(
				"100",
				fixtures.MessageP1,
				message.EventRole,
				time.Now(),
			)
			origin := Origin{
				HandlerName: "<handler>",
				HandlerType: handler.ProcessType,
				InstanceID:  "<instance>",
			}
			now := time.Now()
			child := parent.NewCommand(
				"200",
				fixtures.MessageC1,
				now,
				origin,
			)

			Expect(child).To(Equal(
				&Envelope{
					Correlation: message.Correlation{
						MessageID:     "200",
						CorrelationID: "100",
						CausationID:   "100",
					},
					Message: fixtures.MessageC1,
					Type:    fixtures.MessageCType,
					Role:    message.CommandRole,
					Time:    now,
					Origin:  &origin,
				},
			))
		})
	})

	Describe("func NewEvent", func() {
		It("returns the expected envelope", func() {
			parent := New(
				"100",
				fixtures.MessageP1,
				message.CommandRole,
				time.Now(),
			)
			origin := Origin{
				HandlerName: "<handler>",
				HandlerType: handler.AggregateType,
				InstanceID:  "<instance>",
			}
			now := time.Now()
			child := parent.NewEvent(
				"200",
				fixtures.MessageE1,
				now,
				origin,
			)

			Expect(child).To(Equal(
				&Envelope{
					Correlation: message.Correlation{
						MessageID:     "200",
						CorrelationID: "100",
						CausationID:   "100",
					},
					Message: fixtures.MessageE1,
					Type:    fixtures.MessageEType,
					Role:    message.EventRole,
					Time:    now,
					Origin:  &origin,
				},
			))
		})
	})

	Describe("func NewTimeout", func() {
		It("returns the expected envelope", func() {
			parent := New(
				"100",
				fixtures.MessageP1,
				message.CommandRole,
				time.Now(),
			)
			origin := Origin{
				HandlerName: "<handler>",
				HandlerType: handler.ProcessType,
				InstanceID:  "<instance>",
			}
			now := time.Now()
			t := time.Now()
			child := parent.NewTimeout(
				"200",
				fixtures.MessageT1,
				now,
				t,
				origin,
			)

			Expect(child).To(Equal(
				&Envelope{
					Correlation: message.Correlation{
						MessageID:     "200",
						CorrelationID: "100",
						CausationID:   "100",
					},
					Message:     fixtures.MessageT1,
					Type:        fixtures.MessageTType,
					Role:        message.TimeoutRole,
					Time:        now,
					TimeoutTime: &t,
					Origin:      &origin,
				},
			))
		})
	})
})
