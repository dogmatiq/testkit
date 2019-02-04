package assert

import (
	"fmt"
	"reflect"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/enginekit/message"
	"github.com/dogmatiq/testkit/compare"
	"github.com/dogmatiq/testkit/engine/envelope"
	"github.com/dogmatiq/testkit/engine/fact"
	"github.com/dogmatiq/testkit/render"
)

// MessageAssertion asserts that a specific message is produced.
type MessageAssertion struct {
	// Expected is the message that is expected to be produced.
	expected dogma.Message

	// Role is the expected role of the expected message.
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

	// tracker observers the handlers and messages that are involved in the test.
	tracker tracker
}

// CommandExecuted returns an assertion that passes if m is executed as a command.
func CommandExecuted(m dogma.Message) Assertion {
	return &MessageAssertion{
		expected: m,
		role:     message.CommandRole,
	}
}

// EventRecorded returns an assertion that passes if m is recorded as an event.
func EventRecorded(m dogma.Message) Assertion {
	return &MessageAssertion{
		expected: m,
		role:     message.EventRole,
	}
}

// Prepare is called to prepare the assertion for a new test.
//
// c is the comparator used to compare messages and other entities.
func (a *MessageAssertion) Prepare(c compare.Comparator) {
	// reset everything
	*a = MessageAssertion{
		expected: a.expected,
		role:     a.role,
		cmp:      c,
		tracker:  tracker{role: a.role},
	}
}

// Ok returns true if the assertion passed.
func (a *MessageAssertion) Ok() bool {
	return a.ok
}

// BuildReport generates a report about the assertion.
//
// ok is true if the assertion is considered to have passed. This may not be
// the same value as returned from Ok() when this assertion is used as
// sub-assertion inside a composite.
func (a *MessageAssertion) BuildReport(ok bool, r render.Renderer) *Report {
	rep := &Report{
		TreeOk: ok,
		Ok:     a.ok,
		Criteria: inflect(
			a.role,
			"<produce> a specific '%s' <message>",
			message.TypeOf(a.expected),
		),
	}

	if ok || a.ok {
		return rep
	}

	if a.best == nil {
		buildResultNoMatch(rep, &a.tracker)
	} else if a.best.Role == message.EventRole {
		a.buildResultExpectedRole(r, rep)
	} else {
		a.buildResultUnexpectedRole(r, rep)
	}

	return rep
}

// Notify updates the assertion's state in response to a new fact.
func (a *MessageAssertion) Notify(f fact.Fact) {
	if a.ok {
		return
	}

	a.tracker.Notify(f)

	switch x := f.(type) {
	case fact.EventRecordedByAggregate:
		a.messageProduced(x.EventEnvelope)
	case fact.EventRecordedByIntegration:
		a.messageProduced(x.EventEnvelope)
	case fact.CommandExecutedByProcess:
		a.messageProduced(x.CommandEnvelope)
	}
}

// messageProduced updates the assertion's state to reflect the fact that a
// message has been produced.
func (a *MessageAssertion) messageProduced(env *envelope.Envelope) {
	if !a.cmp.MessageIsEqual(env.Message, a.expected) {
		a.updateBestMatch(env)
		return
	}

	a.best = env
	a.sim = compare.SameTypes
	a.equal = true

	if a.role == env.Role {
		a.ok = true
	}
}

// updateBestMatch replaces a.best with env if it is a better match.
func (a *MessageAssertion) updateBestMatch(env *envelope.Envelope) {
	sim := compare.FuzzyTypeComparison(
		reflect.TypeOf(a.expected),
		reflect.TypeOf(env.Message),
	)

	if sim > a.sim {
		a.best = env
		a.sim = sim
	}
}

