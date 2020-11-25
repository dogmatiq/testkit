package testkit

import (
	"reflect"

	"github.com/dogmatiq/configkit/message"
	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/testkit/assert"
	"github.com/dogmatiq/testkit/compare"
	"github.com/dogmatiq/testkit/engine/envelope"
	"github.com/dogmatiq/testkit/engine/fact"
	"github.com/dogmatiq/testkit/internal/inflect"
	"github.com/dogmatiq/testkit/report"
)

// ToExecuteCommand returns an expectation that passes if a command is executed
// that is equal to m.
func ToExecuteCommand(m dogma.Message) Expectation {
	if m == nil {
		panic("ToExecuteCommand(): message must not be nil")
	}

	return &messageExpectation{
		expected: m,
		role:     message.CommandRole,
	}
}

// ToRecordEvent returns an expectation that passes if an event is recorded that
// is equal to m.
func ToRecordEvent(m dogma.Message) Expectation {
	if m == nil {
		panic("ToRecordEvent(): message must not be nil")
	}

	return &messageExpectation{
		expected: m,
		role:     message.EventRole,
	}
}

// messageExpectation verifies that a specific message is produced.
type messageExpectation struct {
	// Expected is the message that is expected to be produced.
	expected dogma.Message

	// Role is the expected role of the expected message.
	// It must be either CommandRole or EventRole.
	role message.Role

	// cmp is the comparator used to compare messages for equality.
	cmp compare.Comparator

	// ok is true once the expectation is deemed to have passed, after which no
	// further updates are performed.
	ok bool

	// best is an envelope containing the "best-match" message for an
	// expectation that has not yet passed. Note that this message may not have
	// the expected role.
	best *envelope.Envelope

	// sim is a ranking of the similarity between the type of the expected
	// message, and the current best-match.
	sim compare.TypeSimilarity

	// equal is true if the best-match message compared as equal to the expected
	// message. This can occur, and the expectation still fail, if the
	// best-match message has an unexpected role.
	equal bool

	// tracker observers the handlers and messages that are involved in the
	// test.
	tracker tracker
}

// Banner returns a human-readable banner to display in the logs when this
// expectation is used.
//
// The banner text should be in uppercase, and complete the sentence "The
// application is expected ...". For example, "TO DO A THING".
func (e *messageExpectation) Banner() string {
	return inflect.Sprintf(
		e.role,
		"TO <PRODUCE> A SPECIFIC '%s' <MESSAGE>",
		message.TypeOf(e.expected),
	)
}

// Begin is called to prepare the expectation for a new test.
func (e *messageExpectation) Begin(o ExpectOptionSet) {
	*e = messageExpectation{
		expected: e.expected,
		role:     e.role,
		cmp:      o.MessageComparator,
		tracker: tracker{
			role:               e.role,
			matchDispatchCycle: o.MatchMessagesInDispatchCycle,
		},
	}
}

// End is called once the test is complete.
func (e *messageExpectation) End() {
}

// Ok returns true if the expectation passed.
func (e *messageExpectation) Ok() bool {
	return e.ok
}

// BuildReport generates a report about the expectation.
//
// ok is true if the expectation is considered to have passed. This may not be
// the same value as returned from Ok() when this expectation is used as a child
// of a composite expectation.
func (e *messageExpectation) BuildReport(ok bool) *assert.Report {
	rep := &assert.Report{
		TreeOk: ok,
		Ok:     e.ok,
		Criteria: inflect.Sprintf(
			e.role,
			"<produce> a specific '%s' <message>",
			message.TypeOf(e.expected),
		),
	}

	if ok || e.ok {
		return rep
	}

	if e.best == nil {
		buildReportNoMatch(rep, &e.tracker)
	} else if e.best.Role == e.role {
		e.buildReportExpectedRole(rep)
	} else {
		e.buildReportUnexpectedRole(rep)
	}

	return rep
}

// Notify updates the expectation's state in response to a new fact.
func (e *messageExpectation) Notify(f fact.Fact) {
	if e.ok {
		return
	}

	e.tracker.Notify(f)

	switch x := f.(type) {
	case fact.DispatchCycleBegun:
		if e.tracker.matchDispatchCycle {
			e.messageProduced(x.Envelope)
		}
	case fact.EventRecordedByAggregate:
		e.messageProduced(x.EventEnvelope)
	case fact.EventRecordedByIntegration:
		e.messageProduced(x.EventEnvelope)
	case fact.CommandExecutedByProcess:
		e.messageProduced(x.CommandEnvelope)
	}
}

// messageProduced updates the expectation's state to reflect the fact that a
// message has been produced.
func (e *messageExpectation) messageProduced(env *envelope.Envelope) {
	if !e.cmp.MessageIsEqual(env.Message, e.expected) {
		e.updateBestMatch(env)
		return
	}

	e.best = env
	e.sim = compare.SameTypes
	e.equal = true

	if e.role == env.Role {
		e.ok = true
	}
}

