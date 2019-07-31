package projection

import (
	"time"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/testkit/engine/envelope"
	"github.com/dogmatiq/testkit/engine/fact"
)

// scope is an implementation of dogma.ProjectionEventScope.
type scope struct {
	name     string
	handler  dogma.ProjectionMessageHandler
	observer fact.Observer
	event    *envelope.Envelope
}

func (s *scope) RecordedAt() time.Time {
	return s.event.CreatedAt
}

func (s *scope) Log(f string, v ...interface{}) {
	s.observer.Notify(fact.MessageLoggedByProjection{
		HandlerName:  s.name,
		Handler:      s.handler,
		Envelope:     s.event,
		LogFormat:    f,
		LogArguments: v,
	})
}
