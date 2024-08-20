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
		expectedRole: message.CommandRole,
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
		expectedRole: message.CommandRole,
	}
}

// ToRecordEventType returns an expectation that passes if an event of type T is
// recorded.
func ToRecordEventType[T dogma.Event]() Expectation {
	return &messageTypeExpectation{
		expectedType: message.TypeFor[T](),
		expectedRole: message.EventRole,
	}
}

// ToRecordEventOfType returns an expectation that passes if an event of the
// same type as m is recorded.
//
// Deprecated: Use [ToRecordEventType] instead.
func ToRecordEventOfType(m dogma.Command) Expectation {
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

func (e *messageTypeExpectation) Caption() string {
	return inflect.Sprintf(
		e.expectedRole,
		"to <produce> a <message> of type %s",
		e.expectedType,
	)
}

func (e *messageTypeExpectation) Predicate(s PredicateScope) (Predicate, error) {
	return &messageTypePredicate{
		expectedType:      e.expectedType,
		expectedRole:      e.expectedRole,
		bestMatchDistance: typecmp.Unrelated,
		tracker: tracker{
			role:    e.expectedRole,
			options: s.Options,
		},
	}, validateRole(s, e.expectedType, e.expectedRole)
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

	if env, ok := p.tracker.Notify(f); ok {
		p.messageProduced(env)
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

func (p *messageTypePredicate) Report(ctx ReportGenerationContext) *Report {
	rep := &Report{
		TreeOk: ctx.TreeOk,
		Ok:     p.ok,
		Criteria: inflect.Sprintf(
			p.expectedRole,
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
