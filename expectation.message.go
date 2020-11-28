package testkit

import (
	"fmt"
	"reflect"

	"github.com/dogmatiq/configkit/message"
	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/testkit/envelope"
	"github.com/dogmatiq/testkit/fact"
	"github.com/dogmatiq/testkit/internal/inflect"
	"github.com/dogmatiq/testkit/internal/report"
	"github.com/dogmatiq/testkit/internal/typecmp"
)

// ToExecuteCommand returns an expectation that passes if a command is executed
// that is equal to m.
func ToExecuteCommand(m dogma.Message) Expectation {
	if err := dogma.ValidateMessage(m); err != nil {
		panic(fmt.Sprintf("ToExecuteCommand(%T): %s", m, err))
	}

	return &messageExpectation{
		expectedMessage: m,
		expectedType:    message.TypeOf(m),
		expectedRole:    message.CommandRole,
	}
}

// ToRecordEvent returns an expectation that passes if an event is recorded that
// is equal to m.
func ToRecordEvent(m dogma.Message) Expectation {
	if err := dogma.ValidateMessage(m); err != nil {
		panic(fmt.Sprintf("ToRecordEvent(%T): %s", m, err))
	}

	return &messageExpectation{
		expectedMessage: m,
		expectedType:    message.TypeOf(m),
		expectedRole:    message.EventRole,
	}
}

// messageTypeExpectation is an Expectation that checks that specific message is
// produced.
//
// It is the implementation used by ToExecuteCommand() and ToRecordEvent().
type messageExpectation struct {
	expectedMessage dogma.Message
	expectedType    message.Type
	expectedRole    message.Role
}

func (e *messageExpectation) Banner() string {
	return inflect.Sprintf(
		e.expectedRole,
		"TO <PRODUCE> A SPECIFIC '%s' <MESSAGE>",
		e.expectedType,
	)
}

func (e *messageExpectation) Predicate(s PredicateScope) (Predicate, error) {
	return &messagePredicate{
		expectedMessage:   e.expectedMessage,
		expectedType:      e.expectedType,
		expectedRole:      e.expectedRole,
		bestMatchDistance: typecmp.Unrelated,
		tracker: tracker{
			role:               e.expectedRole,
			matchDispatchCycle: s.Options.MatchDispatchCycleStartedFacts,
		},
	}, validateRole(s, e.expectedType, e.expectedRole)
}

// messagePredicate is the Predicate implementation for messageExpectation.
type messagePredicate struct {
	expectedMessage   dogma.Message
	expectedRole      message.Role
	expectedType      message.Type
	ok                bool
	bestMatch         *envelope.Envelope
	bestMatchDistance typecmp.Distance
	bestMatchIsEqual  bool
	tracker           tracker
}

// Notify updates the expectation's state in response to a new fact.
func (p *messagePredicate) Notify(f fact.Fact) {
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

// messageProduced updates the predicate's state to reflect the fact that a
// message has been produced.
func (p *messagePredicate) messageProduced(env *envelope.Envelope) {
	if !reflect.DeepEqual(env.Message, p.expectedMessage) {
		p.updateBestMatch(env)
		return
	}

	p.bestMatch = env
	p.bestMatchDistance = typecmp.Identical
	p.bestMatchIsEqual = true

	if env.Role == p.expectedRole {
		p.ok = true
	}
}

// updateBestMatch replaces p.bestMatch with env if it is a better match.
func (p *messagePredicate) updateBestMatch(env *envelope.Envelope) {
	dist := typecmp.MeasureDistance(
		reflect.TypeOf(p.expectedMessage),
		reflect.TypeOf(env.Message),
	)

	if dist < p.bestMatchDistance {
		p.bestMatch = env
		p.bestMatchDistance = dist
	}
}

func (p *messagePredicate) Ok() bool {
	return p.ok
}

func (p *messagePredicate) Done() {
}

func (p *messagePredicate) Report(treeOk bool) *Report {
	rep := &Report{
		TreeOk: treeOk,
		Ok:     p.ok,
		Criteria: inflect.Sprintf(
			p.expectedRole,
			"<produce> a specific '%s' <message>",
			message.TypeOf(p.expectedMessage),
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

	if p.bestMatchDistance == typecmp.Identical {
		if p.bestMatch.Origin == nil {
			rep.Explanation = inflect.Sprint(
				p.expectedRole,
				"a similar <message> was <produced> via a <dispatcher>",
			)
		} else {
			rep.Explanation = inflect.Sprintf(
				p.expectedRole,
				"a similar <message> was <produced> by the '%s' %s message handler",
				p.bestMatch.Origin.Handler.Identity().Name,
				p.bestMatch.Origin.HandlerType,
			)
		}

		s.AppendListItem("check the content of the message")
	} else {
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
				p.bestMatch.Origin.HandlerType,
			)
		}

		// note this language here is deliberately vague, it doesn't imply
		// whether it currently is or isn't a pointer, just questions if it
		// should be.
		s.AppendListItem("check the message type, should it be a pointer?")
	}

	p.buildDiff(rep)

	return rep
}

// buildDiff adds a "message diff" section to the result.
func (p *messagePredicate) buildDiff(rep *Report) {
	report.WriteDiff(
		&rep.Section("Message Diff").Content,
		report.RenderMessage(p.expectedMessage),
		report.RenderMessage(p.bestMatch.Message),
	)
}
