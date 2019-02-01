package assert

import (
	"github.com/dogmatiq/dogmatest/engine/fact"
	"github.com/dogmatiq/enginekit/handler"
	"github.com/dogmatiq/enginekit/message"
)

// tracker is an observer used by assertions that keeps track of common
// information about handlers and the messages they produce.
type tracker struct {
	// role is the role that the message is expecting to find.
	role message.Role

	// total is the total number of messages that were produced.
	total int

	// produced is the number of messages of the expected role that were produced.
	produced int

	// engaged is the set of handlers that *could* have produced the expected
	// message.
	engaged map[string]handler.Type

	// enabled is the set of handler types that are enabled during the test.
	enabled map[handler.Type]bool
}

// Notify updates the assertion's state in response to a new fact.
func (t *tracker) Notify(f fact.Fact) {
	switch x := f.(type) {
	case fact.DispatchCycleBegun:
		t.enabled = x.EnabledHandlers
	case fact.TickCycleBegun:
		t.enabled = x.EnabledHandlers
	case fact.HandlingBegun:
		t.updateEngaged(x.HandlerName, x.HandlerType)
	case fact.EventRecordedByAggregate:
		t.messageProduced(x.EventEnvelope.Role)
	case fact.EventRecordedByIntegration:
		t.messageProduced(x.EventEnvelope.Role)
	case fact.CommandExecutedByProcess:
		t.messageProduced(x.CommandEnvelope.Role)
	}
}

func (t *tracker) updateEngaged(n string, ht handler.Type) {
	if ht.IsProducerOf(t.role) {
		if t.engaged == nil {
			t.engaged = map[string]handler.Type{}
		}

		t.engaged[n] = ht
	}
}

func (t *tracker) messageProduced(r message.Role) {
	t.total++

	if r == t.role {
		t.produced++
	}
}
