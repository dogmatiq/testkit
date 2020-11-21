package assert

import (
	"fmt"
	"reflect"

	"github.com/dogmatiq/configkit/message"
	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/testkit/compare"
	"github.com/dogmatiq/testkit/engine/envelope"
	"github.com/dogmatiq/testkit/engine/fact"
	"github.com/dogmatiq/testkit/internal/inflect"
	"github.com/dogmatiq/testkit/report"
)

// CommandExecuted returns an assertion that passes if m is executed as a
// command.
func CommandExecuted(m dogma.Message) Assertion {
	if err := dogma.ValidateMessage(m); err != nil {
		panic(fmt.Sprintf(
			"can not assert that this command will be executed, it is invalid: %s",
			err,
		))
	}

	return &messageAssertion{
		expected: m,
		role:     message.CommandRole,
	}
}

// EventRecorded returns an assertion that passes if m is recorded as an event.
func EventRecorded(m dogma.Message) Assertion {
	if err := dogma.ValidateMessage(m); err != nil {
		panic(fmt.Sprintf(
			"can not assert that this event will be recorded, it is invalid: %s",
			err,
		))
	}

	return &messageAssertion{
		expected: m,
		role:     message.EventRole,
	}
}

// messageAssertion asserts that a specific message is produced.
type messageAssertion struct {
	// Expected is the message that is expected to be produced.
	expected dogma.Message

	// Role is the expected role of the expected message.
	// It must be either CommandRole or EventRole.
	role message.Role

	// cmp is the comparator used to compare messages for equality.
	cmp compare.Comparator

	// ok is true once the assertion is deemed to have passed, after which no
	// further updates are performed.
	ok bool

	// best is an envelope containing the "best-match" message for an assertion
	// that has not yet passed. Note that this message may not have the expected
	// role.
	best *envelope.Envelope

	// sim is a ranking of the similarity between the type of the expected
	// message, and the current best-match.
	sim compare.TypeSimilarity

	// equal is true if the best-match message compared as equal to the expected
	// message. This can occur, and the assertion still fail, if the best-match
	// message has an unexpected role.
	equal bool

	// tracker observers the handlers and messages that are involved in the
	// test.
	tracker tracker
}

// Banner returns a human-readable banner to display in the logs when this
// expectation is used.
//
// The banner text should be in uppercase, and complete the sentence "The
// application is expected ...". For example, "TO DO A THING".
func (a *messageAssertion) Banner() string {
	return inflect.Sprintf(
		a.role,
		"TO <PRODUCE> A SPECIFIC '%s' <MESSAGE>",
		message.TypeOf(a.expected),
	)
}

// Begin is called to prepare the assertion for a new test.
func (a *messageAssertion) Begin(o ExpectOptionSet) {
	// reset everything
	*a = messageAssertion{
		expected: a.expected,
		role:     a.role,
		cmp:      o.MessageComparator,
		tracker: tracker{
			role:               a.role,
			matchDispatchCycle: o.MatchMessagesInDispatchCycle,
		},
	}
}

// End is called once the test is complete.
func (a *messageAssertion) End() {
}

// Ok returns true if the assertion passed.
func (a *messageAssertion) Ok() bool {
	return a.ok
}

// BuildReport generates a report about the assertion.
//
// ok is true if the assertion is considered to have passed. This may not be the
// same value as returned from Ok() when this assertion is used as a
// sub-assertion inside a composite.
func (a *messageAssertion) BuildReport(ok bool) *Report {
	rep := &Report{
		TreeOk: ok,
		Ok:     a.ok,
		Criteria: inflect.Sprintf(
			a.role,
			"<produce> a specific '%s' <message>",
			message.TypeOf(a.expected),
		),
	}

	if ok || a.ok {
		return rep
	}

	if a.best == nil {
		buildResultNoMatch(rep, &a.tracker)
	} else if a.best.Role == a.role {
		a.buildResultExpectedRole(rep)
	} else {
		a.buildResultUnexpectedRole(rep)
	}

	return rep
}

// Notify updates the assertion's state in response to a new fact.
func (a *messageAssertion) Notify(f fact.Fact) {
	if a.ok {
		return
	}

	a.tracker.Notify(f)

	switch x := f.(type) {
	case fact.DispatchCycleBegun:
		if a.tracker.matchDispatchCycle {
			a.messageProduced(x.Envelope)
		}
	case fact.EventRecordedByAggregate:
		a.messageProduced(x.EventEnvelope)
	case fact.EventRecordedByIntegration:
		a.messageProduced(x.EventEnvelope)
	case fact.CommandExecutedByProcess:
		a.messageProduced(x.CommandEnvelope)
	}
}

// messageProduced updates the assertion's state to reflect the fact that a
// message has been produced.
func (a *messageAssertion) messageProduced(env *envelope.Envelope) {
	if !a.cmp.MessageIsEqual(env.Message, a.expected) {
		a.updateBestMatch(env)
		return
	}

	a.best = env
	a.sim = compare.SameTypes
	a.equal = true

	if a.role == env.Role {
		a.ok = true
	}
}

