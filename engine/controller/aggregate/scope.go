package aggregate

import (
	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/dogmatest/engine/envelope"
	"github.com/dogmatiq/dogmatest/engine/fact"
	"github.com/dogmatiq/enginekit/handler"
)

// scope is an implementation of dogma.AggregateCommandScope.
type scope struct {
	id        string
	name      string
	handler   dogma.AggregateMessageHandler
	observer  fact.Observer
	root      dogma.AggregateRoot
	exists    bool
	created   bool // true if Create() returned true at least once
	destroyed bool // true if Destroy() returned true at least once
	command   *envelope.Envelope
	events    []*envelope.Envelope
}

func (s *scope) InstanceID() string {
	return s.id
}

func (s *scope) Create() bool {
	if s.exists {
		return false
	}

	s.exists = true
	s.created = true

	s.observer.Notify(fact.AggregateInstanceCreated{
		HandlerName: s.name,
		Handler:     s.handler,
		InstanceID:  s.id,
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
		HandlerName: s.name,
		Handler:     s.handler,
		InstanceID:  s.id,
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

	s.root.ApplyEvent(m)

	env := s.command.NewEvent(
		m,
		envelope.Origin{
			HandlerName: s.name,
			HandlerType: handler.AggregateType,
			InstanceID:  s.id,
		},
	)

	s.events = append(s.events, env)

	s.observer.Notify(fact.EventRecordedByAggregate{
		HandlerName:   s.name,
		Handler:       s.handler,
		InstanceID:    s.id,
		Root:          s.root,
		Envelope:      s.command,
		EventEnvelope: env,
	})
}

func (s *scope) Log(f string, v ...interface{}) {
	s.observer.Notify(fact.MessageLoggedByAggregate{
		HandlerName:  s.name,
		Handler:      s.handler,
		InstanceID:   s.id,
		Root:         s.root,
		Envelope:     s.command,
		LogFormat:    f,
		LogArguments: v,
	})
}
