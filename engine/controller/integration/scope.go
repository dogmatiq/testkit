package integration

import (
	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/enginekit/handler"
	"github.com/dogmatiq/testkit/engine/envelope"
	"github.com/dogmatiq/testkit/engine/fact"
)

// scope is an implementation of dogma.IntegrationCommandScope.
type scope struct {
	id         string
	name       string
	handler    dogma.IntegrationMessageHandler
	messageIDs *envelope.MessageIDGenerator
	observer   fact.Observer
	command    *envelope.Envelope
	events     []*envelope.Envelope
}

func (s *scope) RecordEvent(m dogma.Message) {
	env := s.command.NewEvent(
		s.messageIDs.Next(),
		m,
		envelope.Origin{
			HandlerName: s.name,
			HandlerType: handler.IntegrationType,
		},
	)

	s.events = append(s.events, env)

	s.observer.Notify(fact.EventRecordedByIntegration{
		HandlerName:   s.name,
		Handler:       s.handler,
		Envelope:      s.command,
		EventEnvelope: env,
	})
}

func (s *scope) Log(f string, v ...interface{}) {
	s.observer.Notify(fact.MessageLoggedByIntegration{
		HandlerName:  s.name,
		Handler:      s.handler,
		Envelope:     s.command,
		LogFormat:    f,
		LogArguments: v,
	})
}
