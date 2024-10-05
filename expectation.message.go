package testkit

import (
	"fmt"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/enginekit/message"
	"github.com/dogmatiq/testkit/envelope"
	"github.com/dogmatiq/testkit/fact"
	"github.com/dogmatiq/testkit/internal/inflect"
	"github.com/dogmatiq/testkit/internal/report"
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
	tracker           tracker
}

// Notify updates the expectation's state in response to a new fact.
func (p *messagePredicate) Notify(f fact.Fact) {
	if p.ok {
		return
	}

	env, ok := p.tracker.Notify(f)
	if !ok {
		return
	}

	if message.TypeOf(env.Message) != message.TypeOf(p.expectedMessage) {
		return
	}

	isEqual := p.messageComparator
	if isEqual == nil {
		isEqual = DefaultMessageComparator
	}

	p.bestMatch = env
	p.ok = isEqual(env.Message, p.expectedMessage)
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
