package projection

import (
	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/dogmatest/engine/envelope"
	"github.com/dogmatiq/dogmatest/engine/fact"
)

// scope is an implementation of dogma.ProjectionEventScope.
type scope struct {
	name     string
	handler  dogma.ProjectionMessageHandler
	observer fact.Observer
	event    *envelope.Envelope
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
