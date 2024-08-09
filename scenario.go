package testkit

import (
	"fmt"
	"slices"

	"github.com/dogmatiq/configkit/message"
	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/testkit/location"
)

// TestScenario is a reusable sequence of actions that can be used to prepare
// multiple tests with the same initial state.
type TestScenario struct {
	captions []string
	actions  []Action
}

// Scenario starts a new scenario, which may be used to prepare multiple tests
// with the same initial state.
func Scenario(caption string) *TestScenario {
	return &TestScenario{
		captions: []string{caption},
	}
}

// Scenario creates a new "sub scenario" within s.
func (s *TestScenario) Scenario(caption string) *TestScenario {
	return &TestScenario{
		captions: append(slices.Clone(s.captions), caption),
		actions:  slices.Clone(s.actions),
	}
}

// ExecuteCommand returns an Action that executes a command message.
func (s *TestScenario) ExecuteCommand(m dogma.Message) *TestScenario {
	if err := validateMessage(m); err != nil {
		panic(fmt.Sprintf("ExecuteCommand(%T): %s", m, err))
	}

	return &TestScenario{
		captions: s.captions,
		actions: append(
			slices.Clone(s.actions),
			dispatchAction{
				message.CommandRole,
				m,
				location.OfCall(),
			},
		),
	}
}

// RecordEvent returns an Action that records an event message.
func (s *TestScenario) RecordEvent(m dogma.Message) *TestScenario {
	if err := validateMessage(m); err != nil {
		panic(fmt.Sprintf("RecordEvent(%T): %s", m, err))
	}

	return &TestScenario{
		captions: s.captions,
		actions: append(
			slices.Clone(s.actions),
			dispatchAction{
				message.EventRole,
				m,
				location.OfCall(),
			},
		),
	}
}

// AdvanceTime returns an Action that simulates the passage of time by advancing
// the test's virtual clock.
//
// This allows testing of application logic that depends on time, such as
// processes that use timeout messages and projections that use the "recorded
// at" time of events.
//
// It accepts a TimeAdjustment which calculates the amount of time that the
// clock is advanced.
//
// There are two built-in adjustment types; ToTime() and ByDuration(). Users may
// provide their own TimeAdjustment implementations that model time-related
// concepts within the application's business domain.
func (s *TestScenario) AdvanceTime(adj TimeAdjustment) *TestScenario {
	if adj == nil {
		panic("AdvanceTime(<nil>): adjustment must not be nil")
	}

	return &TestScenario{
		captions: s.captions,
		actions: append(
			slices.Clone(s.actions),
			advanceTimeAction{
				adj,
				location.OfCall(),
			},
		),
	}
}

// Call is an Action that invokes a user-defined function within the context of
// a test.
//
// It is intended to execute application code that makes use of the
// dogma.CommandExecutor or dogma.EventRecorder interfaces. Typically this
// occurs in API handlers, where the "outside world" begins to interface with
// the Dogma application.
//
// If a test does not need to involve such application code, use of the
// ExecuteCommand() and RecordEvent() actions is preferred.
//
// Test implementations of these interfaces can be OBTAINED via the
// Test.CommandExecutor() and Test.EventRecorder() methods at any time; however,
// they may only be USED within a function invoked by a Call() action.
//
// When Call() is used with Test.Expect() the expectation will match the
// messages dispatched via the test's executor and recorder, as well as those
// produced by handlers within the Dogma application.
func (s *TestScenario) Call(fn func(), options ...CallOption) *TestScenario {
	if fn == nil {
		panic("Call(<nil>): function must not be nil")
	}

	act := callAction{
		fn:  fn,
		loc: location.OfCall(),
	}

	for _, opt := range options {
		opt.applyCallOption(&act)
	}

	return &TestScenario{
		captions: s.captions,
		actions: append(
			slices.Clone(s.actions),
			act,
		),
	}
}

// Prepare performs a the scenario's actions within the given test.
func (s *TestScenario) Prepare(t *Test) {
	t.testingT.Helper()

	log(t.testingT, "=== SCENARIO ===")
	for _, caption := range s.captions {
		logf(t.testingT, " • %s", caption)
	}

	t.Prepare(s.actions...)
}
