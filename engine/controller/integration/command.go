package integration

import (
	"time"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/dogmatest/engine/envelope"
	"github.com/dogmatiq/dogmatest/engine/fact"
	"github.com/dogmatiq/dogmatest/internal/enginekit/message"
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
	env := s.command.NewChild(m, message.EventRole, time.Time{})
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
