package assert

// import (
// 	"fmt"
// 	"io"
// 	"reflect"

// 	"github.com/dogmatiq/dogmatest/compare"
// 	"github.com/dogmatiq/dogmatest/engine/envelope"
// 	"github.com/dogmatiq/dogmatest/engine/fact"
// 	"github.com/dogmatiq/dogmatest/render"
// 	"github.com/dogmatiq/enginekit/message"
// 	"github.com/dogmatiq/iago"
// )

// // MessageTypeAssertion is an assertion that requires an exact message type to
// // be produced.
// type MessageTypeAssertion struct {
// 	Type message.Type
// 	Role message.Role

// 	pass     bool                   // true if assertion passed
// 	best     *envelope.Envelope     // envelope containing "best-match"
// 	sim      compare.TypeSimilarity // similarity between the expected type and the best-match
// 	routed   bool                   // true if the message-under-test gets routed to at least one handler
// 	produced map[message.Role]bool  // map of the roles of messages that are produced
// }

// // Notify notifies the assertion of the occurrence of a fact.
// func (a *MessageTypeAssertion) Notify(f fact.Fact) {
// 	if a.pass {
// 		return
// 	}

// 	switch x := f.(type) {
// 	case fact.MessageHandlingBegun:
// 		a.routed = true
// 	case fact.EventRecordedByAggregate:
// 		a.update(x.EventEnvelope)
// 	case fact.EventRecordedByIntegration:
// 		a.update(x.EventEnvelope)
// 	case fact.CommandExecutedByProcess:
// 		a.update(x.CommandEnvelope)
// 	}
// }

// func (a *MessageTypeAssertion) update(env *envelope.Envelope) {
// 	a.produced[env.Role] = true

// 	// check to see if this message is of a similar type to our expected message
// 	sim := compare.FuzzyTypeComparison(
// 		a.Type.ReflectType(),
// 		reflect.TypeOf(env.Message),
// 	)

// 	// look for an identical message type
// 	if sim == compare.SameTypes && env.Role == a.Role {
// 		// if it's the right message role, we've found our match
// 		a.pass = true
// 		return
// 	}

// 	if sim > a.sim {
// 		a.best = env
// 		a.sim = sim
// 	}
// }

// // Begin is called before the message-under-test is dispatched.
// func (a *MessageTypeAssertion) Begin(c compare.Comparator) {
// 	a.Role.MustBe(message.CommandRole, message.EventRole)

// 	a.pass = false
// 	a.best = nil
// 	a.sim = 0
// 	a.routed = false
// 	a.produced = map[message.Role]bool{}
// }

// // End is called after the message-under-test is dispatched.
// func (a *MessageTypeAssertion) End(w io.Writer, r render.Renderer) bool {
// 	rep := &report{Pass: a.pass}
// 	a.buildReport(rep, r)

// 	iago.MustWriteTo(w, rep)

// 	return a.pass
// }

// // buildReport populates rep with the result of the assertion.
// func (a *MessageTypeAssertion) buildReport(rep *report, r render.Renderer) {
// 	rep.Title = byRole(
// 		a.Role,
// 		fmt.Sprintf("execute any '%s' command", a.Type),
// 		fmt.Sprintf("record any '%s' event", a.Type),
// 	)

// 	if a.best != nil {
// 		rep.Details = renderDiff(
// 			a.Type.String(),
// 			a.best.Type.String(),
// 		)
// 	}

// 	if a.pass {
// 		return
// 	}

// 	rep.SubTitle = byRole(
// 		a.Role,
// 		"no commands of this type were executed",
// 		"no events of this type were recorded",
// 	)

// 	// there is no "best match". if any messages were produced at all they weren't
// 	// of a related type.
// 	if a.sim == compare.UnrelatedTypes {
// 		// nothing was produced at all
// 		if len(a.produced) == 0 {
// 			rep.Outcome = "no commands or events were produced at all"

// 			// if the message did get routed somewhere, it's probably a legitimate bug
// 			// with the business logic, otherwise there's a possibility that the routing
// 			// configuration is wrong.
// 			if a.routed {
// 				rep.suggest("check the application logic")
// 			} else {
// 				rep.suggest("check the application routing configuration for this type")
// 			}

// 			return
// 		}

// 		// if messages of the correct role were produced, perhaps there's just a
// 		// simple mispelling of the type. this is common because many messages have
// 		// the same fields.
// 		if a.produced[a.Role] {
// 			rep.suggest("check the assertion's message type")
// 		} else {
// 			a.addWrongAssertionHint(rep)
// 		}

// 		return
// 	}

// 	// the types are equal, so the roles must be a mismatch
// 	if a.sim == compare.SameTypes {
// 		rep.Outcome = byRole(
// 			a.best.Role,
// 			"a message of this type was executed as a command",
// 			"a message of this type was recorded as an event",
// 		)

// 		a.addWrongAssertionHint(rep)

// 		return
// 	}

// 	// finally, a message of a similar type did occur.

// 	// note this language here is deliberately vague, it doesn't imply whether it
// 	// currently is or isn't a pointer, just questions if it should be.
// 	rep.suggest("check the assertion's message type, should it be a pointer?")

// 	if a.Role == a.best.Role {
// 		rep.Outcome = byRole(
// 			a.Role,
// 			"a command of a similar type was executed",
// 			"an event of a similar type was recorded",
// 		)

// 		return
// 	}

// 	// otherwise, both the role and the content are wrong
// 	rep.Outcome = byRole(
// 		a.best.Role,
// 		"a message of a similar type was executed as a command",
// 		"a message of a similar type was recorded as an event",
// 	)

// 	a.addWrongAssertionHint(rep)
// }

// // addWrongAssertionHint adds a common hint about selecting the correct assertion.
// func (a *MessageTypeAssertion) addWrongAssertionHint(rep *report) {
// 	rep.suggest(byRole(
// 		a.Role,
// 		"did you mean to use the EventTypeRecorded assertion instead of CommandTypeExecuted?",
// 		"did you mean to use the CommandTypeExecuted assertion instead of EventTypeRecorded?",
// 	))
// }
