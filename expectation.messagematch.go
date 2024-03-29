package testkit

import (
	"errors"

	"github.com/dogmatiq/configkit/message"
	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/testkit/envelope"
	"github.com/dogmatiq/testkit/fact"
	"github.com/dogmatiq/testkit/internal/inflect"
	"github.com/dogmatiq/testkit/location"
)

// ToExecuteCommandMatching returns an expectation that passes if a command is
// executed that satisfies the given predicate function.
//
// Always prefer using ToExecuteCommand() instead, if possible, as it provides
// more meaningful information in the result of a failure.
//
// pred is the predicate function. It is called for each executed command. It
// must return nil at least once for the expectation to pass.
//
// pred may return the IgnoreMessage error to indicate that the predicate does
// not apply to a specific message.
func ToExecuteCommandMatching(
	pred func(dogma.Message) error,
) Expectation {
	if pred == nil {
		panic("ToExecuteCommandMatching(<nil>): function must not be nil")
	}

	return &messageMatchExpectation{
		pred:         pred,
		expectedRole: message.CommandRole,
		exhaustive:   false,
	}
}

// ToOnlyExecuteCommandsMatching returns an expectation that passes if all
// executed commands satisfy the given predicate function.
//
// The expectation does NOT fail if other unrelated actions are performed, such
// as recording an event.
//
// pred is the predicate function. It is called for each executed command. It
// must return nil (or IgnoreMessage) every time for the expectation to pass.
//
// pred may return the IgnoreMessage error to indicate that the predicate does
// not apply to a specific message.
func ToOnlyExecuteCommandsMatching(
	pred func(dogma.Message) error,
) Expectation {
	if pred == nil {
		panic("ToOnlyExecuteCommandsMatching(<nil>): function must not be nil")
	}

	return &messageMatchExpectation{
		pred:         pred,
		expectedRole: message.CommandRole,
		exhaustive:   true,
	}
}

// ToRecordEventMatching returns an expectation that passes if an event is
// recorded that satisfies the given predicate function.
//
// Always prefer using ToRecordEvent() instead, if possible, as it provides
// more meaningful information in the result of a failure.
//
// pred is the predicate function. It is called for each recorded event. It
// must return nil at least once for the expectation to pass.
//
// pred may return the IgnoreMessage error to indicate that the predicate does
// not apply to a specific message.
func ToRecordEventMatching(
	pred func(dogma.Message) error,
) Expectation {
	if pred == nil {
		panic("ToRecordEventMatching(<nil>): function must not be nil")
	}

	return &messageMatchExpectation{
		pred:         pred,
		expectedRole: message.EventRole,
		exhaustive:   false,
	}
}

// ToOnlyRecordEventsMatching returns an expectation that passes if all
// recorded events satisfy the given predicate function.
//
// The expectation does NOT fail if other unrelated actions are performed, such
// as executing a command.
//
// pred is the predicate function. It is called for each recorded event. It
// must return nil (or IgnoreMessage) every time for the expectation to pass.
//
// pred may return the IgnoreMessage error to indicate that the predicate does
// not apply to a specific message.
func ToOnlyRecordEventsMatching(
	pred func(dogma.Message) error,
) Expectation {
	if pred == nil {
		panic("ToOnlyRecordEventsMatching(<nil>): function must not be nil")
	}

	return &messageMatchExpectation{
		pred:         pred,
		expectedRole: message.EventRole,
		exhaustive:   true,
	}
}

// IgnoreMessage is an error that can be returned by predicate functions to
// indicate that the predicate does not care about the message and therefore the
// predicate's result should not affect the expectation's result.
var IgnoreMessage = errors.New("this message does not need to be inspected by the predicate") //revive:disable-line:error-naming

// messageMatchExpectation is an Expectation that checks that at least one
// message that satisfies a predicate function is produced.
//
// It is the implementation used by ToExecuteCommandMatching() and
// ToRecordEventMatching().
type messageMatchExpectation struct {
	pred         func(dogma.Message) error
	expectedRole message.Role
	exhaustive   bool
}

