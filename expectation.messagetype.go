package testkit

import (
	"github.com/dogmatiq/configkit/message"
	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/testkit/envelope"
	"github.com/dogmatiq/testkit/fact"
	"github.com/dogmatiq/testkit/internal/inflect"
	"github.com/dogmatiq/testkit/internal/typecmp"
	"github.com/dogmatiq/testkit/report"
)

// ToExecuteCommandOfType returns an expectation that passes if a command of the
// same type as m is executed.
func ToExecuteCommandOfType(m dogma.Message) Expectation {
	if m == nil {
		panic("ToExecuteCommandOfType(<nil>): message must not be nil")
	}

	return &messageTypeExpectation{
		expectedType: message.TypeOf(m),
		expectedRole: message.CommandRole,
	}
}

// ToRecordEventOfType returns an expectation that passes if an event of the
// same type as m is recorded.
func ToRecordEventOfType(m dogma.Message) Expectation {
	if m == nil {
		panic("ToRecordEventOfType(<nil>): message must not be nil")
	}

	return &messageTypeExpectation{
		expectedType: message.TypeOf(m),
		expectedRole: message.EventRole,
	}
}

// messageTypeExpectation is an Expectation that checks that a message of a
// specific type is produced.
//
// It is the implementation used by ToExecuteCommandOfType() and
// ToRecordEventOfType().
type messageTypeExpectation struct {
	expectedType message.Type
	expectedRole message.Role
}

func (e *messageTypeExpectation) Banner() string {
	return inflect.Sprintf(
		e.expectedRole,
		"TO <PRODUCE> A <MESSAGE> OF TYPE %s",
		e.expectedType,
	)
}

func (e *messageTypeExpectation) Predicate(o PredicateOptions) (Predicate, error) {
	return &messageTypePredicate{
		expectedType:      e.expectedType,
		expectedRole:      e.expectedRole,
		bestMatchDistance: typecmp.Unrelated,
		tracker: tracker{
			role:               e.expectedRole,
			matchDispatchCycle: o.MatchDispatchCycleStartedFacts,
		},
	}, nil
}

// messageTypePredicate is the Predicate implementation for
// messageTypeExpectation.
type messageTypePredicate struct {
	expectedType      message.Type
	expectedRole      message.Role
	ok                bool
	bestMatch         *envelope.Envelope
	bestMatchDistance typecmp.Distance
	tracker           tracker
}

func (p *messageTypePredicate) Notify(f fact.Fact) {
	if p.ok {
		return
	}

	p.tracker.Notify(f)

	switch x := f.(type) {
	case fact.DispatchCycleBegun:
		if p.tracker.matchDispatchCycle {
			p.messageProduced(x.Envelope)
		}
	case fact.EventRecordedByAggregate:
		p.messageProduced(x.EventEnvelope)
	case fact.EventRecordedByIntegration:
		p.messageProduced(x.EventEnvelope)
	case fact.CommandExecutedByProcess:
		p.messageProduced(x.CommandEnvelope)
	}
}

// messageProduced updates the predicates's state to reflect the fact that a
// message has been produced.
func (p *messageTypePredicate) messageProduced(env *envelope.Envelope) {
	dist := typecmp.MeasureDistance(
		p.expectedType.ReflectType(),
		env.Type.ReflectType(),
	)

	if dist < p.bestMatchDistance {
		p.bestMatch = env
		p.bestMatchDistance = dist
	}

	if dist == typecmp.Identical && p.expectedRole == env.Role {
		p.ok = true
	}
}

func (p *messageTypePredicate) Ok() bool {
	return p.ok
}

func (p *messageTypePredicate) Done() {
}

func (p *messageTypePredicate) Report(treeOk bool) *Report {
	rep := &Report{
		TreeOk: treeOk,
		Ok:     p.ok,
		Criteria: inflect.Sprintf(
			p.expectedRole,
			"<produce> any '%s' <message>",
			p.expectedType,
		),
	}

	if treeOk || p.ok {
		return rep
	}

	if p.bestMatch == nil {
		reportNoMatch(rep, &p.tracker)
	} else if p.bestMatch.Role == p.expectedRole {
		p.reportExpectedRole(rep)
	} else {
		p.reportUnexpectedRole(rep)
	}

	return rep
}

