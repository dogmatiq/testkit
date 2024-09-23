package envelope_test

import (
	"time"

	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/configkit/message"
	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/enginekit/enginetest/stubs"
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
				CommandA1,
				now,
			)

			Expect(env).To(Equal(
				&Envelope{
					MessageID:     "100",
					CorrelationID: "100",
					CausationID:   "100",
					Message:       CommandA1,
					Type:          message.TypeOf(CommandA1),
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
				EventA1,
				now,
			)

			Expect(env).To(Equal(
				&Envelope{
					MessageID:     "100",
					CorrelationID: "100",
					CausationID:   "100",
					Message:       EventA1,
					Type:          message.TypeOf(EventA1),
					Role:          message.EventRole,
					CreatedAt:     now,
				},
			))
		})
	})

	g.Describe("func NewCommand()", func() {
		handler := configkit.FromProcess(&ProcessMessageHandlerStub{
			ConfigureFunc: func(c dogma.ProcessConfigurer) {
				c.Identity("<handler>", "d1c7e18a-4d72-4705-a120-6cfb29eef655")
				c.Routes(
					dogma.HandlesEvent[EventStub[TypeA]](),
					dogma.ExecutesCommand[CommandStub[TypeA]](),
				)
			},
		})

		g.It("returns the expected envelope", func() {
			parent := NewEvent(
				"100",
				CommandP1,
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
				CommandA1,
				now,
				origin,
			)

			Expect(child).To(Equal(
				&Envelope{
					MessageID:     "200",
					CorrelationID: "100",
					CausationID:   "100",
					Message:       CommandA1,
					Type:          message.TypeOf(CommandA1),
					Role:          message.CommandRole,
					CreatedAt:     now,
					Origin:        &origin,
				},
			))
		})
	})

	g.Describe("func NewEvent()", func() {
		handler := configkit.FromAggregate(&AggregateMessageHandlerStub{
			ConfigureFunc: func(c dogma.AggregateConfigurer) {
				c.Identity("<handler>", "8688dc39-b5d0-4468-89fd-0d9452667c0c")
				c.Routes(
					dogma.HandlesCommand[CommandStub[TypeA]](),
					dogma.RecordsEvent[EventStub[TypeA]](),
				)
			},
		})

		g.It("returns the expected envelope", func() {
			parent := NewCommand(
				"100",
				CommandP1,
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
				EventA1,
				now,
				origin,
			)

			Expect(child).To(Equal(
				&Envelope{
					MessageID:     "200",
					CorrelationID: "100",
					CausationID:   "100",
					Message:       EventA1,
					Type:          message.TypeOf(EventA1),
					Role:          message.EventRole,
					CreatedAt:     now,
					Origin:        &origin,
				},
			))
		})
	})

	g.Describe("func NewTimeout()", func() {
		handler := configkit.FromProcess(&ProcessMessageHandlerStub{
			ConfigureFunc: func(c dogma.ProcessConfigurer) {
				c.Identity("<handler>", "1d4e3d22-52fe-4b1b-9bf5-44b2050c08c2")
				c.Routes(
					dogma.HandlesEvent[EventStub[TypeA]](),
					dogma.ExecutesCommand[CommandStub[TypeA]](),
					dogma.SchedulesTimeout[TimeoutStub[TypeA]](),
				)
			},
		})

		g.It("returns the expected envelope", func() {
			parent := NewCommand(
				"100",
				CommandP1,
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
				TimeoutA1,
				now,
				s,
				origin,
			)

			Expect(child).To(Equal(
				&Envelope{
					MessageID:     "200",
					CorrelationID: "100",
					CausationID:   "100",
					Message:       TimeoutA1,
					Type:          message.TypeOf(TimeoutA1),
					Role:          message.TimeoutRole,
					CreatedAt:     now,
					ScheduledFor:  s,
					Origin:        &origin,
				},
			))
		})
	})
})
