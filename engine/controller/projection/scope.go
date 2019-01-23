package projection

import (
	"github.com/dogmatiq/dogmatest/engine/envelope"
	"github.com/dogmatiq/dogmatest/engine/fact"
)

// scope is an implementation of dogma.ProjectionEventScope.
type scope struct {
	name     string
	observer fact.Observer
	event    *envelope.Envelope
}

func (s *scope) Log(f string, v ...interface{}) {
	s.observer.Notify(fact.MessageLoggedByProjection{
		HandlerName:  s.name,
		Envelope:     s.event,
		LogFormat:    f,
		LogArguments: v,
	})
}