// reportExpectedRole builds a test report when a "best-match" message was found
// and it has the expected role.
func (p *messageTypePredicate) reportExpectedRole(rep *Report) {
	s := rep.Section(suggestionsSection)

	if p.bestMatch.Origin == nil {
		rep.Explanation = inflect.Sprint(
			p.expectedRole,
			"a <message> of a similar type was <produced> via a <dispatcher>",
		)
	} else {
		rep.Explanation = inflect.Sprintf(
			p.expectedRole,
			"a <message> of a similar type was <produced> by the '%s' %s message handler",
			p.bestMatch.Origin.Handler.Identity().Name,
			p.bestMatch.Origin.Handler.HandlerType(),
		)
	}

	// note this language here is deliberately vague, it doesn't imply whether
	// it currently is or isn't a pointer, just questions if it should be.
	s.AppendListItem("check the message type, should it be a pointer?")

	p.buildDiff(rep)
}

// reportUnexpectedRole builds a test report when a "best-match" message was
// found but it does not have the expected role.
func (p *messageTypePredicate) reportUnexpectedRole(rep *Report) {
	s := rep.Section(suggestionsSection)

	if p.bestMatch.Origin == nil {
		s.AppendListItem(inflect.Sprint(
			p.bestMatch.Role,
			"verify that a <message> of this type was intended to be <produced> via a <dispatcher>",
		))
	} else {
		s.AppendListItem(inflect.Sprintf(
			p.bestMatch.Role,
			"verify that the '%s' %s message handler intended to <produce> a <message> of this type",
			p.bestMatch.Origin.Handler.Identity().Name,
			p.bestMatch.Origin.Handler.HandlerType(),
		))
	}

	if p.expectedRole == message.CommandRole {
		s.AppendListItem("verify that ToExecuteCommandOfType() is the correct expectation, did you mean ToRecordEventOfType()?")
	} else {
		s.AppendListItem("verify that ToRecordEventOfType() is the correct expectation, did you mean ToExecuteCommandOfType()?")
	}

	if p.bestMatchDistance == typecmp.Identical {
		if p.bestMatch.Origin == nil {
			rep.Explanation = inflect.Sprint(
				p.bestMatch.Role,
				"a message of this type was <produced> as a <message> via a <dispatcher>",
			)
		} else {
			rep.Explanation = inflect.Sprintf(
				p.bestMatch.Role,
				"a message of this type was <produced> as a <message> by the '%s' %s message handler",
				p.bestMatch.Origin.Handler.Identity().Name,
				p.bestMatch.Origin.Handler.HandlerType(),
			)
		}
	} else {
		if p.bestMatch.Origin == nil {
			rep.Explanation = inflect.Sprint(
				p.bestMatch.Role,
				"a message of a similar type was <produced> as a <message> via a <dispatcher>",
			)
		} else {
			rep.Explanation = inflect.Sprintf(
				p.bestMatch.Role,
				"a message of a similar type was <produced> as a <message> by the '%s' %s message handler",
				p.bestMatch.Origin.Handler.Identity().Name,
				p.bestMatch.Origin.Handler.HandlerType(),
			)
		}

		// note this language here is deliberately vague, it doesn't imply
		// whether it currently is or isn't a pointer, just questions if it
		// should be.
		s.AppendListItem("check the message type, should it be a pointer?")

		p.buildDiff(rep)
	}
}

// buildDiff adds a "message type diff" section to the result.
func (p *messageTypePredicate) buildDiff(rep *Report) {
	report.WriteDiff(
		&rep.Section("Message Type Diff").Content,
		p.expectedType.String(),
		p.bestMatch.Type.String(),
	)
}
