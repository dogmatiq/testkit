package envelope_test

import (
	"testing"
	"time"
	"uuid"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/enginekit/config"
	"github.com/dogmatiq/enginekit/config/runtimeconfig"
	. "github.com/dogmatiq/enginekit/enginetest/stubs"
	. "github.com/dogmatiq/testkit/envelope"
	"github.com/dogmatiq/testkit/internal/x/xtesting"
)

var (
	messageID = uuid.MustParse("5e8b1d64-0c3a-4f92-b7e1-8a6d2c9f3b05")
	parentID  = uuid.MustParse("7a0f3c52-9e1d-4b86-a3c4-5d2e8f1b6c09")
	childID   = uuid.MustParse("c41e9d07-2b6a-4f38-8e5d-9a1c3b7f2e64")
)

func TestEnvelope(t *testing.T) {
	t.Run("func NewCommand()", func(t *testing.T) {
		t.Run("it returns the expected envelope", func(t *testing.T) {
			now := time.Now()
			env := NewCommand(
				messageID,
				CommandA1,
				now,
			)

			xtesting.Expect(
				t,
				"unexpected envelope",
				env,
				&Envelope{
					MessageID:     messageID,
					CorrelationID: messageID,
					CausationID:   messageID,
					Message:       CommandA1,
					CreatedAt:     now,
				},
			)
		})
	})

	t.Run("func NewEvent()", func(t *testing.T) {
		t.Run("it returns the expected envelope", func(t *testing.T) {
			now := time.Now()
			streamID := uuid.MustParse("0b6e4f2a-9c1d-4e83-a5f7-3d2b8c6e1f90")
			env := NewEvent(
				messageID,
				EventA1,
				now,
				streamID,
			)

			xtesting.Expect(
				t,
				"unexpected envelope",
				env,
				&Envelope{
					MessageID:         messageID,
					CorrelationID:     messageID,
					CausationID:       messageID,
					Message:           EventA1,
					CreatedAt:         now,
					EventStreamID:     streamID,
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
				parentID,
				EventP1,
				time.Now(),
				uuid.MustParse("0b6e4f2a-9c1d-4e83-a5f7-3d2b8c6e1f90"),
			)
			origin := Origin{
				Handler:     handler,
				HandlerType: config.ProcessHandlerType,
				InstanceID:  "<instance>",
			}
			now := time.Now()
			child := parent.NewCommand(
				childID,
				CommandA1,
				now,
				origin,
			)

			xtesting.Expect(
				t,
				"unexpected envelope",
				child,
				&Envelope{
					MessageID:     childID,
					CorrelationID: parentID,
					CausationID:   parentID,
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

			streamID := uuid.MustParse("10208426-7df8-4f47-ac2a-a83f55c3b1c0")
			const offset = 42

			parent := NewCommand(
				parentID,
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
				childID,
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
					MessageID:         childID,
					CorrelationID:     parentID,
					CausationID:       parentID,
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
				parentID,
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
				childID,
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
					MessageID:     childID,
					CorrelationID: parentID,
					CausationID:   parentID,
					Message:       DeadlineA1,
					CreatedAt:     now,
					ScheduledFor:  s,
					Origin:        &origin,
				},
			)
		})
	})
}
