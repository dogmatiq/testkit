package process

import (
	"time"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/dogmatest/engine/envelope"
	"github.com/dogmatiq/dogmatest/engine/fact"
)

// eventScope is an implementation of dogma.ProcessEventScope.
type eventScope struct {
	id       string
	name     string
	observer fact.Observer
	root     dogma.ProcessRoot
	exists   bool
	event    *envelope.Envelope
	children []*envelope.Envelope
}

func (s *eventScope) InstanceID() string {
	return s.id
}

func (s *eventScope) Begin() bool {
	if s.exists {
		return false
	}

	s.exists = true

	s.observer.Notify(fact.ProcessInstanceBegun{
		HandlerName: s.name,
		InstanceID:  s.id,
		Root:        s.root,
		Envelope:    s.event,
	})

	return true
}

func (s *eventScope) End() {
	if !s.exists {
		panic("can not end non-existent instance")
	}

	s.exists = false

	s.observer.Notify(fact.ProcessInstanceEnded{
		HandlerName: s.name,
		InstanceID:  s.id,
		Root:        s.root,
		Envelope:    s.event,
	})
}

func (s *eventScope) Root() dogma.ProcessRoot {
	if !s.exists {
		panic("can not access process root of non-existent instance")
	}

	return s.root
}

func (s *eventScope) ExecuteCommand(m dogma.Message) {
	if !s.exists {
		panic("can not execute command against non-existent instance")
	}

	env := s.event.NewCommand(m)
	s.children = append(s.children, env)

	s.observer.Notify(fact.CommandExecutedByProcess{
		HandlerName:     s.name,
		InstanceID:      s.id,
		Root:            s.root,
		Envelope:        s.event,
		CommandEnvelope: env,
	})
}

func (s *eventScope) ScheduleTimeout(m dogma.Message, t time.Time) {
	if !s.exists {
		panic("can not schedule timeout against non-existent instance")
	}

	env := s.event.NewTimeout(m, t)
	s.children = append(s.children, env)

	s.observer.Notify(fact.TimeoutScheduledByProcess{
		HandlerName:     s.name,
		InstanceID:      s.id,
		Root:            s.root,
		Envelope:        s.event,
		TimeoutEnvelope: env,
	})
}

func (s *eventScope) Log(f string, v ...interface{}) {
	s.observer.Notify(fact.MessageLoggedByProcess{
		HandlerName:  s.name,
		InstanceID:   s.id,
		Root:         s.root,
		Envelope:     s.event,
		LogFormat:    f,
		LogArguments: v,
	})
}
