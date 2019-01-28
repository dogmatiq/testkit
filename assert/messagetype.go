package assert

import (
	"fmt"
	"io"
	"reflect"

	"github.com/dogmatiq/iago"
	"github.com/dogmatiq/iago/indent"

	"github.com/dogmatiq/dogmatest/compare"
	"github.com/dogmatiq/dogmatest/engine/envelope"
	"github.com/dogmatiq/dogmatest/engine/fact"
	"github.com/dogmatiq/dogmatest/render"
	"github.com/dogmatiq/enginekit/message"
)

// MessageTypeAssertion is an assertion that requires an exact message type to
// be produced.
type MessageTypeAssertion struct {
	Type message.Type
	Role message.Role

	cmp compare.Comparator

	pass bool
	best *envelope.Envelope
	sim  compare.TypeSimilarity
}

// Notify notifies the assertion of the occurrence of a fact.
func (a *MessageTypeAssertion) Notify(f fact.Fact) {
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

func (a *MessageTypeAssertion) update(env *envelope.Envelope) {
	// check to see if this message is of a similar type to our expected message
	sim := compare.FuzzyTypeComparison(
		a.Type.ReflectType(),
		reflect.TypeOf(env.Message),
	)

	// look for an identical message type
	if sim == compare.SameTypes && env.Role == a.Role {
		// if it's the right message role, we've found our match
		a.pass = true
		return
	}

	if sim > a.sim {
		a.best = env
		a.sim = sim
	}
}

// Begin is called before the message-under-test is dispatched.
func (a *MessageTypeAssertion) Begin(c compare.Comparator) {
	a.Role.MustBe(message.CommandRole, message.EventRole)

	a.cmp = c
	a.pass = false
	a.best = nil
	a.sim = 0
}

// End is called after the message-under-test is dispatched.
func (a *MessageTypeAssertion) End(w io.Writer, r render.Renderer) bool {
	writeIcon(w, a.pass)

	// write a description of the assertion
	writeByRole(
		w,
		a.Role,
		fmt.Sprintf(" execute any '%s' command", a.Type),
		fmt.Sprintf(" record any '%s' event", a.Type),
	)

	// we found the exact message type we expected
	if a.pass {
		iago.MustWriteString(w, "\n")
		return true
	}

	// write the failure message
	writeByRole(
		w,
		a.Role,
		" (no commands of this type were executed)\n\n",
		" (no events of this type were recorded)\n\n",
	)

	// if there's no "best match", write a description of the expected message
	// type then bail
	iw := indent.NewIndenter(w, []byte("  | "))
	if a.best == nil {
		iago.MustWriteString(iw, a.Type.String())
		iago.MustWriteString(w, "\n")
		return false
	}

	iago.Must(
		render.WriteDiff(
			w,
			a.Type.String(),
			a.best.Type.String(),
		),
	)
	iago.MustWriteString(iw, "\n\n")

	// write a hint about how the failure might be fixed
	if a.sim == compare.SameTypes {
		writeHintByRole(
			iw,
			a.Role,
			a.best.Role,
			"",
			"A message of this type was executed as a command, did you mean to use the ExpectCommandType() assertion instead of ExpectEventType()?",
			"A message of this type was recorded as an event, did you mean to use the ExpectEventType() assertion instead of ExpectCommandType()?",
		)
	} else {
		writeHintByRole(
			iw,
			a.Role,
			a.best.Role,
			"Check the type of the message.",
			"A message of a similar type was executed as a command, did you mean to use the ExpectCommandType() assertion instead of ExpectEventType()?",
			"A message of a similar type was recorded as an event, did you mean to use the ExpectEventType() assertion instead of ExpectCommandType()?",
		)
	}

	iago.MustWriteString(w, "\n")

	return false
}