func (e *messageMatchExpectation) Caption() string {
	if e.exhaustive {
		return inflect.Sprintf(
			e.expectedRole,
			"to only <produce> <messages> that match the predicate near %s",
			location.OfFunc(e.pred),
		)
	}

	return inflect.Sprintf(
		e.expectedRole,
		"to <produce> a <message> that matches the predicate near %s",
		location.OfFunc(e.pred),
	)
}

func (e *messageMatchExpectation) Predicate(s PredicateScope) (Predicate, error) {
	return &messageMatchPredicate{
		pred:         e.pred,
		expectedRole: e.expectedRole,
		exhaustive:   e.exhaustive,
		tracker: tracker{
			role:    e.expectedRole,
			options: s.Options,
		},
	}, nil
}

// messageMatchPredicate is the Predicate implementation for
// messageMatchExpectation.
type messageMatchPredicate struct {
	pred         func(dogma.Message) error
	expectedRole message.Role
	exhaustive   bool
	failures     []*failedMatch
	matched      int
	ignored      int
	ok           bool
	tracker      tracker
}

// Notify updates the expectation's state in response to a new fact.
func (p *messageMatchPredicate) Notify(f fact.Fact) {
	if p.ok {
		return
	}

	if env, ok := p.tracker.Notify(f); ok {
		p.messageProduced(env)
	}
}

// messageProduced updates the predicate's state to reflect the fact that a
// message of the expected role has been produced.
func (p *messageMatchPredicate) messageProduced(env *envelope.Envelope) {
	if env.Role != p.expectedRole {
		return
	}

	err := p.pred(env.Message)

	if err == nil {
		p.matched++

		if !p.exhaustive {
			// We're only looking for "at least one match", and we've found it.
			// Mark the predicate as satisfied so that we don't bother looking
			// for future matches.
			p.ok = true
			p.failures = nil
		}

		return
	}

	if err == IgnoreMessage {
		p.ignored++
		return
	}

	for _, f := range p.failures {
		if f.MessageType == env.Type && f.Error == err.Error() {
			f.Count++
			return
		}
	}

	p.failures = append(
		p.failures,
		&failedMatch{
			MessageType: env.Type,
			Error:       err.Error(),
			Count:       1,
		},
	)
}

func (p *messageMatchPredicate) Ok() bool {
	return p.ok
}

func (p *messageMatchPredicate) Done() {
	if p.exhaustive && len(p.failures) == 0 {
		p.ok = true
	}
}

func (p *messageMatchPredicate) Report(treeOk bool) *Report {
	rep := &Report{
		TreeOk: treeOk,
		Ok:     p.ok,
		Criteria: inflect.Sprintf(
			p.expectedRole,
			"<produce> a <message> that matches the predicate near %s",
			location.OfFunc(p.pred),
		),
	}

	if p.exhaustive {
		rep.Criteria = inflect.Sprintf(
			p.expectedRole,
			"only <produce> <messages> that match the predicate near %s",
			location.OfFunc(p.pred),
		)
	}

	if treeOk || p.ok {
		return rep
	}

	if len(p.failures) > 0 {
		s := rep.Section(failedMatchesSection)

		for _, f := range p.failures {
			if f.Count > 1 {
				s.AppendListItem(
					"%s: %s (repeated %d times)",
					f.MessageType,
					f.Error,
					f.Count,
				)
			} else {
				s.AppendListItem(
					"%s: %s",
					f.MessageType,
					f.Error,
				)
			}
		}
	}

	suggestions := rep.Section(suggestionsSection)

	if p.ignored > 0 {
		suggestions.AppendListItem(
			"verify the logic within the predicate function, it ignored %s",
			inflect.Sprintf(p.expectedRole, "%d <messages>", p.ignored),
		)
	} else if len(p.failures) > 0 {
		suggestions.AppendListItem("verify the logic within the predicate function")
	}

	reportNoMatch(rep, &p.tracker)

	if p.exhaustive {
		rep.Explanation = inflect.Sprintf(
			p.expectedRole,
			"only %d of %d relevant <messages> matched the predicate",
			p.matched,
			p.tracker.produced-p.ignored,
		)
	}

	return rep
}

type failedMatch struct {
	MessageType message.Type
	Error       string
	Count       int
}
