package aggregate

import (
	"time"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/dogmatest/engine/controller"
	"github.com/dogmatiq/dogmatest/engine/envelope"
	"github.com/dogmatiq/dogmatest/engine/fact"
	"github.com/dogmatiq/dogmatest/internal/enginekit/message"
)

// commandScope is an implementation of dogma.AggregateCommandScope.
type commandScope struct {
	id       string
	name     string
	parent   controller.Scope
	root     dogma.AggregateRoot
	exists   bool
	command  *envelope.Envelope
	children []*envelope.Envelope
}

func (s *commandScope) InstanceID() string {
	return s.id
}

func (s *commandScope) Create() bool {
	if s.exists {
		return false
	}

	s.exists = true

	s.parent.RecordFacts(fact.AggregateInstanceCreated{
		HandlerName: s.name,
		InstanceID:  s.id,
		Root:        s.root,
		Envelope:    s.command,
	})

	return true
}

func (s *commandScope) Destroy() {
	if !s.exists {
		panic("can not destroy non-existent instance")
	}

	s.exists = false

	s.parent.RecordFacts(fact.AggregateInstanceDestroyed{
		HandlerName: s.name,
		InstanceID:  s.id,
		Root:        s.root,
		Envelope:    s.command,
	})
}

func (s *commandScope) Root() dogma.AggregateRoot {
	if !s.exists {
		panic("can not access aggregate root of non-existent instance")
	}

	return s.root
}

func (s *commandScope) RecordEvent(m dogma.Message) {
	if !s.exists {
		panic("can not record event against non-existent instance")
	}

	s.root.ApplyEvent(m)

	env := s.command.NewChild(m, message.EventRole, time.Time{})
	s.children = append(s.children, env)

	s.parent.RecordFacts(fact.EventRecordedByAggregate{
		HandlerName: s.name,
		InstanceID:  s.id,
		Root:        s.root,
		Envelope:    s.command,
	})
}

func (s *commandScope) Log(f string, v ...interface{}) {
	s.parent.RecordFacts(fact.MessageLoggedByAggregate{
		HandlerName:  s.name,
		InstanceID:   s.id,
		Root:         s.root,
		Envelope:     s.command,
		LogFormat:    f,
		LogArguments: v,
	})
}
