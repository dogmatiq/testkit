package assert

import (
	"fmt"
	"io"
	"reflect"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/dogmatest/compare"
	"github.com/dogmatiq/dogmatest/engine/envelope"
	"github.com/dogmatiq/dogmatest/engine/fact"
	"github.com/dogmatiq/dogmatest/render"
	"github.com/dogmatiq/enginekit/message"
	"github.com/dogmatiq/iago"
)

// MessageAssertion is an assertion that requires an exact message to be
// produced.
type MessageAssertion struct {
	Message dogma.Message
	Role    message.Role

	cmp      compare.Comparator     // the comparator to use for message comparison
	pass     bool                   // true if assertion passed
	best     *envelope.Envelope     // envelope containing "best-match"
	sim      compare.TypeSimilarity // similarity between the types of the expected and best-match message
	equal    bool                   // true if the best-match message is equal to the expected message
	routed   bool                   // true if the message-under-test gets routed to at least one handler
	produced map[message.Role]bool  // map of the roles of messages that are produced
}

// Notify notifies the assertion of the occurrence of a fact.
func (a *MessageAssertion) Notify(f fact.Fact) {
	if a.pass {
		return
	}

	switch x := f.(type) {
	case fact.MessageHandlingBegun:
		a.routed = true
	case fact.EventRecordedByAggregate:
		a.update(x.EventEnvelope)
	case fact.EventRecordedByIntegration:
		a.update(x.EventEnvelope)
	case fact.CommandExecutedByProcess:
		a.update(x.CommandEnvelope)
	}
}

func (a *MessageAssertion) update(env *envelope.Envelope) {
	a.produced[env.Role] = true

	// look for an identical message
	if a.cmp.MessageIsEqual(env.Message, a.Message) {
		if env.Role == a.Role {
			// if it's the right message role, we've found our match
			a.pass = true
			return
		}

		// otherwise, this will always be the best match, assuming we don't find an
		// actual match
		a.best = env
		a.sim = compare.SameTypes
		a.equal = true
		return
	}

	// check to see if this message is of a similar type to our expected message
	sim := compare.FuzzyTypeComparison(
		reflect.TypeOf(a.Message),
		reflect.TypeOf(env.Message),
	)

	if sim > a.sim {
		a.best = env
		a.sim = sim
	}
}

// Begin is called before the message-under-test is dispatched.
func (a *MessageAssertion) Begin(c compare.Comparator) {
	a.Role.MustBe(message.CommandRole, message.EventRole)

	a.cmp = c
	a.pass = false
	a.best = nil
	a.sim = 0
	a.equal = false
	a.routed = false
	a.produced = map[message.Role]bool{}
}

// End is called after the message-under-test is dispatched.
func (a *MessageAssertion) End(w io.Writer, r render.Renderer) bool {
	rep := &report{Pass: a.pass}
	a.buildReport(rep, r)

	iago.MustWriteTo(w, rep)

	return a.pass
}

// buildReport populates rep with the result of the assertion.
func (a *MessageAssertion) buildReport(rep *report, r render.Renderer) {
	mt := message.TypeOf(a.Message)

	rep.Title = byRole(
		a.Role,
		fmt.Sprintf("execute a specific '%s' command", mt),
		fmt.Sprintf("record a specific '%s' event", mt),
	)

	if a.pass || a.best == nil {
		rep.Details = renderMessage(r, a.Message)
	} else {
		rep.Details = renderDiff(
			renderMessage(r, a.Message),
			renderMessage(r, a.best.Message),
		)
	}

	if a.pass {
		return
	}

	rep.SubTitle = byRole(
		a.Role,
		"this command was not executed",
		"this event was not recorded",
	)

	// the "best match" is equal to the expected message. this means that only the
	// roles were mismatched.
	if a.equal {
		rep.Outcome = byRole(
			a.best.Role,
			"this message was executed as a command",
			"this message was recorded as an event",
		)

		a.addWrongAssertionHint(rep)

		return
	}

	// there is no "best match". if any messages were produced at all they weren't
	// of a related type.
	if a.sim == compare.UnrelatedTypes {
		// nothing was produced at all
		if len(a.produced) == 0 {
			rep.Outcome = "no commands or events were produced at all"

			// if the message did get routed somewhere, it's probably a legitimate bug
			// with the business logic, otherwise there's a possibility that the routing
			// configuration is wrong.
			if a.routed {
				rep.suggest("check the application logic")
			} else {
				rep.suggest("check the application routing configuration for this type")
			}

			return
		}

		// if messages of the correct role were produced, perhaps there's just a
		// simple mispelling of the type. this is common because many messages have
		// the same fields.
		if a.produced[a.Role] {
			rep.suggest("check the assertion's message type")
		} else {
			a.addWrongAssertionHint(rep)
		}

		return
	}

	// the messages weren't equal, but a message of the exact same type occurred
	if a.sim == compare.SameTypes {
		rep.suggest("check the message content")

		// if the roles are equal, that means only the content was incorrect
		if a.Role == a.best.Role {
			rep.Outcome = byRole(
				a.Role,
				"a similar command was executed",
				"a similar event was recorded",
			)

			return
		}

		// otherwise, both the role and the content are wrong
		rep.Outcome = byRole(
			a.best.Role,
			"a similar message was executed as a command",
			"a similar message was recorded as an event",
		)

		a.addWrongAssertionHint(rep)

		return
	}

	// finally, a message of a similar type did occur. the content may or may not
	// be the same.

	// note this language here is deliberately vague, it doesn't imply whether it
	// currently is or isn't a pointer, just questions if it should be.
	rep.suggest("check the assertion's message type, should it be a pointer?")

	if a.Role == a.best.Role {
		rep.Outcome = byRole(
			a.Role,
			"a similar command was executed",
			"a similar event was recorded",
		)

		return
	}

	// otherwise, both the role and the content are wrong
	rep.Outcome = byRole(
		a.best.Role,
		"a similar message was executed as a command",
		"a similar message was recorded as an event",
	)

	a.addWrongAssertionHint(rep)
}

// addWrongAssertionHint adds a common hint about selecting the correct assertion.
func (a *MessageAssertion) addWrongAssertionHint(rep *report) {
	rep.suggest(byRole(
		a.Role,
		"did you mean to use the EventRecorded assertion instead of CommandExecuted?",
		"did you mean to use the CommandExecuted assertion instead of EventRecorded?",
	))
}
