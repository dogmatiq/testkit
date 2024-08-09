package testkit

import (
	"context"
	"fmt"
	"slices"

	"github.com/dogmatiq/configkit/message"
	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/testkit/location"
)

type ScenarioAction struct {
	caption string
	loc     location.Location
	actions []Action
}

func Scenario(caption string) *ScenarioAction {
	return &ScenarioAction{
		caption: caption,
		loc:     location.OfCall(),
	}
}

// Caption returns the caption that should be used for this action in the
// test report.
func (s *ScenarioAction) Caption() string {
	return s.caption
}

// Location returns the location within the code that the action was
// constructed.
func (s *ScenarioAction) Location() location.Location {
	return s.loc
}

// ConfigurePredicate updates o with any options required by the action.
//
// It is called before Do() when the action is used with Test.Expect().
func (s *ScenarioAction) ConfigurePredicate(o *PredicateOptions) {
	for _, a := range s.actions {
		a.ConfigurePredicate(o)
	}
}

// Do performs the action within the context of a specific test.
func (s *ScenarioAction) Do(ctx context.Context, as ActionScope) error {
	for _, a := range s.actions {
		if err := a.Do(ctx, as); err != nil {
			return err
		}
	}

	return nil
}

// ExecuteCommand returns an Action that executes a command message.
func (s *ScenarioAction) ExecuteCommand(m dogma.Message) *ScenarioAction {
	if err := validateMessage(m); err != nil {
		panic(fmt.Sprintf("ExecuteCommand(%T): %s", m, err))
	}

	return &ScenarioAction{
		caption: s.caption,
		loc:     s.loc,
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
func (s *ScenarioAction) RecordEvent(m dogma.Message) *ScenarioAction {
	if err := validateMessage(m); err != nil {
		panic(fmt.Sprintf("RecordEvent(%T): %s", m, err))
	}

	return &ScenarioAction{
		caption: s.caption,
		loc:     s.loc,
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
func (s *ScenarioAction) AdvanceTime(adj TimeAdjustment) *ScenarioAction {
	if adj == nil {
		panic("AdvanceTime(<nil>): adjustment must not be nil")
	}

	return &ScenarioAction{
		caption: s.caption,
		loc:     s.loc,
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
func (s *ScenarioAction) Call(fn func(), options ...CallOption) *ScenarioAction {
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

	return &ScenarioAction{
		caption: s.caption,
		loc:     s.loc,
		actions: append(
			slices.Clone(s.actions),
			act,
		),
	}
}

// whenThereAreOpenAccounts := testkit.
// Scenario().
// ExecuteCommand(commands.OpenAccount{
// 	CustomerID:  annaCustomerID,
// 	AccountID:   annaAccountID,
// 	AccountName: "Anna Smith",
// }).
// ExecuteCommand(commands.OpenAccount{
// 	CustomerID:  bobCustomerID,
// 	AccountID:   bobAccountID,
// 	AccountName: "Bob Jones",
// })

// t.Run(
// "when there are sufficient funds",
// func(t *testing.T) {
// 	whenOneOfTheAccountsIsFunded := whenThereAreOpenAccounts.
// 		ExecuteCommand(commands.Deposit{
// 			TransactionID: "D001",
// 			AccountID:     annaAccountID,
// 			Amount:        500,
// 		})

// // ExecuteCommand returns an Action that executes a command message.
// func ExecuteCommand(m dogma.Message) Action {
// 	if err := validateMessage(m); err != nil {
// 		panic(fmt.Sprintf("ExecuteCommand(%T): %s", m, err))
// 	}

// 	return dispatchAction{
// 		message.CommandRole,
// 		m,
// 		location.OfCall(),
// 	}
// }

// // RecordEvent returns an Action that records an event message.
// func RecordEvent(m dogma.Message) Action {
// 	if err := validateMessage(m); err != nil {
// 		panic(fmt.Sprintf("RecordEvent(%T): %s", m, err))
// 	}

// 	return dispatchAction{
// 		message.EventRole,
// 		m,
// 		location.OfCall(),
// 	}
// }