// updateBestMatch replaces a.best with env if it is a better match.
func (a *messageAssertion) updateBestMatch(env *envelope.Envelope) {
	sim := compare.FuzzyTypeComparison(
		reflect.TypeOf(a.expected),
		reflect.TypeOf(env.Message),
	)

	if sim > a.sim {
		a.best = env
		a.sim = sim
	}
}

// buildResultExpectedRole builds the assertion result when there is a
// "best-match" message available and it is of the expected role.
func (a *messageAssertion) buildResultExpectedRole(rep *Report) {
	s := rep.Section(suggestionsSection)

	if a.sim == compare.SameTypes {
		if a.best.Origin == nil {
			rep.Explanation = inflect.Sprint(
				a.role,
				"a similar <message> was <produced> via a <dispatcher>",
			)
		} else {
			rep.Explanation = inflect.Sprintf(
				a.role,
				"a similar <message> was <produced> by the '%s' %s message handler",
				a.best.Origin.Handler.Identity().Name,
				a.best.Origin.HandlerType,
			)
		}

		s.AppendListItem("check the content of the message")
	} else {
		if a.best.Origin == nil {
			rep.Explanation = inflect.Sprint(
				a.role,
				"a <message> of a similar type was <produced> via a <dispatcher>",
			)
		} else {
			rep.Explanation = inflect.Sprintf(
				a.role,
				"a <message> of a similar type was <produced> by the '%s' %s message handler",
				a.best.Origin.Handler.Identity().Name,
				a.best.Origin.HandlerType,
			)
		}

		// note this language here is deliberately vague, it doesn't imply
		// whether it currently is or isn't a pointer, just questions if it
		// should be.
		s.AppendListItem("check the message type, should it be a pointer?")
	}

	a.buildDiff(rep)
}

// buildDiff adds a "message diff" section to the result.
func (a *messageAssertion) buildDiff(rep *Report) {
	report.WriteDiff(
		&rep.Section("Message Diff").Content,
		report.RenderMessage(a.expected),
		report.RenderMessage(a.best.Message),
	)
}

// buildResultUnexpectedRole builds the assertion result when there is a
// "best-match" message available but it is of an unexpected role.
func (a *messageAssertion) buildResultUnexpectedRole(rep *Report) {
	s := rep.Section(suggestionsSection)

	if a.best.Origin == nil {
		s.AppendListItem(inflect.Sprint(
			a.best.Role,
			"verify that a <message> of this type was intended to be <produced> via a <dispatcher>",
		))
	} else {
		s.AppendListItem(inflect.Sprintf(
			a.best.Role,
			"verify that the '%s' %s message handler intended to <produce> a <message> of this type",
			a.best.Origin.Handler.Identity().Name,
			a.best.Origin.HandlerType,
		))
	}

	if a.role == message.CommandRole {
		s.AppendListItem("verify that CommandExecuted() is the correct assertion, did you mean EventRecorded()?")
	} else {
		s.AppendListItem("verify that EventRecorded() is the correct assertion, did you mean CommandExecuted()?")
	}

	// the "best-match" is equal to the expected message. this means that only
	// the roles were mismatched.
	if a.equal {
		if a.best.Origin == nil {
			rep.Explanation = inflect.Sprint(
				a.best.Role,
				"the expected message was <produced> as a <message> via a <dispatcher>",
			)
		} else {
			rep.Explanation = inflect.Sprintf(
				a.best.Role,
				"the expected message was <produced> as a <message> by the '%s' %s message handler",
				a.best.Origin.Handler.Identity().Name,
				a.best.Origin.HandlerType,
			)
		}

		return // skip diff rendering
	}

	if a.sim == compare.SameTypes {
		if a.best.Origin == nil {
			rep.Explanation = inflect.Sprint(
				a.best.Role,
				"a similar message was <produced> as a <message> via a <dispatcher>",
			)
		} else {
			rep.Explanation = inflect.Sprintf(
				a.best.Role,
				"a similar message was <produced> as a <message> by the '%s' %s message handler",
				a.best.Origin.Handler.Identity().Name,
				a.best.Origin.HandlerType,
			)
		}
	} else {
		if a.best.Origin == nil {
			rep.Explanation = inflect.Sprint(
				a.best.Role,
				"a message of a similar type was <produced> as a <message> via a <dispatcher>",
			)
		} else {
			rep.Explanation = inflect.Sprintf(
				a.best.Role,
				"a message of a similar type was <produced> as a <message> by the '%s' %s message handler",
				a.best.Origin.Handler.Identity().Name,
				a.best.Origin.HandlerType,
			)
		}

		// note this language here is deliberately vague, it doesn't imply
		// whether it currently is or isn't a pointer, just questions if it
		// should be.
		s.AppendListItem("check the message type, should it be a pointer?")
	}

	a.buildDiff(rep)
}
