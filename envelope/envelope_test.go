package envelope_test

import (
	"time"

	"github.com/dogmatiq/configkit"
	. "github.com/dogmatiq/configkit/fixtures"
	"github.com/dogmatiq/configkit/message"
	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogma/fixtures"
	. "github.com/dogmatiq/testkit/envelope"
	g "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = g.Describe("type Envelope", func() {
	g.Describe("func NewCommand()", func() {
		g.It("returns the expected envelope", func() {
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

	g.Describe("func NewEvent()", func() {
		g.It("returns the expected envelope", func() {
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

	g.Describe("func NewCommand()", func() {
		handler := configkit.FromProcess(&ProcessMessageHandler{
			ConfigureFunc: func(c dogma.ProcessConfigurer) {
				c.Identity("<handler>", "d1c7e18a-4d72-4705-a120-6cfb29eef655")
				c.Routes(
					dogma.HandlesEvent[MessageE](),
					dogma.ExecutesCommand[MessageC](),
				)
			},
		})

		g.It("returns the expected envelope", func() {
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

	g.Describe("func NewEvent()", func() {
		handler := configkit.FromAggregate(&AggregateMessageHandler{
			ConfigureFunc: func(c dogma.AggregateConfigurer) {
				c.Identity("<handler>", "8688dc39-b5d0-4468-89fd-0d9452667c0c")
				c.Routes(
					dogma.HandlesCommand[MessageC](),
					dogma.RecordsEvent[MessageE](),
				)
			},
		})

		g.It("returns the expected envelope", func() {
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

	g.Describe("func NewTimeout()", func() {
		handler := configkit.FromProcess(&ProcessMessageHandler{
			ConfigureFunc: func(c dogma.ProcessConfigurer) {
				c.Identity("<handler>", "1d4e3d22-52fe-4b1b-9bf5-44b2050c08c2")
				c.Routes(
					dogma.HandlesEvent[MessageE](),
					dogma.ExecutesCommand[MessageC](),
					dogma.SchedulesTimeout[MessageT](),
				)
			},
		})

		g.It("returns the expected envelope", func() {
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
