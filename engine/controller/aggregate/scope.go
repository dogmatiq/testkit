package aggregate

import (
	"fmt"
	"time"

	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/enginekit/identity"
	"github.com/dogmatiq/configkit/message"
	"github.com/dogmatiq/testkit/engine/envelope"
	"github.com/dogmatiq/testkit/engine/fact"
)

// scope is an implementation of dogma.AggregateCommandScope.
type scope struct {
	instanceID string
	identity   identity.Identity
	handler    dogma.AggregateMessageHandler
	messageIDs *envelope.MessageIDGenerator
	observer   fact.Observer
	root       dogma.AggregateRoot
	now        time.Time
	exists     bool
	created    bool // true if Create() returned true at least once
	destroyed  bool // true if Destroy() returned true at least once
	produced   message.TypeCollection
	command    *envelope.Envelope
	events     []*envelope.Envelope
}

func (s *scope) InstanceID() string {
	return s.instanceID
}

func (s *scope) Create() bool {
	if s.exists {
		return false
	}

	s.exists = true
	s.created = true

	s.observer.Notify(fact.AggregateInstanceCreated{
		HandlerName: s.identity.Name,
		Handler:     s.handler,
		InstanceID:  s.instanceID,
		Root:        s.root,
		Envelope:    s.command,
	})

	return true
}

func (s *scope) Destroy() {
	if !s.exists {
		panic("can not destroy non-existent instance")
	}

	s.exists = false
	s.destroyed = true

	s.observer.Notify(fact.AggregateInstanceDestroyed{
		HandlerName: s.identity.Name,
		Handler:     s.handler,
		InstanceID:  s.instanceID,
		Root:        s.root,
		Envelope:    s.command,
	})
}

func (s *scope) Root() dogma.AggregateRoot {
	if !s.exists {
		panic("can not access aggregate root of non-existent instance")
	}

	return s.root
}

func (s *scope) RecordEvent(m dogma.Message) {
	if !s.exists {
		panic("can not record event against non-existent instance")
	}

	if !s.produced.HasM(m) {
		panic(fmt.Sprintf(
			"the '%s' handler is not configured to record events of type %T",
			s.identity.Name,
			m,
		))
	}

	s.root.ApplyEvent(m)

	env := s.command.NewEvent(
		s.messageIDs.Next(),
		m,
		s.now,
		envelope.Origin{
			HandlerName: s.identity.Name,
			HandlerType: configkit.AggregateHandlerType,
			InstanceID:  s.instanceID,
		},
	)

	s.events = append(s.events, env)

	s.observer.Notify(fact.EventRecordedByAggregate{
		HandlerName:   s.identity.Name,
		Handler:       s.handler,
		InstanceID:    s.instanceID,
		Root:          s.root,
		Envelope:      s.command,
		EventEnvelope: env,
	})
}

func (s *scope) Log(f string, v ...interface{}) {
	s.observer.Notify(fact.MessageLoggedByAggregate{
		HandlerName:  s.identity.Name,
		Handler:      s.handler,
		InstanceID:   s.instanceID,
		Root:         s.root,
		Envelope:     s.command,
		LogFormat:    f,
		LogArguments: v,
	})
}
