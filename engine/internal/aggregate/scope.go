package aggregate

import (
	"fmt"
	"time"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/enginekit/config"
	"github.com/dogmatiq/enginekit/message"
	"github.com/dogmatiq/testkit/engine/internal/panicx"
	"github.com/dogmatiq/testkit/envelope"
	"github.com/dogmatiq/testkit/fact"
	"github.com/dogmatiq/testkit/internal/validation"
	"github.com/dogmatiq/testkit/location"
)

// scope is an implementation of dogma.AggregateCommandScope.
type scope struct {
	instanceID string
	config     *config.Aggregate
	messageIDs *envelope.MessageIDGenerator
	observer   fact.Observer
	root       dogma.AggregateRoot
	now        time.Time
	command    *envelope.Envelope
	streamID   string
	offset     uint64
	events     []*envelope.Envelope
}

func (s *scope) InstanceID() string {
	return s.instanceID
}

func (s *scope) RecordEvent(m dogma.Event) {
	mt := message.TypeOf(m)

	if !s.config.RouteSet().DirectionOf(mt).Has(config.OutboundDirection) {
		panic(panicx.UnexpectedBehavior{
			Handler:        s.config,
			Interface:      "AggregateMessageHandler",
			Method:         "HandleCommand",
			Implementation: s.config.Source.Get(),
			Message:        s.command.Message,
			Description:    fmt.Sprintf("recorded an event of type %s, which is not produced by this handler", mt),
			Location:       location.OfCall(),
		})
	}

	if err := m.Validate(validation.EventValidationScope()); err != nil {
		panic(panicx.UnexpectedBehavior{
			Handler:        s.config,
			Interface:      "AggregateMessageHandler",
			Method:         "HandleCommand",
			Implementation: s.config.Source.Get(),
			Message:        s.command.Message,
			Description:    fmt.Sprintf("recorded an invalid %s event: %s", mt, err),
			Location:       location.OfCall(),
		})
	}

	if s.offset == 0 {
		s.observer.Notify(fact.AggregateInstanceCreated{
			Handler:    s.config,
			InstanceID: s.instanceID,
			Root:       s.root,
			Envelope:   s.command,
		})
	}

	panicx.EnrichUnexpectedMessage(
		s.config,
		"AggregateRoot",
		"ApplyEvent",
		s.root,
		m,
		func() {
			s.root.ApplyEvent(m)
		},
	)

	env := s.command.NewEvent(
		s.messageIDs.Next(),
		m,
		s.now,
		envelope.Origin{
			Handler:     s.config,
			HandlerType: config.AggregateHandlerType,
			InstanceID:  s.instanceID,
		},
		s.streamID,
		s.offset,
	)

	s.events = append(s.events, env)
	s.offset++

	s.observer.Notify(fact.EventRecordedByAggregate{
		Handler:       s.config,
		InstanceID:    s.instanceID,
		Root:          s.root,
		Envelope:      s.command,
		EventEnvelope: env,
	})
}

func (s *scope) Now() time.Time {
	return s.now
}

func (s *scope) Log(f string, v ...any) {
	s.observer.Notify(fact.MessageLoggedByAggregate{
		Handler:      s.config,
		InstanceID:   s.instanceID,
		Root:         s.root,
		Envelope:     s.command,
		LogFormat:    f,
		LogArguments: v,
	})
}
