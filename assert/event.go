package assert

import (
	"fmt"
	"reflect"

	"github.com/dogmatiq/enginekit/handler"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/dogmatest/compare"
	"github.com/dogmatiq/dogmatest/engine/envelope"
	"github.com/dogmatiq/dogmatest/engine/fact"
	"github.com/dogmatiq/dogmatest/render"
	"github.com/dogmatiq/enginekit/message"
)

// EventRecorded asserts that a specific event is recorded.
type EventRecorded struct {
	// Expected is the event that is expected to be recorded.
	Expected dogma.Message

	// cmp is the comparator used to compare messages for equality.
	cmp compare.Comparator

	// ok is true once the assertion is deemed to have passed, after which no
	// further updates are performed.
	ok bool

	// best is an envelope containing the "best-match" message for an assertion
	// that has not yet passed. Note that this is not necessarily an event.
	best *envelope.Envelope

	// sim is a ranking of the similarity between the type of the expected event,
	// and the current best-match.
	sim compare.TypeSimilarity

	// equal is true if the best-match message compared as equal to the expected message.
	// this can occur, and the assertion still fail, if the best match is a command.
	equal bool

	// events is the total number of events that were recorded.
	events int

	// commands is the total number of commands that were executed.
	commands int

	// engagedHandlers is the set of handlers that *could* have recorded the
	// expected event.
	engagedHandlers map[string]handler.Type

	// enabledHandlers is the set of handler types that are enabled during the
	// test.
	enabledHandlers map[handler.Type]bool
}

// Notify updates the assertion's state in response to a new fact.
func (a *EventRecorded) Notify(f fact.Fact) {
	if a.ok {
		return
	}

	switch x := f.(type) {
	case fact.MessageDispatchBegun:
		a.enabledHandlers = x.EnabledHandlers
	case fact.EngineTickBegun:
		a.enabledHandlers = x.EnabledHandlers
	case fact.MessageHandlingBegun:
		a.messageHandlingBegun(x)
	case fact.EventRecordedByAggregate:
		a.eventRecorded(x.EventEnvelope)
	case fact.EventRecordedByIntegration:
		a.eventRecorded(x.EventEnvelope)
	case fact.CommandExecutedByProcess:
		a.commandExecuted(x.CommandEnvelope)
	}
}

// messageHandlingBegun updates the assertion's state to reflect f.
func (a *EventRecorded) messageHandlingBegun(f fact.MessageHandlingBegun) {
	// only aggregates and integration handlers can record events, so we're not
	// interested in anything else
	if !f.HandlerType.Is(handler.AggregateType, handler.IntegrationType) {
		return
	}

	if a.engagedHandlers == nil {
		a.engagedHandlers = map[string]handler.Type{}
	}

	a.engagedHandlers[f.HandlerName] = f.HandlerType
}

// commandExecuted updates the assertion's state to reflect the fact that a
// command was executed.
func (a *EventRecorded) commandExecuted(env *envelope.Envelope) {
	a.commands++

	if a.cmp.MessageIsEqual(env.Message, a.Expected) {
		// an equal COMMAND is the best match we'll ever get that doesn't actually
		// result in a passing assertion.
		a.best = env
		a.sim = compare.SameTypes
		a.equal = true
		return
	}

	a.updateBestMatch(env)
}

// eventRecorded updates the assertion's state to reflect the fact that an event
// was recorded.
func (a *EventRecorded) eventRecorded(env *envelope.Envelope) {
	a.events++

	if a.cmp.MessageIsEqual(env.Message, a.Expected) {
		a.ok = true
		a.best = nil
		a.sim = compare.UnrelatedTypes
		a.equal = true
		return
	}

	a.updateBestMatch(env)
}

// updateBestMatch replaces a.best with env if it is a better match.
func (a *EventRecorded) updateBestMatch(env *envelope.Envelope) {
	sim := compare.FuzzyTypeComparison(
		reflect.TypeOf(a.Expected),
		reflect.TypeOf(env.Message),
	)

	if sim > a.sim {
		a.best = env
		a.sim = sim
	}
}

