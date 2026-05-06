package envelope_test

import (
	"testing"
	"time"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/enginekit/config"
	"github.com/dogmatiq/enginekit/config/runtimeconfig"
	. "github.com/dogmatiq/enginekit/enginetest/stubs"
	. "github.com/dogmatiq/testkit/envelope"
	"github.com/dogmatiq/testkit/internal/x/xtesting"
)

func TestEnvelope(t *testing.T) {
	t.Run("func NewCommand()", func(t *testing.T) {
		t.Run("it returns the expected envelope", func(t *testing.T) {
			now := time.Now()
			env := NewCommand(
				"100",
				CommandA1,
				now,
			)

			xtesting.Expect(
				t,
				"unexpected envelope",
				env,
				&Envelope{
					MessageID:     "100",
					CorrelationID: "100",
					CausationID:   "100",
					Message:       CommandA1,
					CreatedAt:     now,
				},
			)
		})
	})

	t.Run("func NewEvent()", func(t *testing.T) {
		t.Run("it returns the expected envelope", func(t *testing.T) {
			now := time.Now()
			env := NewEvent(
				"100",
				EventA1,
				now,
			)

			xtesting.Expect(
				t,
				"unexpected envelope",
				env,
				&Envelope{
					MessageID:         "100",
					CorrelationID:     "100",
					CausationID:       "100",
					Message:           EventA1,
					CreatedAt:         now,
					EventStreamID:     "87d0e883-8f15-5eaf-9601-fe3a7dd517a4",
					EventStreamOffset: 0,
				},
			)
		})
	})

	t.Run("func (Envelope) NewCommand()", func(t *testing.T) {
		t.Run("it returns the expected envelope", func(t *testing.T) {
			handler := runtimeconfig.FromProcess(&ProcessMessageHandlerStub[*ProcessRootStub]{
				ConfigureFunc: func(c dogma.ProcessConfigurer) {
					c.Identity("<handler>", "d1c7e18a-4d72-4705-a120-6cfb29eef655")
					c.Routes(
						dogma.HandlesEvent[*EventStub[TypeA]](),
						dogma.ExecutesCommand[*CommandStub[TypeA]](),
					)
				},
			})

			parent := NewEvent(
				"100",
				EventP1,
				time.Now(),
			)
			origin := Origin{
				Handler:     handler,
				HandlerType: config.ProcessHandlerType,
				InstanceID:  "<instance>",
			}
			now := time.Now()
			child := parent.NewCommand(
				"200",
				CommandA1,
				now,
				origin,
			)

			xtesting.Expect(
				t,
				"unexpected envelope",
				child,
				&Envelope{
					MessageID:     "200",
					CorrelationID: "100",
					CausationID:   "100",
					Message:       CommandA1,
					CreatedAt:     now,
					Origin:        &origin,
				},
			)
		})
	})

	t.Run("func (Envelope) NewEvent()", func(t *testing.T) {
		t.Run("it returns the expected envelope", func(t *testing.T) {
			handler := runtimeconfig.FromAggregate(&AggregateMessageHandlerStub[*AggregateRootStub]{
				ConfigureFunc: func(c dogma.AggregateConfigurer) {
					c.Identity("<handler>", "8688dc39-b5d0-4468-89fd-0d9452667c0c")
					c.Routes(
						dogma.HandlesCommand[*CommandStub[TypeA]](),
						dogma.RecordsEvent[*EventStub[TypeA]](),
					)
				},
			})

			const streamID = "10208426-7df8-4f47-ac2a-a83f55c3b1c0"
			const offset = 42

			parent := NewCommand(
				"100",
				CommandP1,
				time.Now(),
			)
			origin := Origin{
				Handler:     handler,
				HandlerType: config.AggregateHandlerType,
				InstanceID:  "<instance>",
			}
			now := time.Now()
			child := parent.NewEvent(
				"200",
				EventA1,
				now,
				origin,
				streamID,
				offset,
			)

			xtesting.Expect(
				t,
				"unexpected envelope",
				child,
				&Envelope{
					MessageID:         "200",
					CorrelationID:     "100",
					CausationID:       "100",
					Message:           EventA1,
					CreatedAt:         now,
					Origin:            &origin,
					EventStreamID:     streamID,
					EventStreamOffset: offset,
				},
			)
		})
	})

	t.Run("func (Envelope) NewDeadline()", func(t *testing.T) {
		t.Run("it returns the expected envelope", func(t *testing.T) {
			handler := runtimeconfig.FromProcess(&ProcessMessageHandlerStub[*ProcessRootStub]{
				ConfigureFunc: func(c dogma.ProcessConfigurer) {
					c.Identity("<handler>", "1d4e3d22-52fe-4b1b-9bf5-44b2050c08c2")
					c.Routes(
						dogma.HandlesEvent[*EventStub[TypeA]](),
						dogma.ExecutesCommand[*CommandStub[TypeA]](),
						dogma.SchedulesDeadline[*DeadlineStub[TypeA]](),
					)
				},
			})

			parent := NewCommand(
				"100",
				CommandP1,
				time.Now(),
			)
			origin := Origin{
				Handler:     handler,
				HandlerType: config.ProcessHandlerType,
				InstanceID:  "<instance>",
			}
			now := time.Now()
			s := time.Now()
			child := parent.NewDeadline(
				"200",
				DeadlineA1,
				now,
				s,
				origin,
			)

			xtesting.Expect(
				t,
				"unexpected envelope",
				child,
				&Envelope{
					MessageID:     "200",
					CorrelationID: "100",
					CausationID:   "100",
					Message:       DeadlineA1,
					CreatedAt:     now,
					ScheduledFor:  s,
					Origin:        &origin,
				},
			)
		})
	})
}
