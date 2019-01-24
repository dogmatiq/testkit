package integration

import (
	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/dogmatest/engine/envelope"
	"github.com/dogmatiq/dogmatest/engine/fact"
	"github.com/dogmatiq/dogmatest/internal/enginekit/handler"
)

// scope is an implementation of dogma.IntegrationCommandScope.
type scope struct {
	id       string
	name     string
	observer fact.Observer
	command  *envelope.Envelope
	events   []*envelope.Envelope
}

func (s *scope) RecordEvent(m dogma.Message) {
	env := s.command.NewEvent(
		m,
		envelope.Origin{
			HandlerName: s.name,
			HandlerType: handler.IntegrationType,
		},
	)

	s.events = append(s.events, env)

	s.observer.Notify(fact.EventRecordedByIntegration{
		HandlerName:   s.name,
		Envelope:      s.command,
		EventEnvelope: env,
	})
}

func (s *scope) Log(f string, v ...interface{}) {
	s.observer.Notify(fact.MessageLoggedByIntegration{
		HandlerName:  s.name,
		Envelope:     s.command,
		LogFormat:    f,
		LogArguments: v,
	})
}
