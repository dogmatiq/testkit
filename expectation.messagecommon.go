package testkit

import (
	"fmt"
	"strings"

	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/configkit/message"
	"github.com/dogmatiq/testkit/envelope"
	"github.com/dogmatiq/testkit/fact"
	"github.com/dogmatiq/testkit/internal/inflect"
)

// reportNoMatch is used by message-related predicates to build a test report
// when no "best-match" message is found.
func reportNoMatch(rep *Report, t *tracker) {
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

		if !t.options.MatchDispatchCycleStartedFacts {
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
	} else if t.options.MatchDispatchCycleStartedFacts {
		rep.Explanation = inflect.Sprint(t.role, "nothing <produced> a matching <message>")
	} else {
		rep.Explanation = inflect.Sprint(t.role, "none of the engaged handlers <produced> a matching <message>")
	}

	for _, n := range t.engagedOrder {
		s.AppendListItem("verify the logic within the '%s' %s message handler", n, t.engagedType[n])
	}

	if t.options.MatchDispatchCycleStartedFacts {
		s.AppendListItem(inflect.Sprint(t.role, "verify the logic within the code that uses the <dispatcher>"))
	}
}

// validateRole returns an error if the message type t does not have a role of r
// within the application.
func validateRole(
	s PredicateScope,
	t message.Type,
	r message.Role,
) error {
	actual, ok := s.App.MessageTypes().RoleOf(t)

	// TODO: These checks should result in information being added to the
	// report, not just returning an error.
	//
	// See https://github.com/dogmatiq/testkit/issues/162
	if !ok {
		return inflect.Errorf(
			r,
			"a <message> of type %s can never be <produced>, the application does not use this message type",
			t,
		)
	} else if actual != r {
		return inflect.Errorf(
			r,
			"%s is a %s, it can never be <produced> as a <message>",
			t,
			actual,
		)
	} else if !s.Options.MatchDispatchCycleStartedFacts {
		// If we're NOT matching messages from DispatchCycleStarted facts that
		// means this expectation can only ever pass if the message is produced
		// by a handler.
		if _, ok := s.App.MessageTypes().Produced[t]; !ok {
			return inflect.Errorf(
				r,
				"no handlers <produce> <messages> of type %s, it is only ever consumed",
				t,
			)
		}
	}

	return nil
}

// tracker is a fact.Observer used by expectations that need to keep track of
// information about handlers and the messages they produce.
type tracker struct {
	// role is the role that the message is expecting to find.
	role message.Role

	// options is the set of options passed to the predicate.
	options PredicateOptions

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

// Notify updates the tracker's state in response to a new fact.
//
// It returns the envelope containing the message that was tracked.
func (t *tracker) Notify(f fact.Fact) (*envelope.Envelope, bool) {
	switch x := f.(type) {
	case fact.DispatchCycleBegun:
		t.cycleBegun = true
		t.enabled = x.EnabledHandlerTypes
		if t.options.MatchDispatchCycleStartedFacts {
			t.messageProduced(x.Envelope.Role)
			return x.Envelope, true
		}
	case fact.HandlingBegun:
		t.updateEngaged(
			x.Handler.Identity().Name,
			x.Handler.HandlerType(),
		)
	case fact.EventRecordedByAggregate:
		t.messageProduced(x.EventEnvelope.Role)
		return x.EventEnvelope, true
	case fact.EventRecordedByIntegration:
		t.messageProduced(x.EventEnvelope.Role)
		return x.EventEnvelope, true
	case fact.CommandExecutedByProcess:
		t.messageProduced(x.CommandEnvelope.Role)
		return x.CommandEnvelope, true
	}

	return nil, false
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