// buildResultExpectedRole builds the assertion result when there is a
// "best-match" message available and it is of the expected role.
func (a *MessageAssertion) buildResultExpectedRole(r render.Renderer, rep *Report) {
	s := rep.Section(suggestionsSection)

	if a.sim == compare.SameTypes {
		rep.Explanation = inflect(
			a.role,
			"a similar <message> was <produced> by the '%s' %s message handler",
			a.best.Origin.HandlerName,
			a.best.Origin.HandlerType,
		)
		s.AppendListItem("check the content of the message")
	} else {
		rep.Explanation = inflect(
			a.role,
			"a <message> of a similar type was <produced> by the '%s' %s message handler",
			a.best.Origin.HandlerName,
			a.best.Origin.HandlerType,
		)
		// note this language here is deliberately vague, it doesn't imply whether it
		// currently is or isn't a pointer, just questions if it should be.
		s.AppendListItem("check the message type, should it be a pointer?")
	}

	render.WriteDiff(
		&rep.Section(messageDiffSection).Content,
		render.Message(r, a.expected),
		render.Message(r, a.best.Message),
	)
}

// buildDiff adds a "message diff" section to the result.
func (a *MessageAssertion) buildDiff(r render.Renderer, rep *Report) {
	render.WriteDiff(
		&rep.Section("Message Diff").Content,
		render.Message(r, a.expected),
		render.Message(r, a.best.Message),
	)
}

// buildResultUnexpectedRole builds the assertion result when there is a
// "best-match" message available but it is of an expected role.
func (a *MessageAssertion) buildResultUnexpectedRole(r render.Renderer, rep *Report) {
	s := rep.Section(suggestionsSection)

	s.AppendListItem(inflect(
		a.best.Role,
		"verify that the '%s' %s message handler intended to <produce> an <message> of this type",
		a.best.Origin.HandlerName,
		a.best.Origin.HandlerType,
	))

	if a.role == message.CommandRole {
		s.AppendListItem("verify that CommandExecuted() is the correct assertion, did you mean EventRecorded()?")
	} else {
		s.AppendListItem("verify that EventRecorded() is the correct assertion, did you mean CommandExecuted()?")
	}

	// the "best-match" is equal to the expected message. this means that only the
	// roles were mismatched.
	if a.equal {
		rep.Explanation = inflect(
			a.best.Role,
			"the expected message was <produced> as a <message> by the '%s' %s message handler",
			a.best.Origin.HandlerName,
			a.best.Origin.HandlerType,
		)

		return // skip diff rendering
	}

	if a.sim == compare.SameTypes {
		rep.Explanation = inflect(
			a.best.Role,
			"a similar message was <produced> as an <message> by the '%s' %s message handler",
			a.best.Origin.HandlerName,
			a.best.Origin.HandlerType,
		)
	} else {
		rep.Explanation = inflect(
			a.best.Role,
			"a message of a similar type was <produced> as an <message> by the '%s' %s message handler",
			a.best.Origin.HandlerName,
			a.best.Origin.HandlerType,
		)
		// note this language here is deliberately vague, it doesn't imply whether it
		// currently is or isn't a pointer, just questions if it should be.
		s.AppendListItem("check the message type, should it be a pointer?")
	}

	a.buildDiff(r, rep)
}

// buildResultNoMatch is a helper used by MessageAssertion and
// MessageTypeAssertion when there is no "best-match".
func buildResultNoMatch(rep *Report, t *tracker) {
	s := rep.Section(suggestionsSection)

	allDisabled := true
	for ht, e := range t.enabled {
		if ht.IsProducerOf(t.role) {
			if e {
				allDisabled = false
			} else {
				s.AppendListItem(
					fmt.Sprintf("enable %s handlers using the EnableHandlerType() option", ht),
				)
			}
		}
	}

	if allDisabled {
		rep.Explanation = "no relevant handler types were enabled"
		return
	}

	if len(t.engaged) == 0 {
		rep.Explanation = "no relevant handlers (aggregates or integrations) were engaged"
		s.AppendListItem("check the application's routing configuration")
		return
	}

	if t.total == 0 {
		rep.Explanation = "no messages were produced at all"
	} else if t.produced == 0 {
		rep.Explanation = inflect(t.role, "no <messages> were <produced> at all")
	} else {
		rep.Explanation = inflect(t.role, "none of the engaged handlers <produced> the expected <message>")
	}

	for n, t := range t.engaged {
		s.AppendListItem("verify the logic within the '%s' %s message handler", n, t)
	}
}
