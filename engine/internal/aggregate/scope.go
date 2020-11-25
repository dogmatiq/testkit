package aggregate

import (
	"fmt"
	"time"

	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/testkit/engine/envelope"
	"github.com/dogmatiq/testkit/engine/fact"
	"github.com/dogmatiq/testkit/engine/panicx"
	"github.com/dogmatiq/testkit/internal/location"
)

// scope is an implementation of dogma.AggregateCommandScope.
type scope struct {
	instanceID string
	config     configkit.RichAggregate
	messageIDs *envelope.MessageIDGenerator
	observer   fact.Observer
	root       dogma.AggregateRoot
	now        time.Time
	exists     bool
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

	s.root = s.config.Handler().New()
	s.exists = false

	s.observer.Notify(fact.AggregateInstanceDestroyed{
		Handler:    s.config,
		InstanceID: s.instanceID,
		Root:       s.root,
		Envelope:   s.command,
	})
}

func (s *scope) RecordEvent(m dogma.Message) {
	if !s.config.MessageTypes().Produced.HasM(m) {
		panic(panicx.UnexpectedBehavior{
			Handler:        s.config,
			Interface:      "AggregateMessageHandler",
			Method:         "HandleCommand",
			Implementation: s.config.Handler(),
			Message:        s.command.Message,
			Description:    fmt.Sprintf("recorded an event of type %T, which is not produced by this handler", m),
			Location:       location.OfCall(),
		})
	}

	if err := dogma.ValidateMessage(m); err != nil {
		panic(panicx.UnexpectedBehavior{
			Handler:        s.config,
			Interface:      "AggregateMessageHandler",
			Method:         "HandleCommand",
			Implementation: s.config.Handler(),
			Message:        s.command.Message,
			Description:    fmt.Sprintf("recorded an invalid %T event: %s", m, err),
			Location:       location.OfCall(),
		})
	}

	if !s.exists {
		s.observer.Notify(fact.AggregateInstanceCreated{
			Handler:    s.config,
			InstanceID: s.instanceID,
			Root:       s.root,
			Envelope:   s.command,
		})

		s.exists = true
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
			HandlerType: configkit.AggregateHandlerType,
			InstanceID:  s.instanceID,
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

func (s *scope) Log(f string, v ...interface{}) {
	s.observer.Notify(fact.MessageLoggedByAggregate{
		Handler:      s.config,
		InstanceID:   s.instanceID,
		Root:         s.root,
		Envelope:     s.command,
		LogFormat:    f,
		LogArguments: v,
	})
}
