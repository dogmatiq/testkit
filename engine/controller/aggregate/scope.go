package aggregate

import (
	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/dogmatest/engine/controller"
	"github.com/dogmatiq/dogmatest/engine/envelope"
	"github.com/dogmatiq/dogmatest/engine/fact"
	"github.com/dogmatiq/dogmatest/internal/enginekit/message"
)

// scope is an implementation of dogma.AggregateCommandScope.
type scope struct {
	id      string
	name    string
	parent  controller.Scope
	root    dogma.AggregateRoot
	exists  bool
	command *envelope.Envelope
	events  []*envelope.Envelope
}

func (s *scope) InstanceID() string {
	return s.id
}

func (s *scope) Create() bool {
	if s.exists {
		return false
	}

	s.exists = true

	s.parent.RecordFacts(fact.AggregateInstanceCreated{
		HandlerName:     s.name,
		InstanceID:      s.id,
		Root:            s.root,
		CommandEnvelope: s.command,
	})

	return true
}

func (s *scope) Destroy() {
	if !s.exists {
		panic("can not destroy non-existent instance")
	}

	s.exists = false

	s.parent.RecordFacts(fact.AggregateInstanceDestroyed{
		HandlerName:     s.name,
		InstanceID:      s.id,
		Root:            s.root,
		CommandEnvelope: s.command,
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

	s.root.ApplyEvent(m)

	env := s.command.NewChild(m, message.EventRole)
	s.events = append(s.events, env)

	s.parent.RecordFacts(fact.EventRecordedByAggregate{
		HandlerName:     s.name,
		InstanceID:      s.id,
		Root:            s.root,
		CommandEnvelope: s.command,
	})
}

func (s *scope) Log(f string, v ...interface{}) {
	s.parent.RecordFacts(fact.AggregateLoggedMessage{
		HandlerName:     s.name,
		InstanceID:      s.id,
		Root:            s.root,
		CommandEnvelope: s.command,
		LogFormat:       f,
		LogArguments:    v,
	})
}
