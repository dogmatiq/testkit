package assert

import (
	"reflect"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/dogmatest/compare"
	"github.com/dogmatiq/dogmatest/engine/envelope"
	"github.com/dogmatiq/dogmatest/engine/fact"
	"github.com/dogmatiq/enginekit/handler"
	"github.com/dogmatiq/enginekit/message"
)

type messageAssertionBehavior struct {
	// expected is the message that is expected to be produced.
	expected dogma.Message

	// role is the expected role of the expected message.
	// It must be either CommandRole or EventRole.
	role message.Role

	// cmp is the comparator used to compare messages for equality.
	cmp compare.Comparator

	// ok is true once the assertion is deemed to have passed, after which no
	// further updates are performed.
	ok bool

	// best is an envelope containing the "best-match" message for an assertion
	// that has not yet passed. Note that this message may not have the expected
	// role.
	best *envelope.Envelope

	// sim is a ranking of the similarity between the type of the expected message,
	// and the current best-match.
	sim compare.TypeSimilarity

	// equal is true if the best-match message compared as equal to the expected
	// message. This can occur, and the assertion still fail, if the best-match
	// message has an unexpected role.
	equal bool

	// events is the total number of events that were recorded.
	events int

	// commands is the total number of commands that were executed.
	commands int

	// engagedHandlers is the set of handlers that *could* have produced the
	// expected message.
	engagedHandlers map[string]handler.Type

	// enabledHandlers is the set of handler types that are enabled during the
	// test.
	enabledHandlers map[handler.Type]bool
}

// Notify updates the assertion's state in response to a new fact.
func (a *messageAssertionBehavior) Notify(f fact.Fact) {
	if a.ok {
		return
	}

	switch x := f.(type) {
	case fact.MessageDispatchBegun:
		a.enabledHandlers = x.EnabledHandlers

	case fact.EngineTickBegun:
		a.enabledHandlers = x.EnabledHandlers

	case fact.MessageHandlingBegun:
		a.messageHandlingBegun(x)

	case fact.EventRecordedByAggregate:
		a.events++
		a.messageProduced(x.EventEnvelope)

	case fact.EventRecordedByIntegration:
		a.events++
		a.messageProduced(x.EventEnvelope)

	case fact.CommandExecutedByProcess:
		a.commands++
		a.messageProduced(x.CommandEnvelope)
	}
}

// messageHandlingBegun updates the assertion's state to reflect f.
func (a *messageAssertionBehavior) messageHandlingBegun(f fact.MessageHandlingBegun) {
	if !f.HandlerType.IsProducerOf(a.role) {
		return
	}

	if a.engagedHandlers == nil {
		a.engagedHandlers = map[string]handler.Type{}
	}

	a.engagedHandlers[f.HandlerName] = f.HandlerType
}

// messageProduced updates the assertion's state to reflect the fact that a
// message has been produced.
func (a *messageAssertionBehavior) messageProduced(env *envelope.Envelope) {
	if !a.cmp.MessageIsEqual(env.Message, a.expected) {
		a.updateBestMatch(env)
		return
	}

	a.best = env
	a.sim = compare.SameTypes
	a.equal = true

	if a.role == message.EventRole {
		a.ok = true
	}
}

// updateBestMatch replaces a.best with env if it is a better match.
func (a *messageAssertionBehavior) updateBestMatch(env *envelope.Envelope) {
	sim := compare.FuzzyTypeComparison(
		reflect.TypeOf(a.expected),
		reflect.TypeOf(env.Message),
	)

	if sim > a.sim {
		a.best = env
		a.sim = sim
	}
}
