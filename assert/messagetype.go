package assert

import (
	"github.com/dogmatiq/configkit/message"
	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/testkit/compare"
	"github.com/dogmatiq/testkit/engine/envelope"
	"github.com/dogmatiq/testkit/engine/fact"
	"github.com/dogmatiq/testkit/internal/inflect"
	"github.com/dogmatiq/testkit/report"
)

// EventTypeRecorded returns an assertion that passes if a message with the same
// type as m is recorded as an event.
func EventTypeRecorded(m dogma.Message) Assertion {
	return &messageTypeAssertion{
		expected: message.TypeOf(m),
		role:     message.EventRole,
	}
}

// messageTypeAssertion asserts that a specific message is produced.
type messageTypeAssertion struct {
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

	// sim is a ranking of the similarity between the type of the expected
	// message, and the current best-match.
	sim compare.TypeSimilarity

	// tracker observers the handlers and messages that are involved in the
	// test.
	tracker tracker
}

// Banner returns a human-readable banner to display in the logs when this
// expectation is used.
//
// The banner text should be in uppercase, and complete the sentence "The
// application is expected ...". For example, "TO DO A THING".
func (a *messageTypeAssertion) Banner() string {
	return inflect.Sprintf(
		a.role,
		"TO <PRODUCE> ANY '%s' <MESSAGE>",
		a.expected,
	)
}

// Begin is called to prepare the assertion for a new test.
func (a *messageTypeAssertion) Begin(o ExpectOptionSet) {
	// reset everything
	*a = messageTypeAssertion{
		expected: a.expected,
		role:     a.role,
		tracker: tracker{
			role:               a.role,
			matchDispatchCycle: o.MatchMessagesInDispatchCycle,
		},
	}
}

// End is called once the test is complete.
func (a *messageTypeAssertion) End() {
}

// Ok returns true if the assertion passed.
func (a *messageTypeAssertion) Ok() bool {
	return a.ok
}

// BuildReport generates a report about the assertion.
//
// ok is true if the assertion is considered to have passed. This may not be the
// same value as returned from Ok() when this assertion is used as a
// sub-assertion inside a composite.
func (a *messageTypeAssertion) BuildReport(ok bool) *Report {
	rep := &Report{
		TreeOk: ok,
		Ok:     a.ok,
		Criteria: inflect.Sprintf(
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
	} else if a.best.Role == a.role {
		a.buildResultExpectedRole(rep)
	} else {
		a.buildResultUnexpectedRole(rep)
	}

	return rep
}

// Notify updates the assertion's state in response to a new fact.
func (a *messageTypeAssertion) Notify(f fact.Fact) {
	if a.ok {
		return
	}

	a.tracker.Notify(f)

	switch x := f.(type) {
	case fact.DispatchCycleBegun:
		if a.tracker.matchDispatchCycle {
			a.messageProduced(x.Envelope)
		}
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
func (a *messageTypeAssertion) messageProduced(env *envelope.Envelope) {
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
func (a *messageTypeAssertion) buildDiff(rep *Report) {
	report.WriteDiff(
		&rep.Section("Message Type Diff").Content,
		a.expected.String(),
		a.best.Type.ReflectType().String(),
	)
}

// buildResultExpectedRole builds the assertion result when there is a
// "best-match" message available but it is of an unexpected role.
func (a *messageTypeAssertion) buildResultExpectedRole(rep *Report) {
	s := rep.Section(suggestionsSection)

	if a.best.Origin == nil {
		rep.Explanation = inflect.Sprint(
			a.role,
			"a <message> of a similar type was <produced> via a <dispatcher>",
		)
	} else {
		rep.Explanation = inflect.Sprintf(
			a.role,
			"a <message> of a similar type was <produced> by the '%s' %s message handler",
			a.best.Origin.Handler.Identity().Name,
			a.best.Origin.HandlerType,
		)
	}

	// note this language here is deliberately vague, it doesn't imply whether
	// it currently is or isn't a pointer, just questions if it should be.
	s.AppendListItem("check the message type, should it be a pointer?")

	a.buildDiff(rep)
}

// buildResultUnexpectedRole builds the assertion result when there is a
// "best-match" message available but it is of an expected role.
func (a *messageTypeAssertion) buildResultUnexpectedRole(rep *Report) {
	s := rep.Section(suggestionsSection)

	if a.best.Origin == nil {
		s.AppendListItem(inflect.Sprint(
			a.best.Role,
			"verify that a <message> of this type was intended to be <produced> via a <dispatcher>",
		))
	} else {
		s.AppendListItem(inflect.Sprintf(
			a.best.Role,
			"verify that the '%s' %s message handler intended to <produce> a <message> of this type",
			a.best.Origin.Handler.Identity().Name,
			a.best.Origin.HandlerType,
		))
	}

	if a.role == message.CommandRole {
		s.AppendListItem("verify that CommandTypeExecuted() is the correct assertion, did you mean EventTypeRecorded()?")
	} else {
		s.AppendListItem("verify that EventTypeRecorded() is the correct assertion, did you mean CommandTypeExecuted()?")
	}

	if a.sim == compare.SameTypes {
		if a.best.Origin == nil {
			rep.Explanation = inflect.Sprint(
				a.best.Role,
				"a message of this type was <produced> as a <message> via a <dispatcher>",
			)
		} else {
			rep.Explanation = inflect.Sprintf(
				a.best.Role,
				"a message of this type was <produced> as a <message> by the '%s' %s message handler",
				a.best.Origin.Handler.Identity().Name,
				a.best.Origin.HandlerType,
			)
		}
	} else {
		if a.best.Origin == nil {
			rep.Explanation = inflect.Sprint(
				a.best.Role,
				"a message of a similar type was <produced> as a <message> via a <dispatcher>",
			)
		} else {
			rep.Explanation = inflect.Sprintf(
				a.best.Role,
				"a message of a similar type was <produced> as a <message> by the '%s' %s message handler",
				a.best.Origin.Handler.Identity().Name,
				a.best.Origin.HandlerType,
			)
		}

		// note this language here is deliberately vague, it doesn't imply
		// whether it currently is or isn't a pointer, just questions if it
		// should be.
		s.AppendListItem("check the message type, should it be a pointer?")

		a.buildDiff(rep)
	}
}
