package assert

import (
	"fmt"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/dogmatest/compare"
	"github.com/dogmatiq/dogmatest/render"
	"github.com/dogmatiq/enginekit/handler"
	"github.com/dogmatiq/enginekit/message"
)

// CommandExecuted asserts that a specific command is executed.
type CommandExecuted struct {
	messageAssertionBehavior

	Expected dogma.Message
}

// Begin is called before the test is executed.
//
// c is the comparator used to compare messages and other entities.
func (a *CommandExecuted) Begin(c compare.Comparator) {
	// reset everything
	a.messageAssertionBehavior = messageAssertionBehavior{
		expected: a.Expected,
		role:     message.CommandRole,
		cmp:      c,
	}
}

// End is called after the test is executed.
//
// It returns the result of the assertion.
func (a *CommandExecuted) End(r render.Renderer) *Result {
	res := &Result{
		Ok: a.ok,
		Criteria: fmt.Sprintf(
			"edxecuted a specific '%s' command",
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
func (a *CommandExecuted) buildResultNoMatch(r render.Renderer, res *Result) {
	s := res.Section(suggestionsSection)

	if !a.enabledHandlers[handler.ProcessType] {
		res.Explanation = "the relevant handler type (process) was not enabled"
		s.AppendListItem("enable process handlers using the EnableHandlerType() option")
		return
	}

	if len(a.engagedHandlers) == 0 {
		res.Explanation = "no relevant handlers (processes) were engaged"
		s.AppendListItem("check the application's routing configuration")
		return
	}

	if a.commands == 0 && a.events == 0 {
		res.Explanation = "no messages were produced at all"
	} else if a.commands == 0 {
		res.Explanation = "no commands were executed at all"
	} else {
		res.Explanation = "none of the engaged handlers executed the expected command"
	}

	for n, t := range a.engagedHandlers {
		s.AppendListItem("verify the logic within the '%s' %s message handler", n, t)
	}
}

// buildResultNoMatch builds the assertion result when there is a "best-match"
// message available.
func (a *CommandExecuted) buildResult(r render.Renderer, res *Result) {
	s := res.Section(suggestionsSection)

	// the "best match" is equal to the expected message. this means that only the
	// roles were mismatched.
	if a.equal {
		res.Explanation = fmt.Sprintf(
			"the expected message was recorded as an event by the '%s' %s message handler",
			a.best.Origin.HandlerName,
			a.best.Origin.HandlerType,
		)

		s.AppendListItem(
			"verify that the '%s' %s message handler intended to record an event of this type",
			a.best.Origin.HandlerName,
			a.best.Origin.HandlerType,
		)

		s.AppendListItem("verify that CommandExecuted is the correct assertion, did you mean EventRecorded?")
		return
	}

	if a.sim == compare.SameTypes {
		res.Explanation = fmt.Sprintf(
			"a similar command was executed by the '%s' %s message handler",
			a.best.Origin.HandlerName,
			a.best.Origin.HandlerType,
		)
		s.AppendListItem("check the content of the message")
	} else {
		res.Explanation = fmt.Sprintf(
			"a command of a similar type was executed by the '%s' %s message handler",
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
