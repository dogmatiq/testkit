package integration

import (
	"fmt"
	"time"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/enginekit/handler"
	"github.com/dogmatiq/enginekit/message"
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
	now        time.Time
	produced   message.TypeContainer
	command    *envelope.Envelope
	events     []*envelope.Envelope
}

func (s *scope) RecordEvent(m dogma.Message) {
	if !s.produced.HasM(m) {
		panic(fmt.Sprintf(
			"the '%s' handler is not configured to record events of type %T",
			s.name,
			m,
		))
	}

	env := s.command.NewEvent(
		s.messageIDs.Next(),
		m,
		s.now,
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
