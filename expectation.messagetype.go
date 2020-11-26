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

func (e *messageTypeExpectation) Predicate(
	s PredicateScope,
	o PredicateOptions,
) (Predicate, error) {
	r, ok := s.App.MessageTypes().RoleOf(e.expectedType)

	// TODO: These checks should result in information being added to the
	// report, not just returning an error.
	//
	// See https://github.com/dogmatiq/testkit/issues/162
	if !ok {
		return nil, inflect.Errorf(
			e.expectedRole,
			"a <message> of type %s can never be <produced>, the application does not use this message type",
			e.expectedType,
		)
	} else if r != e.expectedRole {
		return nil, inflect.Errorf(
			e.expectedRole,
			"%s is a %s, it can never be <produced> as a <message>",
			e.expectedType,
			r,
		)
	} else if !o.MatchDispatchCycleStartedFacts {
		// If we're NOT matching messages from DispatchCycleStarted facts that
		// means this expectation can only ever pass if the message is produced
		// by a handler.
		if _, ok := s.App.MessageTypes().Produced[e.expectedType]; !ok {
			return nil, inflect.Errorf(
				e.expectedRole,
				"no handlers <produce> <messages> of type %s, it is only ever consumed",
				e.expectedType,
			)
		}
	}

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
		return rep
	}

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

	return rep
}

// buildDiff adds a "message type diff" section to the result.
func (p *messageTypePredicate) buildDiff(rep *Report) {
	report.WriteDiff(
		&rep.Section("Message Type Diff").Content,
		p.expectedType.String(),
		p.bestMatch.Type.String(),
	)
}
