package aggregate

import (
	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/dogmatest/engine/envelope"
	"github.com/dogmatiq/dogmatest/engine/fact"
)

// commandScope is an implementation of dogma.AggregateCommandScope.
type commandScope struct {
	id       string
	name     string
	observer fact.Observer
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

	s.observer.Notify(fact.AggregateInstanceCreated{
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

	s.observer.Notify(fact.AggregateInstanceDestroyed{
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

	env := s.command.NewEvent(m)
	s.children = append(s.children, env)

	s.observer.Notify(fact.EventRecordedByAggregate{
		HandlerName:   s.name,
		InstanceID:    s.id,
		Root:          s.root,
		Envelope:      s.command,
		EventEnvelope: env,
	})
}

func (s *commandScope) Log(f string, v ...interface{}) {
	s.observer.Notify(fact.MessageLoggedByAggregate{
		HandlerName:  s.name,
		InstanceID:   s.id,
		Root:         s.root,
		Envelope:     s.command,
		LogFormat:    f,
		LogArguments: v,
	})
}
