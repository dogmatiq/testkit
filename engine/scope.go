package engine

import (
	"github.com/dogmatiq/dogmatest/engine/envelope"
	"github.com/dogmatiq/dogmatest/engine/fact"
)

// scope is an implementation of controller.Scope that is used when handling a
// message outside the context of a test.
type scope struct {
	observers fact.ObserverSet
	env       *envelope.Envelope
}

func (s *scope) Envelope() *envelope.Envelope {
	return s.env
}

func (s *scope) RecordFacts(facts ...fact.Fact) {
	s.observers.Notify(facts...)
}
