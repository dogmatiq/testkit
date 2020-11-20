package assert

import (
	"fmt"
	"strings"

	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/configkit/message"
	"github.com/dogmatiq/testkit/engine/fact"
	"github.com/dogmatiq/testkit/internal/inflect"
)

// tracker is an observer used by assertions that keeps track of common
// information about handlers and the messages they produce.
type tracker struct {
	// role is the role that the message is expecting to find.
	role message.Role

	// matchDispatchCycle, if true, tracks messages that originate from a
	// command executor or event recorder, not just those that originate from
	// handlers within the application.
	matchDispatchCycle bool

	// cycleBegun is true if at least one dispatch or tick cycle was started.
	cycleBegun bool

	// total is the total number of messages that were produced.
	total int

	// produced is the number of messages of the expected role that were
	// produced.
	produced int

	// engagedOrder and engagedType track the set of handlers that *could* have
	// produced the expected message.
	engagedOrder []string
	engagedType  map[string]configkit.HandlerType

	// enabled is the set of handler types that are enabled during the test.
	enabled map[configkit.HandlerType]bool
}

// Notify updates the assertion's state in response to a new fact.
func (t *tracker) Notify(f fact.Fact) {
	switch x := f.(type) {
	case fact.DispatchCycleBegun:
		t.cycleBegun = true
		t.enabled = x.EnabledHandlers
		if t.matchDispatchCycle {
			t.messageProduced(x.Envelope.Role)
		}
	case fact.TickCycleBegun:
		t.cycleBegun = true
		t.enabled = x.EnabledHandlers
	case fact.HandlingBegun:
		t.updateEngaged(x.HandlerIdentity.Name, x.HandlerType)
	case fact.EventRecordedByAggregate:
		t.messageProduced(x.EventEnvelope.Role)
	case fact.EventRecordedByIntegration:
		t.messageProduced(x.EventEnvelope.Role)
	case fact.CommandExecutedByProcess:
		t.messageProduced(x.CommandEnvelope.Role)
	}
}

func (t *tracker) updateEngaged(n string, ht configkit.HandlerType) {
	if ht.IsProducerOf(t.role) {
		if t.engagedType == nil {
			t.engagedType = map[string]configkit.HandlerType{}
		}

		if _, ok := t.engagedType[n]; !ok {
			t.engagedOrder = append(t.engagedOrder, n)
			t.engagedType[n] = ht
		}
	}
}

func (t *tracker) messageProduced(r message.Role) {
	t.total++

	if r == t.role {
		t.produced++
	}
}

// buildResultNoMatch is a helper used by MessageAssertion and
// MessageTypeAssertion when there is no "best-match".
func buildResultNoMatch(rep *Report, t *tracker) {
	s := rep.Section(suggestionsSection)

	allDisabled := true
	var relevant []string

	if t.cycleBegun {
		for _, ht := range configkit.HandlerTypes {
			e := t.enabled[ht]

			if ht.IsProducerOf(t.role) {
				relevant = append(relevant, ht.String())

				if e {
					allDisabled = false
				} else {
					s.AppendListItem(
						fmt.Sprintf("enable %s handlers using the EnableHandlerType() option", ht),
					)
				}
			}
		}

		if !t.matchDispatchCycle {
			if allDisabled {
				rep.Explanation = "no relevant handler types were enabled"
				return
			}

			if len(t.engagedOrder) == 0 {
				rep.Explanation = fmt.Sprintf(
					"no relevant handlers (%s) were engaged",
					strings.Join(relevant, " or "),
				)
				s.AppendListItem("check the application's routing configuration")
				return
			}
		}
	}

	if t.total == 0 {
		rep.Explanation = "no messages were produced at all"
	} else if t.produced == 0 {
		rep.Explanation = inflect.Sprint(t.role, "no <messages> were <produced> at all")
	} else if t.matchDispatchCycle {
		rep.Explanation = inflect.Sprint(t.role, "nothing <produced> the expected <message>")
	} else {
		rep.Explanation = inflect.Sprint(t.role, "none of the engaged handlers <produced> the expected <message>")
	}

	for _, n := range t.engagedOrder {
		s.AppendListItem("verify the logic within the '%s' %s message handler", n, t.engagedType[n])
	}

	if t.matchDispatchCycle {
		s.AppendListItem(inflect.Sprint(t.role, "verify the logic within the code that uses the <dispatcher>"))
	}
}
