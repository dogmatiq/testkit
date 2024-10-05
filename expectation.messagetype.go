package testkit

import (
	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/enginekit/message"
	"github.com/dogmatiq/testkit/fact"
	"github.com/dogmatiq/testkit/internal/inflect"
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
		expectedType: e.expectedType,
		tracker: tracker{
			kind:    e.expectedType.Kind(),
			options: s.Options,
		},
	}, nil
}

// messageTypePredicate is the Predicate implementation for
// messageTypeExpectation.
type messageTypePredicate struct {
	expectedType message.Type
	ok           bool
	tracker      tracker
}

func (p *messageTypePredicate) Notify(f fact.Fact) {
	if p.ok {
		return
	}

	if env, ok := p.tracker.Notify(f); ok {
		p.ok = message.TypeOf(env.Message) == p.expectedType
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

	reportNoMatch(rep, &p.tracker)
	return rep
}
