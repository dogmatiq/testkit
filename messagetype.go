package testkit

import (
	"github.com/dogmatiq/configkit/message"
	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/testkit/assert"
	"github.com/dogmatiq/testkit/compare"
	"github.com/dogmatiq/testkit/engine/envelope"
	"github.com/dogmatiq/testkit/engine/fact"
	"github.com/dogmatiq/testkit/internal/inflect"
	"github.com/dogmatiq/testkit/report"
)

// ToExecuteCommandOfType returns an expectation that passes a command of the
// same type as m is executed.
func ToExecuteCommandOfType(m dogma.Message) Expectation {
	if m == nil {
		panic("ToExecuteCommandOfType(): message must not be nil")
	}

	return &produceMessageOfType{
		expected: message.TypeOf(m),
		role:     message.CommandRole,
	}
}

// ToRecordEventOfType returns an expectation that passes if an event of the
// same type as m is recorded.
func ToRecordEventOfType(m dogma.Message) Expectation {
	if m == nil {
		panic("ToRecordEventOfType(): message must not be nil")
	}

	return &produceMessageOfType{
		expected: message.TypeOf(m),
		role:     message.EventRole,
	}
}

// produceMessageOfType verifies that a message of a specific type is produced,
// either as a command or an event.
type produceMessageOfType struct {
	// Expected is the type of the message that is expected to be produced.
	expected message.Type

	// Role is the expected role that the message must have.
	// It must be either CommandRole or EventRole.
	role message.Role

	// ok is true once the expectation has passed, after which no further
	// updates are performed.
	ok bool

	// best is an envelope containing the "best-match" message found so far.
	// Note that this message may not have the expected role.
	best *envelope.Envelope

	// sim is a ranking of the similarity between the type of the expected
	// message, and the current best-match.
	sim compare.TypeSimilarity

	// tracker observes the handlers and messages that are involved in the test.
	tracker tracker
}

// Banner returns a human-readable banner to display in the logs when this
// expectation is used.
//
// The banner text should be in uppercase, and complete the sentence "The
// application is expected ...". For example, "TO DO A THING".
func (e *produceMessageOfType) Banner() string {
	return inflect.Sprintf(
		e.role,
		"TO <PRODUCE> A <MESSAGE> OF TYPE %s",
		e.expected,
	)
}

// Begin is called to prepare the expectation for a new test.
func (e *produceMessageOfType) Begin(o ExpectOptionSet) {
	*e = produceMessageOfType{
		expected: e.expected,
		role:     e.role,
		tracker: tracker{
			role:               e.role,
			matchDispatchCycle: o.MatchMessagesInDispatchCycle,
		},
	}
}

// End is called once the test is complete.
func (e *produceMessageOfType) End() {
}

// Ok returns true if the expectation passed.
func (e *produceMessageOfType) Ok() bool {
	return e.ok
}

// BuildReport generates a report about the expectation.
//
// ok is true if the expectation is considered to have passed. This may not be
// the same value as returned from Ok() when this expectation is used as a child
// of a composite expectation.
func (e *produceMessageOfType) BuildReport(ok bool) *assert.Report {
	rep := &assert.Report{
		TreeOk: ok,
		Ok:     e.ok,
		Criteria: inflect.Sprintf(
			e.role,
			"<produce> any '%s' <message>",
			e.expected,
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
func (e *produceMessageOfType) Notify(f fact.Fact) {
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
func (e *produceMessageOfType) messageProduced(env *envelope.Envelope) {
	sim := compare.FuzzyTypeComparison(
		e.expected.ReflectType(),
		env.Type.ReflectType(),
	)

	if sim > e.sim {
		e.best = env
		e.sim = sim
	}

	if sim == compare.SameTypes && e.role == env.Role {
		e.ok = true
	}
}

// buildDiff adds a "message type diff" section to the result.
func (e *produceMessageOfType) buildDiff(rep *assert.Report) {
	report.WriteDiff(
		&rep.Section("Message Type Diff").Content,
		e.expected.String(),
		e.best.Type.ReflectType().String(),
	)
}

// buildReportExpectedRole builds a test report when there is a "best-match"
// message available but it is of an unexpected role.
func (e *produceMessageOfType) buildReportExpectedRole(rep *assert.Report) {
	s := rep.Section(suggestionsSection)

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
			e.best.Origin.Handler.HandlerType(),
		)
	}

	// note this language here is deliberately vague, it doesn't imply whether
	// it currently is or isn't a pointer, just questions if it should be.
	s.AppendListItem("check the message type, should it be a pointer?")

	e.buildDiff(rep)
}

// buildReportUnexpectedRole builds a test report when there is a "best-match"
// message available but it does not have the expected role.
func (e *produceMessageOfType) buildReportUnexpectedRole(rep *assert.Report) {
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
			e.best.Origin.Handler.HandlerType(),
		))
	}

	if e.role == message.CommandRole {
		s.AppendListItem("verify that ToExecuteCommandOfType() is the correct expectation, did you mean ToRecordEventOfType()?")
	} else {
		s.AppendListItem("verify that ToRecordEventOfType() is the correct expectation, did you mean ToExecuteCommandOfType()?")
	}

	if e.sim == compare.SameTypes {
		if e.best.Origin == nil {
			rep.Explanation = inflect.Sprint(
				e.best.Role,
				"a message of this type was <produced> as a <message> via a <dispatcher>",
			)
		} else {
			rep.Explanation = inflect.Sprintf(
				e.best.Role,
				"a message of this type was <produced> as a <message> by the '%s' %s message handler",
				e.best.Origin.Handler.Identity().Name,
				e.best.Origin.Handler.HandlerType(),
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
				e.best.Origin.Handler.HandlerType(),
			)
		}

		// note this language here is deliberately vague, it doesn't imply
		// whether it currently is or isn't a pointer, just questions if it
		// should be.
		s.AppendListItem("check the message type, should it be a pointer?")

		e.buildDiff(rep)
	}
}
