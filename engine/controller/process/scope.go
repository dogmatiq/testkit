package process

import (
	"time"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/dogmatest/engine/envelope"
	"github.com/dogmatiq/dogmatest/engine/fact"
	"github.com/dogmatiq/dogmatest/internal/enginekit/handler"
)

// scope is an implementation of dogma.ProcessEventScope.
type scope struct {
	id       string
	name     string
	observer fact.Observer
	root     dogma.ProcessRoot
	exists   bool
	event    *envelope.Envelope
	commands []*envelope.Envelope
	timeouts []*envelope.Envelope
}

func (s *scope) InstanceID() string {
	return s.id
}

func (s *scope) Begin() bool {
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

func (s *scope) End() {
	if !s.exists {
		panic("can not end non-existent instance")
	}

	s.exists = false
	s.timeouts = nil

	s.observer.Notify(fact.ProcessInstanceEnded{
		HandlerName: s.name,
		InstanceID:  s.id,
		Root:        s.root,
		Envelope:    s.event,
	})
}

func (s *scope) Root() dogma.ProcessRoot {
	if !s.exists {
		panic("can not access process root of non-existent instance")
	}

	return s.root
}

func (s *scope) ExecuteCommand(m dogma.Message) {
	if !s.exists {
		panic("can not execute command against non-existent instance")
	}

	env := s.event.NewCommand(
		m,
		envelope.Origin{
			HandlerName: s.name,
			HandlerType: handler.ProcessType,
			InstanceID:  s.id,
		},
	)

	s.commands = append(s.commands, env)

	s.observer.Notify(fact.CommandExecutedByProcess{
		HandlerName:     s.name,
		InstanceID:      s.id,
		Root:            s.root,
		Envelope:        s.event,
		CommandEnvelope: env,
	})
}

func (s *scope) ScheduleTimeout(m dogma.Message, t time.Time) {
	if !s.exists {
		panic("can not schedule timeout against non-existent instance")
	}

	env := s.event.NewTimeout(
		m,
		t,
		envelope.Origin{
			HandlerName: s.name,
			HandlerType: handler.ProcessType,
			InstanceID:  s.id,
		},
	)

	s.timeouts = append(s.timeouts, env)

	s.observer.Notify(fact.TimeoutScheduledByProcess{
		HandlerName:     s.name,
		InstanceID:      s.id,
		Root:            s.root,
		Envelope:        s.event,
		TimeoutEnvelope: env,
	})
}

func (s *scope) Log(f string, v ...interface{}) {
	s.observer.Notify(fact.MessageLoggedByProcess{
		HandlerName:  s.name,
		InstanceID:   s.id,
		Root:         s.root,
		Envelope:     s.event,
		LogFormat:    f,
		LogArguments: v,
	})
}
