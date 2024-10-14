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
	exists     bool
	destroyed  bool
	command    *envelope.Envelope
	events     []*envelope.Envelope
}

func (s *scope) InstanceID() string {
	return s.instanceID
}

func (s *scope) Destroy() {
	if !s.exists {
		return
	}

	s.root = s.config.Interface().New()
	s.exists = false
	s.destroyed = true

	s.observer.Notify(fact.AggregateInstanceDestroyed{
		Handler:    s.config,
		InstanceID: s.instanceID,
		Root:       s.root,
		Envelope:   s.command,
	})
}

func (s *scope) RecordEvent(m dogma.Event) {
	mt := message.TypeOf(m)

	if !s.config.RouteSet().DirectionOf(mt).Has(config.OutboundDirection) {
		panic(panicx.UnexpectedBehavior{
			Handler:        s.config,
			Interface:      "AggregateMessageHandler",
			Method:         "HandleCommand",
			Implementation: s.config.Interface(),
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
			Implementation: s.config.Interface(),
			Message:        s.command.Message,
			Description:    fmt.Sprintf("recorded an invalid %s event: %s", mt, err),
			Location:       location.OfCall(),
		})
	}

	if !s.exists {
		if s.destroyed {
			s.observer.Notify(fact.AggregateInstanceDestructionReverted{
				Handler:    s.config,
				InstanceID: s.instanceID,
				Root:       s.root,
				Envelope:   s.command,
			})
		} else {
			s.observer.Notify(fact.AggregateInstanceCreated{
				Handler:    s.config,
				InstanceID: s.instanceID,
				Root:       s.root,
				Envelope:   s.command,
			})
		}

		s.exists = true
		s.destroyed = false
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
			Handler:    s.config,
			InstanceID: s.instanceID,
		},
	)

	s.events = append(s.events, env)

	s.observer.Notify(fact.EventRecordedByAggregate{
		Handler:       s.config,
		InstanceID:    s.instanceID,
		Root:          s.root,
		Envelope:      s.command,
		EventEnvelope: env,
	})
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
