package integration

import (
	"fmt"
	"time"

	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/testkit/engine/internal/panicx"
	"github.com/dogmatiq/testkit/envelope"
	"github.com/dogmatiq/testkit/fact"
	"github.com/dogmatiq/testkit/location"
)

// scope is an implementation of dogma.IntegrationCommandScope.
type scope struct {
	config     configkit.RichIntegration
	messageIDs *envelope.MessageIDGenerator
	observer   fact.Observer
	now        time.Time
	command    *envelope.Envelope
	events     []*envelope.Envelope
}

func (s *scope) RecordEvent(m dogma.Message) {
	if !s.config.MessageTypes().Produced.HasM(m) {
		panic(panicx.UnexpectedBehavior{
			Handler:        s.config,
			Interface:      "IntegrationMessageHandler",
			Method:         "HandleCommand",
			Implementation: s.config.Handler(),
			Message:        s.command.Message,
			Description:    fmt.Sprintf("recorded an event of type %T, which is not produced by this handler", m),
			Location:       location.OfCall(),
		})
	}

	if err := dogma.ValidateMessage(m); err != nil {
		panic(panicx.UnexpectedBehavior{
			Handler:        s.config,
			Interface:      "IntegrationMessageHandler",
			Method:         "HandleCommand",
			Implementation: s.config.Handler(),
			Message:        s.command.Message,
			Description:    fmt.Sprintf("recorded an invalid %T event: %s", m, err),
			Location:       location.OfCall(),
		})
	}

	env := s.command.NewEvent(
		s.messageIDs.Next(),
		m,
		s.now,
		envelope.Origin{
			Handler:     s.config,
			HandlerType: configkit.IntegrationHandlerType,
		},
	)

	s.events = append(s.events, env)

	s.observer.Notify(fact.EventRecordedByIntegration{
		Handler:       s.config,
		Envelope:      s.command,
		EventEnvelope: env,
	})
}

func (s *scope) Log(f string, v ...interface{}) {
	s.observer.Notify(fact.MessageLoggedByIntegration{
		Handler:      s.config,
		Envelope:     s.command,
		LogFormat:    f,
		LogArguments: v,
	})
}
