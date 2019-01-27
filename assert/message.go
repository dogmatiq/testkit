package assert

import (
	"fmt"
	"io"
	"reflect"

	"github.com/dogmatiq/iago"
	"github.com/dogmatiq/iago/indent"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/dogmatest/compare"
	"github.com/dogmatiq/dogmatest/engine/envelope"
	"github.com/dogmatiq/dogmatest/engine/fact"
	"github.com/dogmatiq/dogmatest/render"
	"github.com/dogmatiq/enginekit/message"
)

// MessageAssertion is an assertion that requires an exact message to be
// produced.
type MessageAssertion struct {
	Message dogma.Message
	Role    message.Role

	cmp compare.Comparator

	pass bool
	best *envelope.Envelope
	sim  compare.TypeSimilarity
	eq   bool
}

// Notify notifies the assertion of the occurrence of a fact.
func (a *MessageAssertion) Notify(f fact.Fact) {
	if a.pass {
		return
	}

	switch x := f.(type) {
	case fact.EventRecordedByAggregate:
		a.update(x.EventEnvelope)
	case fact.EventRecordedByIntegration:
		a.update(x.EventEnvelope)
	case fact.CommandExecutedByProcess:
		a.update(x.CommandEnvelope)
	}
}

func (a *MessageAssertion) update(env *envelope.Envelope) {
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
		a.eq = true
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
	a.eq = false
}

// End is called after the message-under-test is dispatched.
func (a *MessageAssertion) End(w io.Writer, r render.Renderer) bool {
	writeIcon(w, a.pass)

	// write a description of the assertion
	mt := message.TypeOf(a.Message)
	writeByRole(
		w,
		a.Role,
		fmt.Sprintf(" execute specific '%s' command", mt),
		fmt.Sprintf(" record specific '%s' command", mt),
	)

	// we found the exact message we expected
	if a.pass {
		iago.MustWriteString(w, "\n")
		return true
	}

	// write the failure message
	writeByRole(
		w,
		a.Role,
		" (this command was not executed)\n",
		" (this event was not recorded)\n",
	)

	// write a diff if we have a "best match", otherwise write the expected message
	iw := indent.NewIndenter(w, []byte("| "))
	if a.best == nil {
		iago.Must(r.WriteMessage(iw, a.Message))
	} else {
		writeDiff(iw, r, a.Message, a.best.Message)
	}

	// write a hint about how the failure might be fixed
	if a.eq {
		writeHintByRole(
			w,
			a.Role,
			a.best.Role,
			"",
			"This message was executed as a command, did you mean to use the ExpectCommand() assertion instead of ExpectEvent()?\n",
			"This message was recorded as an event, did you mean to use the ExpectEvent() assertion instead of ExpectCommand()?\n",
		)
	} else if a.sim == compare.SameTypes {
		writeHintByRole(
			w,
			a.Role,
			a.best.Role,
			"Check the content of the message.",
			"A similar message was executed as a command, did you mean to use the ExpectCommand() assertion instead of ExpectEvent()?\n",
			"A similar message was recorded as an event, did you mean to use the ExpectEvent() assertion instead of ExpectCommand()?\n",
		)
	} else if a.sim != compare.UnrelatedTypes {
		writeHintByRole(
			w,
			a.Role,
			a.best.Role,
			"Check the type of the message.",
			"A message of a similar type was executed as a command, did you mean to use the ExpectCommand() assertion instead of ExpectEvent()?\n",
			"A message of a similar type was recorded as an event, did you mean to use the ExpectEvent() assertion instead of ExpectCommand()?\n",
		)
	}

	return false
}