// Begin is called before the test is executed.
//
// c is the comparator used to compare messages and other entities.
func (a *EventRecorded) Begin(c compare.Comparator) {
	// reset everything
	*a = EventRecorded{
		Expected: a.Expected,
		cmp:      c,
	}
}

// End is called after the test is executed.
//
// It returns the result of the assertion.
func (a *EventRecorded) End(r render.Renderer) *Result {
	res := &Result{
		Ok: a.ok,
		Criteria: fmt.Sprintf(
			"record a specific '%s' event",
			message.TypeOf(a.Expected),
		),
	}

	if !a.ok {
		if a.best == nil {
			a.buildResultNoMatch(r, res)
		} else {
			a.buildResult(r, res)
		}
	}

	return res
}

// buildResultNoMatch builds the assertion result when there is no "best-match"
// message.
func (a *EventRecorded) buildResultNoMatch(r render.Renderer, res *Result) {
	s := res.Section(suggestionsSection)

	if !a.enabledHandlers[handler.AggregateType] &&
		!a.enabledHandlers[handler.IntegrationType] {
		res.Explanation = "no relevant handler types (aggregate and integration) were enabled"
		s.AppendListItem("enable the relevant handler types using the EnableHandlerType() option")
		return
	}

	if !a.enabledHandlers[handler.AggregateType] {
		s.AppendListItem("enable aggregate handlers using the EnableHandlerType() option")
	}

	if !a.enabledHandlers[handler.IntegrationType] {
		s.AppendListItem("enable integration handlers using the EnableHandlerType() option")
	}

	if len(a.engagedHandlers) == 0 {
		res.Explanation = "no relevant handlers (aggregates or integrations) were engaged"
		s.AppendListItem("check the application's routing configuration")
		return
	}

	if a.commands == 0 && a.events == 0 {
		res.Explanation = "no messages were produced at all"
	} else if a.events == 0 {
		res.Explanation = "no events were recorded at all"
	} else {
		res.Explanation = "none of the engaged handlers recorded the expected event"
	}

	for n, t := range a.engagedHandlers {
		s.AppendListItem("verify the logic within the '%s' %s message handler", n, t)
	}
}

// buildResultNoMatch builds the assertion result when there is a "best-match"
// message available.
func (a *EventRecorded) buildResult(r render.Renderer, res *Result) {
	s := res.Section(suggestionsSection)

	// the "best match" is equal to the expected message. this means that only the
	// roles were mismatched.
	if a.equal {
		res.Explanation = fmt.Sprintf(
			"the expected message was executed as a command by the '%s' %s message handler",
			a.best.Origin.HandlerName,
			a.best.Origin.HandlerType,
		)

		s.AppendListItem(
			"verify that the '%s' %s message handler intended to execute a command of this type",
			a.best.Origin.HandlerName,
			a.best.Origin.HandlerType,
		)

		s.AppendListItem("verify that EventRecorded is the correct assertion, did you mean CommandExecuted?")
		return
	}

	if a.sim == compare.SameTypes {
		res.Explanation = fmt.Sprintf(
			"a similar event was recorded by the '%s' %s message handler",
			a.best.Origin.HandlerName,
			a.best.Origin.HandlerType,
		)
		s.AppendListItem("check the content of the message")
	} else {
		res.Explanation = fmt.Sprintf(
			"an event of a similar type was recorded by the '%s' %s message handler",
			a.best.Origin.HandlerName,
			a.best.Origin.HandlerType,
		)
		// note this language here is deliberately vague, it doesn't imply whether it
		// currently is or isn't a pointer, just questions if it should be.
		s.AppendListItem("check the message type, should it be a pointer?")
	}

	render.WriteDiff(
		&res.Section(diffSection).Content,
		render.Message(r, a.Expected),
		render.Message(r, a.best.Message),
	)
}