// updateBestMatch replaces e.best with env if it is a better match.
func (e *messageExpectation) updateBestMatch(env *envelope.Envelope) {
	sim := compare.FuzzyTypeComparison(
		reflect.TypeOf(e.expected),
		reflect.TypeOf(env.Message),
	)

	if sim > e.sim {
		e.best = env
		e.sim = sim
	}
}

// buildReportExpectedRole builds a test report when there is a "best-match"
// message available and it is of the expected role.
func (e *messageExpectation) buildReportExpectedRole(rep *assert.Report) {
	s := rep.Section(suggestionsSection)

	if e.sim == compare.SameTypes {
		if e.best.Origin == nil {
			rep.Explanation = inflect.Sprint(
				e.role,
				"a similar <message> was <produced> via a <dispatcher>",
			)
		} else {
			rep.Explanation = inflect.Sprintf(
				e.role,
				"a similar <message> was <produced> by the '%s' %s message handler",
				e.best.Origin.Handler.Identity().Name,
				e.best.Origin.HandlerType,
			)
		}

		s.AppendListItem("check the content of the message")
	} else {
		if e.best.Origin == nil {
			rep.Explanation = inflect.Sprint(
				e.role,
				"a <message> of a similar type was <produced> via a <dispatcher>",
			)
		} else {
			rep.Explanation = inflect.Sprintf(
				e.role,
				"a <message> of a similar type was <produced> by the '%s' %s message handler",
				e.best.Origin.Handler.Identity().Name,
				e.best.Origin.HandlerType,
			)
		}

		// note this language here is deliberately vague, it doesn't imply
		// whether it currently is or isn't a pointer, just questions if it
		// should be.
		s.AppendListItem("check the message type, should it be a pointer?")
	}

	e.buildDiff(rep)
}

// buildDiff adds a "message diff" section to the result.
func (e *messageExpectation) buildDiff(rep *assert.Report) {
	report.WriteDiff(
		&rep.Section("Message Diff").Content,
		report.RenderMessage(e.expected),
		report.RenderMessage(e.best.Message),
	)
}

// buildReportExpectedRole builds a test report when there is a "best-match"
// message available but it is of an unexpected role.
func (e *messageExpectation) buildReportUnexpectedRole(rep *assert.Report) {
	s := rep.Section(suggestionsSection)

	if e.best.Origin == nil {
		s.AppendListItem(inflect.Sprint(
			e.best.Role,
			"verify that a <message> of this type was intended to be <produced> via a <dispatcher>",
		))
	} else {
		s.AppendListItem(inflect.Sprintf(
			e.best.Role,
			"verify that the '%s' %s message handler intended to <produce> a <message> of this type",
			e.best.Origin.Handler.Identity().Name,
			e.best.Origin.HandlerType,
		))
	}

	if e.role == message.CommandRole {
		s.AppendListItem("verify that ToExecuteCommand() is the correct expectation, did you mean ToRecordEvent()?")
	} else {
		s.AppendListItem("verify that ToRecordEvent() is the correct expectation, did you mean ToExecuteCommand()?")
	}

	// the "best-match" is equal to the expected message. this means that only
	// the roles were mismatched.
	if e.equal {
		if e.best.Origin == nil {
			rep.Explanation = inflect.Sprint(
				e.best.Role,
				"the expected message was <produced> as a <message> via a <dispatcher>",
			)
		} else {
			rep.Explanation = inflect.Sprintf(
				e.best.Role,
				"the expected message was <produced> as a <message> by the '%s' %s message handler",
				e.best.Origin.Handler.Identity().Name,
				e.best.Origin.HandlerType,
			)
		}

		return // skip diff rendering
	}

	if e.sim == compare.SameTypes {
		if e.best.Origin == nil {
			rep.Explanation = inflect.Sprint(
				e.best.Role,
				"a similar message was <produced> as a <message> via a <dispatcher>",
			)
		} else {
			rep.Explanation = inflect.Sprintf(
				e.best.Role,
				"a similar message was <produced> as a <message> by the '%s' %s message handler",
				e.best.Origin.Handler.Identity().Name,
				e.best.Origin.HandlerType,
			)
		}
	} else {
		if e.best.Origin == nil {
			rep.Explanation = inflect.Sprint(
				e.best.Role,
				"a message of a similar type was <produced> as a <message> via a <dispatcher>",
			)
		} else {
			rep.Explanation = inflect.Sprintf(
				e.best.Role,
				"a message of a similar type was <produced> as a <message> by the '%s' %s message handler",
				e.best.Origin.Handler.Identity().Name,
				e.best.Origin.HandlerType,
			)
		}

		// note this language here is deliberately vague, it doesn't imply
		// whether it currently is or isn't a pointer, just questions if it
		// should be.
		s.AppendListItem("check the message type, should it be a pointer?")
	}

	e.buildDiff(rep)
}
