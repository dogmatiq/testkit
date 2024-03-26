package testkit

import (
	"context"
	"time"

	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/testkit/engine"
	"github.com/dogmatiq/testkit/location"
)

// Action is an interface for any action that can be performed within a test.
//
// Actions always attempt to cause some state change within the engine or
// application.
type Action interface {
	// Caption returns the caption that should be used for this action in the
	// test report.
	Caption() string

	// Location returns the location within the code that the action was
	// constructed.
	Location() location.Location

	// ConfigurePredicate updates o with any options required by the action.
	//
	// It is called before Do() when the action is used with Test.Expect().
	ConfigurePredicate(o *PredicateOptions)

	// Do performs the action within the context of a specific test.
	Do(ctx context.Context, s ActionScope) error
}

// ActionScope encapsulates the element's of a Test's state that may be
// inspected and manipulated by Action implementations.
type ActionScope struct {
	// App is the application being tested.
	App configkit.RichApplication

	// VirtualClock is the time that the Test uses as the engine time for the
	// NEXT Action.
	VirtualClock *time.Time

	// Engine is the engine used to handle messages.
	Engine *engine.Engine

	// Executor is the command executor returned by the Test's CommandExecutor()
	// method.
	Executor *CommandExecutor

	// OperationOptions is the set of options that should be used with calling
	// Engine.Dispatch() or Engine.Tick().
	OperationOptions []engine.OperationOption
}
