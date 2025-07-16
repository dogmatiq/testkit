package integration

import (
	"fmt"
	"time"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/enginekit/config"
	"github.com/dogmatiq/enginekit/message"
	"github.com/dogmatiq/testkit/engine/internal/panicx"
	"github.com/dogmatiq/testkit/envelope"
	"github.com/dogmatiq/testkit/fact"
	"github.com/dogmatiq/testkit/internal/validation"
	"github.com/dogmatiq/testkit/location"
)

// scope is an implementation of dogma.IntegrationCommandScope.
type scope struct {
	config     *config.Integration
	messageIDs *envelope.MessageIDGenerator
	observer   fact.Observer
	now        time.Time
	command    *envelope.Envelope
	events     []*envelope.Envelope
}

func (s *scope) RecordEvent(m dogma.Event) {
	mt := message.TypeOf(m)

	if !s.config.RouteSet().DirectionOf(mt).Has(config.OutboundDirection) {
		panic(panicx.UnexpectedBehavior{
			Handler:        s.config,
			Interface:      "IntegrationMessageHandler",
			Method:         "HandleCommand",
			Implementation: s.config.Source.Get(),
			Message:        s.command.Message,
			Description:    fmt.Sprintf("recorded an event of type %s, which is not produced by this handler", mt),
			Location:       location.OfCall(),
		})
	}

	if err := m.Validate(validation.EventValidationScope()); err != nil {
		panic(panicx.UnexpectedBehavior{
			Handler:        s.config,
			Interface:      "IntegrationMessageHandler",
			Method:         "HandleCommand",
			Implementation: s.config.Source.Get(),
			Message:        s.command.Message,
			Description:    fmt.Sprintf("recorded an invalid %s event: %s", mt, err),
			Location:       location.OfCall(),
		})
	}

	env := s.command.NewEvent(
		s.messageIDs.Next(),
		m,
		s.now,
		envelope.Origin{
			Handler:     s.config,
			HandlerType: config.IntegrationHandlerType,
		},
	)

	s.events = append(s.events, env)

	s.observer.Notify(fact.EventRecordedByIntegration{
		Handler:       s.config,
		Envelope:      s.command,
		EventEnvelope: env,
	})
}

func (s *scope) Now() time.Time {
	return s.now
}

func (s *scope) Log(f string, v ...any) {
	s.observer.Notify(fact.MessageLoggedByIntegration{
		Handler:      s.config,
		Envelope:     s.command,
		LogFormat:    f,
		LogArguments: v,
	})
}
