package aggregate

import (
	"fmt"
	"time"

	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/testkit/engine/envelope"
	"github.com/dogmatiq/testkit/engine/fact"
	"github.com/dogmatiq/testkit/engine/panicx"
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
		panic(fmt.Sprintf(
			"the '%s' handler is not configured to record events of type %T",
			s.config.Identity().Name,
			m,
		))
	}

	if err := dogma.ValidateMessage(m); err != nil {
		panic(fmt.Sprintf(
			"can not record event of type %T, it is invalid: %s",
			m,
			err,
		))
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
