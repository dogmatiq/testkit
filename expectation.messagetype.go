package testkit

import (
	"github.com/dogmatiq/configkit/message"
	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/testkit/envelope"
	"github.com/dogmatiq/testkit/fact"
	"github.com/dogmatiq/testkit/internal/inflect"
	"github.com/dogmatiq/testkit/internal/report"
	"github.com/dogmatiq/testkit/internal/typecmp"
)

// ToExecuteCommandType returns an expectation that passes if a command of type
// T is executed.
func ToExecuteCommandType[T dogma.Command]() Expectation {
	return &messageTypeExpectation{
		expectedType: message.TypeFor[T](),
	}
}

// ToExecuteCommandOfType returns an expectation that passes if a command of the
// same type as m is executed.
//
// Deprecated: Use [ToExecuteCommandType] instead.
func ToExecuteCommandOfType(m dogma.Command) Expectation {
	if m == nil {
		panic("ToExecuteCommandOfType(<nil>): message must not be nil")
	}

	return &messageTypeExpectation{
		expectedType: message.TypeOf(m),
	}
}

// ToRecordEventType returns an expectation that passes if an event of type T is
// recorded.
func ToRecordEventType[T dogma.Event]() Expectation {
	return &messageTypeExpectation{
		expectedType: message.TypeFor[T](),
	}
}

// ToRecordEventOfType returns an expectation that passes if an event of the
// same type as m is recorded.
//
// Deprecated: Use [ToRecordEventType] instead.
func ToRecordEventOfType(m dogma.Event) Expectation {
	if m == nil {
		panic("ToRecordEventOfType(<nil>): message must not be nil")
	}

	return &messageTypeExpectation{
		expectedType: message.TypeOf(m),
	}
}

// messageTypeExpectation is an Expectation that checks that a message of a
// specific type is produced.
//
// It is the implementation used by ToExecuteCommandOfType() and
// ToRecordEventOfType().
type messageTypeExpectation struct {
	expectedType message.Type
}

func (e *messageTypeExpectation) Caption() string {
	return inflect.Sprintf(
		e.expectedType.Kind(),
		"to <produce> a <message> of type %s",
		e.expectedType,
	)
}

func (e *messageTypeExpectation) Predicate(s PredicateScope) (Predicate, error) {
	if err := guardAgainstExpectationOnImpossibleType(s, e.expectedType); err != nil {
		return nil, err
	}

	return &messageTypePredicate{
		expectedType:      e.expectedType,
		bestMatchDistance: typecmp.Unrelated,
		tracker: tracker{
			kind:    e.expectedType.Kind(),
			options: s.Options,
		},
	}, nil
}

// messageTypePredicate is the Predicate implementation for
// messageTypeExpectation.
type messageTypePredicate struct {
	expectedType      message.Type
	ok                bool
	bestMatch         *envelope.Envelope
	bestMatchDistance typecmp.Distance
	tracker           tracker
}

func (p *messageTypePredicate) Notify(f fact.Fact) {
	if p.ok {
		return
	}

	if env, ok := p.tracker.Notify(f); ok {
		p.messageProduced(env)
	}
}

// messageProduced updates the predicates's state to reflect the fact that a
// message has been produced.
func (p *messageTypePredicate) messageProduced(env *envelope.Envelope) {
	producedType := message.TypeOf(env.Message)

	dist := typecmp.MeasureDistance(
		p.expectedType.ReflectType(),
		producedType.ReflectType(),
	)

	if dist < p.bestMatchDistance {
		p.bestMatch = env
		p.bestMatchDistance = dist
	}

	if dist == typecmp.Identical {
		p.ok = true
	}
}

func (p *messageTypePredicate) Ok() bool {
	return p.ok
}

func (p *messageTypePredicate) Done() {
}

func (p *messageTypePredicate) Report(ctx ReportGenerationContext) *Report {
	rep := &Report{
		TreeOk: ctx.TreeOk,
		Ok:     p.ok,
		Criteria: inflect.Sprintf(
			p.expectedType.Kind(),
			"<produce> any '%s' <message>",
			p.expectedType,
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
			p.expectedType.Kind(),
			"a <message> of a similar type was <produced> via a <dispatcher>",
		)
	} else {
		rep.Explanation = inflect.Sprintf(
			p.expectedType.Kind(),
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
		message.TypeOf(p.bestMatch.Message).String(),
	)
}
