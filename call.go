package testkit

import "context"

// Call is an Action that invokes a user-defined function within the context of
// a test.
//
// It is intended to execute application code that makes use of the
// dogma.CommandExecutor or dogma.EventRecorder interfaces.
//
// Test implementations of these interfaces can be OBTAINED via the
// Test.CommandExecutor() and Test.EventRecorder() methods at any time; however,
// they may only be USED within a function invoked by a Call() action.
//
// When Call() is used with Test.Expect() the expectation will match the
// messages dispatched via the test's executor and recorder, as well as those
// produced by handlers within the Dogma application.
func Call(fn func()) Action {
	if fn == nil {
		panic("Call(): function must not be nil")
	}

	return call{fn}
}

// call is an implementation of Action that invokes a user-defined function.
type call struct {
	fn func()
}

// ExpectOptions returns the options to use by default when this action is
// used with Test.Expect().
func (a call) ExpectOptions() []ExpectOption {
	return []ExpectOption{
		func(o *ExpectOptionSet) {
			o.MatchMessagesInDispatchCycle = true
		},
	}
}

// Apply performs the action within the context of a specific test.
func (a call) Apply(ctx context.Context, s ActionScope) error {
	s.Test.executor.Engine = s.Engine
	s.Test.recorder.Engine = s.Engine
	s.Test.executor.Options = s.OperationOptions
	s.Test.recorder.Options = s.OperationOptions

	defer func() {
		// Reset the engine and options to nil so that the executor and recorder
		// can not be used after this Call() action ends.
		s.Test.executor.Engine = nil
		s.Test.recorder.Engine = nil
		s.Test.executor.Options = nil
		s.Test.recorder.Options = nil
	}()

	log(
		s.TestingT,
		"--- CALLING USER-DEFINED FUNCTION ---",
	)

	a.fn()

	return nil
}
