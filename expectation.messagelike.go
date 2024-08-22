package testkit

import (
	"reflect"

	"github.com/dogmatiq/configkit/message"
	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/testkit/envelope"
	"github.com/dogmatiq/testkit/fact"
	"github.com/dogmatiq/testkit/internal/inflect"
	"github.com/dogmatiq/testkit/internal/report"
	"github.com/dogmatiq/testkit/internal/typecmp"
)

// ToRecordEventLike returns an expectation that passes if an event is recorded
// that is a superset of the given message.
//
// In order for the value 'A' to be a superset of the value 'B', both values
// must have the same type and meet the following type-specific conditions:
//
//   - If A and B are structs, each field in A must be a superset of the
//     corresponding field in B, unless the the field in B is the zero-value.
//   - If A and B are slices, for each element in B, A must contain a value that
//     is a superset of that element. The order of the elements is not relevant.
//   - If A and B are maps, for each element in B, A must have an element with
//     the same key, and the value must be a superset of the value in B.
//   - Otherwise, the values must compare as equivalent using the == operator.
//
// While these rules appear complex, they ultimately result in a comparison that
// behaves the same as [ToRecordEvent], but ignores any values omitted from m.
//
// m is not validated, as it is expected to be a "partially complete" message.
func ToRecordEventLike(m dogma.Event) Expectation {
	if m == nil {
		panic("ToRecordEventLike(<nil>): message must not be nil")
	}

	return &messageLikeExpectation{
		expectedSubsetMessage: m,
		expectedType:          message.TypeOf(m),
		expectedRole:          message.EventRole,
	}
}

type messageLikeExpectation struct {
	expectedSubsetMessage dogma.Message
	expectedType          message.Type
	expectedRole          message.Role
}

func (e *messageLikeExpectation) Caption() string {
	return inflect.Sprintf(
		e.expectedRole,
		"to <produce> a <message> that is a superset of a specific '%s' <message>",
		e.expectedType,
	)
}

func (e *messageLikeExpectation) Predicate(s PredicateScope) (Predicate, error) {
	return &messageLikePredicate{
		messageComparator:     s.Options.MessageComparator,
		expectedSubsetMessage: e.expectedSubsetMessage,
		expectedType:          e.expectedType,
		expectedRole:          e.expectedRole,
		bestMatchDistance:     typecmp.Unrelated,
		tracker: tracker{
			role:    e.expectedRole,
			options: s.Options,
		},
	}, validateRole(s, e.expectedType, e.expectedRole)
}

// messageLikePredicate is the Predicate implementation for messageExpectation.
type messageLikePredicate struct {
	messageComparator     MessageComparator
	expectedSubsetMessage dogma.Message
	expectedRole          message.Role
	expectedType          message.Type
	ok                    bool
	bestMatch             *envelope.Envelope
	bestMatchDistance     typecmp.Distance
	bestMatchIsEqual      bool
	tracker               tracker
}

// Notify updates the expectation's state in response to a new fact.
func (p *messageLikePredicate) Notify(f fact.Fact) {
	if p.ok {
		return
	}

	if env, ok := p.tracker.Notify(f); ok {
		p.messageProduced(env)
	}
}

// messageProduced updates the predicate's state to reflect the fact that a
// message has been produced.
func (p *messageLikePredicate) messageProduced(env *envelope.Envelope) {
	isEqual := p.messageComparator
	if isEqual == nil {
		isEqual = DefaultMessageComparator
	}

	if !isEqual(env.Message, p.expectedSubsetMessage) {
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
func (p *messageLikePredicate) updateBestMatch(env *envelope.Envelope) {
	dist := typecmp.MeasureDistance(
		reflect.TypeOf(p.expectedSubsetMessage),
		reflect.TypeOf(env.Message),
	)

	if dist < p.bestMatchDistance {
		p.bestMatch = env
		p.bestMatchDistance = dist
	}
}

func (p *messageLikePredicate) Ok() bool {
	return p.ok
}

func (p *messageLikePredicate) Done() {
}

func (p *messageLikePredicate) Report(ctx ReportGenerationContext) *Report {
	rep := &Report{
		TreeOk: ctx.TreeOk,
		Ok:     p.ok,
		Criteria: inflect.Sprintf(
			p.expectedRole,
			"<produce> a <message> that is a superset of a specific '%s' <message>",
			message.TypeOf(p.expectedSubsetMessage),
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

	p.buildDiff(ctx, rep)

	return rep
}

// buildDiff adds a "message diff" section to the result.
func (p *messageLikePredicate) buildDiff(ctx ReportGenerationContext, rep *Report) {
	report.WriteDiff(
		&rep.Section("Message Diff").Content,
		ctx.renderMessage(p.expectedSubsetMessage),
		ctx.renderMessage(p.bestMatch.Message),
	)
}
