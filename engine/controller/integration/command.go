package integration

import (
	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/dogmatest/engine/envelope"
	"github.com/dogmatiq/dogmatest/engine/fact"
)

// commandScope is an implementation of dogma.IntegrationCommandScope.
type commandScope struct {
	id       string
	name     string
	observer fact.Observer
	command  *envelope.Envelope
	children []*envelope.Envelope
}

func (s *commandScope) RecordEvent(m dogma.Message) {
	env := s.command.NewEvent(m)
	s.children = append(s.children, env)

	s.observer.Notify(fact.EventRecordedByIntegration{
		HandlerName:   s.name,
		Envelope:      s.command,
		EventEnvelope: env,
	})
}

func (s *commandScope) Log(f string, v ...interface{}) {
	s.observer.Notify(fact.MessageLoggedByIntegration{
		HandlerName:  s.name,
		Envelope:     s.command,
		LogFormat:    f,
		LogArguments: v,
	})
}
