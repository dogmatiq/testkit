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
	"github.com/dogmatiq/testkit/internal/validation"
)

// ToExecuteCommand returns an expectation that passes if a command is executed
// that is equal to m.
func ToExecuteCommand(m dogma.Command) Expectation {
	if m == nil {
		panic("ToExecuteCommand(<nil>): message must not be nil")
	}

	mt := message.TypeOf(m)

	if err := m.Validate(validation.CommandValidationScope()); err != nil {
		panic(fmt.Sprintf("ToExecuteCommand(%s): %s", mt, err))
	}

	return &messageExpectation{
		expectedMessage: m,
	}
}

// ToRecordEvent returns an expectation that passes if an event is recorded that
// is equal to m.
func ToRecordEvent(m dogma.Event) Expectation {
	if m == nil {
		panic("ToRecordEvent(<nil>): message must not be nil")
	}

	mt := message.TypeOf(m)

	if err := m.Validate(validation.EventValidationScope()); err != nil {
		panic(fmt.Sprintf("ToRecordEvent(%s): %s", mt, err))
	}

	return &messageExpectation{
		expectedMessage: m,
	}
}

// messageTypeExpectation is an Expectation that checks that specific message is
// produced.
//
// It is the implementation used by ToExecuteCommand() and ToRecordEvent().
type messageExpectation struct {
	expectedMessage dogma.Message
}

func (e *messageExpectation) Caption() string {
	return inflect.Sprintf(
		message.KindOf(e.expectedMessage),
		"to <produce> a specific '%s' <message>",
		message.TypeOf(e.expectedMessage),
	)
}

func (e *messageExpectation) Predicate(s PredicateScope) (Predicate, error) {
	mt := message.TypeOf(e.expectedMessage)

	if err := guardAgainstExpectationOnImpossibleType(s, mt); err != nil {
		return nil, err
	}

	return &messagePredicate{
		messageComparator: s.Options.MessageComparator,
		expectedMessage:   e.expectedMessage,
		bestMatchDistance: typecmp.Unrelated,
		tracker: tracker{
			kind:    mt.Kind(),
			options: s.Options,
		},
	}, nil
}

// messagePredicate is the Predicate implementation for messageExpectation.
type messagePredicate struct {
	messageComparator MessageComparator
	expectedMessage   dogma.Message
	ok                bool
	bestMatch         *envelope.Envelope
	bestMatchDistance typecmp.Distance
	tracker           tracker
}

// Notify updates the expectation's state in response to a new fact.
func (p *messagePredicate) Notify(f fact.Fact) {
	if p.ok {
		return
	}

	if env, ok := p.tracker.Notify(f); ok {
		p.messageProduced(env)
	}
}

// messageProduced updates the predicate's state to reflect the fact that a
// message has been produced.
func (p *messagePredicate) messageProduced(env *envelope.Envelope) {
	isEqual := p.messageComparator
	if isEqual == nil {
		isEqual = DefaultMessageComparator
	}

	if !isEqual(env.Message, p.expectedMessage) {
		p.updateBestMatch(env)
		return
	}

	p.bestMatch = env
	p.bestMatchDistance = typecmp.Identical
	p.ok = true
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

func (p *messagePredicate) Report(ctx ReportGenerationContext) *Report {
	mt := message.TypeOf(p.expectedMessage)

	rep := &Report{
		TreeOk: ctx.TreeOk,
		Ok:     p.ok,
		Criteria: inflect.Sprintf(
			mt.Kind(),
			"<produce> a specific '%s' <message>",
			mt,
		),
	}

	if p.ok || ctx.TreeOk || ctx.IsInverted {
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
				mt.Kind(),
				"a similar <message> was <produced> via a <dispatcher>",
			)
		} else {
			rep.Explanation = inflect.Sprintf(
				mt.Kind(),
				"a similar <message> was <produced> by the '%s' %s message handler",
				p.bestMatch.Origin.Handler.Identity().Name,
				p.bestMatch.Origin.HandlerType,
			)
		}

		s.AppendListItem("check the content of the message")
	} else {
		if p.bestMatch.Origin == nil {
			rep.Explanation = inflect.Sprint(
				mt.Kind(),
				"a <message> of a similar type was <produced> via a <dispatcher>",
			)
		} else {
			rep.Explanation = inflect.Sprintf(
				mt.Kind(),
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

	p.buildDiff(ctx, rep)

	return rep
}

// buildDiff adds a "message diff" section to the result.
func (p *messagePredicate) buildDiff(ctx ReportGenerationContext, rep *Report) {
	report.WriteDiff(
		&rep.Section("Message Diff").Content,
		ctx.renderMessage(p.expectedMessage),
		ctx.renderMessage(p.bestMatch.Message),
	)
}
