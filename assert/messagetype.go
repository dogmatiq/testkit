package assert

import (
	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/dogmatest/compare"
	"github.com/dogmatiq/dogmatest/engine/envelope"
	"github.com/dogmatiq/dogmatest/engine/fact"
	"github.com/dogmatiq/dogmatest/render"
	"github.com/dogmatiq/enginekit/message"
)

// MessageTypeAssertion asserts that a specific message is produced.
type MessageTypeAssertion struct {
	// Expected is the type of the message that is expected to be produced.
	expected message.Type

	// Role is the expected role of the expected message.
	// It must be either CommandRole or EventRole.
	role message.Role

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

	// tracker observers the handlers and messages that are involved in the test.
	tracker tracker
}

// CommandTypeExecuted returns an assertion that passes if a message with the
// same type as m is executed as a command.
func CommandTypeExecuted(m dogma.Message) Assertion {
	return &MessageTypeAssertion{
		expected: message.TypeOf(m),
		role:     message.CommandRole,
	}
}

// EventTypeRecorded returns an assertion that passes if a message witn the same
// type as m is recorded as an event.
func EventTypeRecorded(m dogma.Message) Assertion {
	return &MessageTypeAssertion{
		expected: message.TypeOf(m),
		role:     message.EventRole,
	}
}

// Prepare is called to prepare the assertion for a new test.
//
// c is the comparator used to compare messages and other entities.
func (a *MessageTypeAssertion) Prepare(c compare.Comparator) {
	// reset everything
	*a = MessageTypeAssertion{
		expected: a.expected,
		role:     a.role,
		tracker:  tracker{role: a.role},
	}
}

// Ok returns true if the assertion passed.
func (a *MessageTypeAssertion) Ok() bool {
	return a.ok
}

// BuildReport generates a report about the assertion.
//
// ok is true if the assertion is considered to have passed. This may not be
// the same value as returned from Ok() when this assertion is used as
// sub-assertion inside a composite.
func (a *MessageTypeAssertion) BuildReport(ok bool, r render.Renderer) *Report {
	rep := &Report{
		TreeOk: ok,
		Ok:     a.ok,
		Criteria: inflect(
			a.role,
			"<produce> any '%s' <message>",
			a.expected,
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
func (a *MessageTypeAssertion) Notify(f fact.Fact) {
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
func (a *MessageTypeAssertion) messageProduced(env *envelope.Envelope) {
	sim := compare.FuzzyTypeComparison(
		a.expected.ReflectType(),
		env.Type.ReflectType(),
	)

	if sim > a.sim {
		a.best = env
		a.sim = sim
	}

	if sim == compare.SameTypes && a.role == env.Role {
		a.ok = true
	}
}

// buildDiff adds a "message type diff" section to the result.
func (a *MessageTypeAssertion) buildDiff(rep *Report) {
	render.WriteDiff(
		&rep.Section("Message Type Diff").Content,
		a.expected.String(),
		a.best.Type.ReflectType().String(),
	)
}

// buildResultExpectedRole builds the assertion result when there is a
// "best-match" message available and it is of the expected role.
func (a *MessageTypeAssertion) buildResultExpectedRole(r render.Renderer, rep *Report) {
	s := rep.Section(suggestionsSection)

	rep.Explanation = inflect(
		a.role,
		"a <message> of a similar type was <produced> by the '%s' %s message handler",
		a.best.Origin.HandlerName,
		a.best.Origin.HandlerType,
	)
	// note this language here is deliberately vague, it doesn't imply whether it
	// currently is or isn't a pointer, just questions if it should be.
	s.AppendListItem("check the message type, should it be a pointer?")

	a.buildDiff(rep)
}

// buildResultUnexpectedRole builds the assertion result when there is a
// "best-match" message available but it is of an expected role.
func (a *MessageTypeAssertion) buildResultUnexpectedRole(r render.Renderer, rep *Report) {
	s := rep.Section(suggestionsSection)

	s.AppendListItem(inflect(
		a.best.Role,
		"verify that the '%s' %s message handler intended to <produce> an <message> of this type",
		a.best.Origin.HandlerName,
		a.best.Origin.HandlerType,
	))

	if a.role == message.CommandRole {
		s.AppendListItem("verify that CommandTypeExecuted() is the correct assertion, did you mean EventTypeRecorded()?")
	} else {
		s.AppendListItem("verify that EventTypeRecorded() is the correct assertion, did you mean CommandTypeExecuted()?")
	}

	if a.sim == compare.SameTypes {
		rep.Explanation = inflect(
			a.best.Role,
			"a message of this type was <produced> as an <message> by the '%s' %s message handler",
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

		a.buildDiff(rep)
	}
}
