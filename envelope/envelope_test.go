package envelope_test

import (
	"time"

	"github.com/dogmatiq/configkit"
	. "github.com/dogmatiq/configkit/fixtures"
	"github.com/dogmatiq/configkit/message"
	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogma/fixtures"
	. "github.com/dogmatiq/testkit/envelope"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("type Envelope", func() {
	Describe("func NewCommand()", func() {
		It("returns the expected envelope", func() {
			now := time.Now()
			env := NewCommand(
				"100",
				MessageC1,
				now,
			)

			Expect(env).To(Equal(
				&Envelope{
					MessageID:     "100",
					CorrelationID: "100",
					CausationID:   "100",
					Message:       MessageC1,
					Type:          MessageCType,
					Role:          message.CommandRole,
					CreatedAt:     now,
				},
			))
		})
	})

	Describe("func NewEvent()", func() {
		It("returns the expected envelope", func() {
			now := time.Now()
			env := NewEvent(
				"100",
				MessageE1,
				now,
			)

			Expect(env).To(Equal(
				&Envelope{
					MessageID:     "100",
					CorrelationID: "100",
					CausationID:   "100",
					Message:       MessageE1,
					Type:          MessageEType,
					Role:          message.EventRole,
					CreatedAt:     now,
				},
			))
		})
	})

	Describe("func NewCommand()", func() {
		handler := configkit.FromProcess(&ProcessMessageHandler{
			ConfigureFunc: func(c dogma.ProcessConfigurer) {
				c.Identity("<handler>", "<handler-key>")
				c.ConsumesEventType(MessageE{})
				c.ProducesCommandType(MessageC{})
			},
		})

		It("returns the expected envelope", func() {
			parent := NewEvent(
				"100",
				MessageP1,
				time.Now(),
			)
			origin := Origin{
				Handler:     handler,
				HandlerType: configkit.ProcessHandlerType,
				InstanceID:  "<instance>",
			}
			now := time.Now()
			child := parent.NewCommand(
				"200",
				MessageC1,
				now,
				origin,
			)

			Expect(child).To(Equal(
				&Envelope{
					MessageID:     "200",
					CorrelationID: "100",
					CausationID:   "100",
					Message:       MessageC1,
					Type:          MessageCType,
					Role:          message.CommandRole,
					CreatedAt:     now,
					Origin:        &origin,
				},
			))
		})
	})

	Describe("func NewEvent()", func() {
		handler := configkit.FromAggregate(&AggregateMessageHandler{
			ConfigureFunc: func(c dogma.AggregateConfigurer) {
				c.Identity("<handler>", "<handler-key>")
				c.ConsumesCommandType(MessageC{})
				c.ProducesEventType(MessageE{})
			},
		})

		It("returns the expected envelope", func() {
			parent := NewCommand(
				"100",
				MessageP1,
				time.Now(),
			)
			origin := Origin{
				Handler:     handler,
				HandlerType: configkit.AggregateHandlerType,
				InstanceID:  "<instance>",
			}
			now := time.Now()
			child := parent.NewEvent(
				"200",
				MessageE1,
				now,
				origin,
			)

			Expect(child).To(Equal(
				&Envelope{
					MessageID:     "200",
					CorrelationID: "100",
					CausationID:   "100",
					Message:       MessageE1,
					Type:          MessageEType,
					Role:          message.EventRole,
					CreatedAt:     now,
					Origin:        &origin,
				},
			))
		})
	})

	Describe("func NewTimeout()", func() {
		handler := configkit.FromProcess(&ProcessMessageHandler{
			ConfigureFunc: func(c dogma.ProcessConfigurer) {
				c.Identity("<handler>", "<handler-key>")
				c.ConsumesEventType(MessageE{})
				c.ProducesCommandType(MessageC{})
				c.SchedulesTimeoutType(MessageT{})
			},
		})

		It("returns the expected envelope", func() {
			parent := NewCommand(
				"100",
				MessageP1,
				time.Now(),
			)
			origin := Origin{
				Handler:     handler,
				HandlerType: configkit.ProcessHandlerType,
				InstanceID:  "<instance>",
			}
			now := time.Now()
			s := time.Now()
			child := parent.NewTimeout(
				"200",
				MessageT1,
				now,
				s,
				origin,
			)

			Expect(child).To(Equal(
				&Envelope{
					MessageID:     "200",
					CorrelationID: "100",
					CausationID:   "100",
					Message:       MessageT1,
					Type:          MessageTType,
					Role:          message.TimeoutRole,
					CreatedAt:     now,
					ScheduledFor:  s,
					Origin:        &origin,
				},
			))
		})
	})
})
