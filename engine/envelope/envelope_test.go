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
	Describe("func NewCommand", func() {
		It("returns the expected envelope", func() {
			now := time.Now()
			env := NewCommand(
				"100",
				fixtures.MessageC1,
				now,
			)

			Expect(env).To(Equal(
				&Envelope{
					Correlation: message.Correlation{
						MessageID:     "100",
						CorrelationID: "100",
						CausationID:   "100",
					},
					Message:   fixtures.MessageC1,
					Type:      fixtures.MessageCType,
					Role:      message.CommandRole,
					CreatedAt: now,
				},
			))
		})
	})

	Describe("func NewEvent", func() {
		It("returns the expected envelope", func() {
			now := time.Now()
			env := NewEvent(
				"100",
				fixtures.MessageE1,
				now,
			)

			Expect(env).To(Equal(
				&Envelope{
					Correlation: message.Correlation{
						MessageID:     "100",
						CorrelationID: "100",
						CausationID:   "100",
					},
					Message:   fixtures.MessageE1,
					Type:      fixtures.MessageEType,
					Role:      message.EventRole,
					CreatedAt: now,
				},
			))
		})
	})

	Describe("func NewCommand", func() {
		It("returns the expected envelope", func() {
			parent := NewEvent(
				"100",
				fixtures.MessageP1,
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
					Message:   fixtures.MessageC1,
					Type:      fixtures.MessageCType,
					Role:      message.CommandRole,
					CreatedAt: now,
					Origin:    &origin,
				},
			))
		})
	})

	Describe("func NewEvent", func() {
		It("returns the expected envelope", func() {
			parent := NewCommand(
				"100",
				fixtures.MessageP1,
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
					Message:   fixtures.MessageE1,
					Type:      fixtures.MessageEType,
					Role:      message.EventRole,
					CreatedAt: now,
					Origin:    &origin,
				},
			))
		})
	})

	Describe("func NewTimeout", func() {
		It("returns the expected envelope", func() {
			parent := NewCommand(
				"100",
				fixtures.MessageP1,
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
					CreatedAt:   now,
					TimeoutTime: &t,
					Origin:      &origin,
				},
			))
		})
	})
})
