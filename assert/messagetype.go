package assert

import (
	"reflect"

	"github.com/dogmatiq/dogmatest/compare"
	"github.com/dogmatiq/dogmatest/engine/envelope"
	"github.com/dogmatiq/dogmatest/engine/fact"
	"github.com/dogmatiq/dogmatest/render"
	"github.com/dogmatiq/enginekit/handler"
	"github.com/dogmatiq/enginekit/message"
)

// MessageTypeAssertion asserts that a specific message is produced.
type MessageTypeAssertion struct {
	// Expected is the type of the message that is expected to be produced.
	Expected reflect.Type

	// Role is the expected role of the expected message.
	// It must be either CommandRole or EventRole.
	Role message.Role

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

	// total is the total number of messages that were produced.
	total int

	// produces is the number of messages of the expected role that were produced.
	produced int

	// engaged is the set of handlers that *could* have produced the
	// expected message.
	engaged map[string]handler.Type

	// enabled is the set of handler types that are enabled during the
	// test.
	enabled map[handler.Type]bool
}

// Begin is called before the test is executed.
//
// c is the comparator used to compare messages and other entities.
func (a *MessageTypeAssertion) Begin(c compare.Comparator) {
	// reset everything
	*a = MessageTypeAssertion{
		Expected: a.Expected,
		Role:     a.Role,
		engaged:  map[string]handler.Type{},
	}
}

// End is called after the test is executed.
//
// It returns the result of the assertion.
func (a *MessageTypeAssertion) End(r render.Renderer) *Result {
	res := &Result{
		Ok: a.ok,
		Criteria: inflect(
			a.Role,
			"<produce> any '%s' <message>",
			message.TypeOf(a.Expected),
		),
	}

	if !a.ok {
		if a.best == nil {
			buildResultNoMatch(
				res,
				a.Role,
				a.enabled,
				a.engaged,
				a.total,
				a.produced,
			)
		} else if a.best.Role == message.EventRole {
			a.buildResultExpectedRole(r, res)
		} else {
			a.buildResultUnexpectedRole(r, res)
		}
	}

	return res
}

// Notify updates the assertion's state in response to a new fact.
func (a *MessageTypeAssertion) Notify(f fact.Fact) {
	if a.ok {
		return
	}

	switch x := f.(type) {
	case fact.MessageDispatchBegun:
		a.enabled = x.EnabledHandlers

	case fact.EngineTickBegun:
		a.enabled = x.EnabledHandlers

	case fact.MessageHandlingBegun:
		a.messageHandlingBegun(x)

	case fact.EventRecordedByAggregate:
		a.messageProduced(x.EventEnvelope)

	case fact.EventRecordedByIntegration:
		a.messageProduced(x.EventEnvelope)

	case fact.CommandExecutedByProcess:
		a.messageProduced(x.CommandEnvelope)
	}
}

// messageHandlingBegun updates the assertion's state to reflect f.
func (a *MessageTypeAssertion) messageHandlingBegun(f fact.MessageHandlingBegun) {
	if f.HandlerType.IsProducerOf(a.Role) {
		a.engaged[f.HandlerName] = f.HandlerType
	}
}

// messageProduced updates the assertion's state to reflect the fact that a
// message has been produced.
func (a *MessageTypeAssertion) messageProduced(env *envelope.Envelope) {
	a.total++
	if env.Role == a.Role {
		a.produced++
	}

	sim := compare.FuzzyTypeComparison(
		reflect.TypeOf(a.Expected),
		reflect.TypeOf(env.Message),
	)

	if sim > a.sim {
		a.best = env
		a.sim = sim
	}

	if sim == compare.SameTypes && a.Role == env.Role {
		a.ok = true
	}
}

// buildDiff adds a "message type diff" section to the result.
func (a *MessageTypeAssertion) buildDiff(res *Result) {
	render.WriteDiff(
		&res.Section("Message Type Diff").Content,
		a.Expected.String(),
		a.best.Type.ReflectType().String(),
	)
}

// buildResultExpectedRole builds the assertion result when there is a
// "best-match" message available and it is of the expected role.
func (a *MessageTypeAssertion) buildResultExpectedRole(r render.Renderer, res *Result) {
	s := res.Section(suggestionsSection)

	res.Explanation = inflect(
		"a <message> of a similar type was <produced> by the '%s' %s message handler",
		a.best.Origin.HandlerName,
		a.best.Origin.HandlerType,
	)
	// note this language here is deliberately vague, it doesn't imply whether it
	// currently is or isn't a pointer, just questions if it should be.
	s.AppendListItem("check the message type, should it be a pointer?")

	a.buildDiff(res)
}

// buildResultUnexpectedRole builds the assertion result when there is a
// "best-match" message available but it is of an expected role.
func (a *MessageTypeAssertion) buildResultUnexpectedRole(r render.Renderer, res *Result) {
	s := res.Section(suggestionsSection)

	s.AppendListItem(
		"verify that the '%s' %s message handler intended to <other-produce> an <other-message> of this type",
		a.best.Origin.HandlerName,
		a.best.Origin.HandlerType,
	)

	if a.Role == message.CommandRole {
		s.AppendListItem("verify that CommandTypeExecuted() is the correct assertion, did you mean EventTypeRecorded()?")
	} else {
		s.AppendListItem("verify that EventTypeRecorded() is the correct assertion, did you mean CommandTypeExecuted()?")
	}

	if a.sim == compare.SameTypes {
		res.Explanation = inflect(
			"a message of this type was <other-produced> as an <other-message> by the '%s' %s message handler",
			a.best.Origin.HandlerName,
			a.best.Origin.HandlerType,
		)
	} else {
		res.Explanation = inflect(
			"a message of a similar type was <other-produced> as an <other-message> by the '%s' %s message handler",
			a.best.Origin.HandlerName,
			a.best.Origin.HandlerType,
		)
		// note this language here is deliberately vague, it doesn't imply whether it
		// currently is or isn't a pointer, just questions if it should be.
		s.AppendListItem("check the message type, should it be a pointer?")

		a.buildDiff(res)
	}
}
