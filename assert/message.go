package assert

import (
	"fmt"
	"reflect"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/dogmatest/compare"
	"github.com/dogmatiq/dogmatest/engine/envelope"
	"github.com/dogmatiq/dogmatest/engine/fact"
	"github.com/dogmatiq/dogmatest/render"
	"github.com/dogmatiq/enginekit/handler"
	"github.com/dogmatiq/enginekit/message"
)

// MessageAssertion asserts that a specific message is produced.
type MessageAssertion struct {
	// Expected is the message that is expected to be produced.
	Expected dogma.Message

	// Role is the expected role of the expected message.
	// It must be either CommandRole or EventRole.
	Role message.Role

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
func (a *MessageAssertion) Begin(c compare.Comparator) {
	// reset everything
	*a = MessageAssertion{
		Expected: a.Expected,
		Role:     a.Role,
		cmp:      c,
		engaged:  map[string]handler.Type{},
	}
}

// End is called after the test is executed.
//
// It returns the result of the assertion.
func (a *MessageAssertion) End(r render.Renderer) *Result {
	res := &Result{
		Ok: a.ok,
		Criteria: inflect(
			a.Role,
			"<produce> a specific '%s' <message>",
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
func (a *MessageAssertion) Notify(f fact.Fact) {
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
func (a *MessageAssertion) messageHandlingBegun(f fact.MessageHandlingBegun) {
	if f.HandlerType.IsProducerOf(a.Role) {
		a.engaged[f.HandlerName] = f.HandlerType
	}
}

// messageProduced updates the assertion's state to reflect the fact that a
// message has been produced.
func (a *MessageAssertion) messageProduced(env *envelope.Envelope) {
	a.total++
	if env.Role == a.Role {
		a.produced++
	}

	if !a.cmp.MessageIsEqual(env.Message, a.Expected) {
		a.updateBestMatch(env)
		return
	}

	a.best = env
	a.sim = compare.SameTypes
	a.equal = true

	if a.Role == env.Role {
		a.ok = true
	}
}

// updateBestMatch replaces a.best with env if it is a better match.
func (a *MessageAssertion) updateBestMatch(env *envelope.Envelope) {
	sim := compare.FuzzyTypeComparison(
		reflect.TypeOf(a.Expected),
		reflect.TypeOf(env.Message),
	)

	if sim > a.sim {
		a.best = env
		a.sim = sim
	}
}

// buildResultExpectedRole builds the assertion result when there is a
// "best-match" message available and it is of the expected role.
func (a *MessageAssertion) buildResultExpectedRole(r render.Renderer, res *Result) {
	s := res.Section(suggestionsSection)

	if a.sim == compare.SameTypes {
		res.Explanation = inflect(
			"a similar <message> was <produced> by the '%s' %s message handler",
			a.best.Origin.HandlerName,
			a.best.Origin.HandlerType,
		)
		s.AppendListItem("check the content of the message")
	} else {
		res.Explanation = inflect(
			"a <message> of a similar type was <produced> by the '%s' %s message handler",
			a.best.Origin.HandlerName,
			a.best.Origin.HandlerType,
		)
		// note this language here is deliberately vague, it doesn't imply whether it
		// currently is or isn't a pointer, just questions if it should be.
		s.AppendListItem("check the message type, should it be a pointer?")
	}

	render.WriteDiff(
		&res.Section(messageDiffSection).Content,
		render.Message(r, a.Expected),
		render.Message(r, a.best.Message),
	)
}

// buildDiff adds a "message diff" section to the result.
func (a *MessageAssertion) buildDiff(r render.Renderer, res *Result) {
	render.WriteDiff(
		&res.Section("Message Diff").Content,
		render.Message(r, a.Expected),
		render.Message(r, a.best.Message),
	)
}

// buildResultUnexpectedRole builds the assertion result when there is a
// "best-match" message available but it is of an expected role.
func (a *MessageAssertion) buildResultUnexpectedRole(r render.Renderer, res *Result) {
	s := res.Section(suggestionsSection)

	s.AppendListItem(
		"verify that the '%s' %s message handler intended to <other-produce> an <other-message> of this type",
		a.best.Origin.HandlerName,
		a.best.Origin.HandlerType,
	)

	if a.Role == message.CommandRole {
		s.AppendListItem("verify that CommandExecuted() is the correct assertion, did you mean EventRecorded()?")
	} else {
		s.AppendListItem("verify that EventRecorded() is the correct assertion, did you mean CommandExecuted()?")
	}

	// the "best-match" is equal to the expected message. this means that only the
	// roles were mismatched.
	if a.equal {
		res.Explanation = inflect(
			"the expected message was <other-produced> as an <other-message> by the '%s' %s message handler",
			a.best.Origin.HandlerName,
			a.best.Origin.HandlerType,
		)

		return // skip diff rendering
	}

	if a.sim == compare.SameTypes {
		res.Explanation = inflect(
			"a similar message was <other-produced> as an <other-message> by the '%s' %s message handler",
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
	}

	a.buildDiff(r, res)
}

// buildResultNoMatch is a helper used by MessageAssertion and
// MessageTypeAssertion when there is no "best-match".
func buildResultNoMatch(
	res *Result,
	r message.Role,
	enabled map[handler.Type]bool,
	engaged map[string]handler.Type,
	total, produced int,
) {
	s := res.Section(suggestionsSection)

	allDisabled := true
	for t, e := range enabled {
		if t.IsProducerOf(r) {
			if e {
				allDisabled = false
			} else {
				s.AppendListItem(
					fmt.Sprintf("enable %s handlers using the EnableHandlerType() option", t),
				)
			}
		}
	}

	if allDisabled {
		res.Explanation = "no relevant handler types were enabled"
		return
	}

	if len(engaged) == 0 {
		res.Explanation = "no relevant handlers (aggregates or integrations) were engaged"
		s.AppendListItem("check the application's routing configuration")
		return
	}

	if total == 0 {
		res.Explanation = "no messages were produced at all"
	} else if produced == 0 {
		res.Explanation = inflect(r, "no <messages> were <produced> at all")
	} else {
		res.Explanation = inflect(r, "none of the engaged handlers <produced> the expected <message>")
	}

	for n, t := range engaged {
		s.AppendListItem("verify the logic within the '%s' %s message handler", n, t)
	}
}
