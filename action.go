package testkit

import (
	"context"
	"time"

	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/testkit/engine"
)

// Action is an interface for any action that can be performed within a test.
//
// Actions always attempt to cause some state change within the engine or
// application.
type Action interface {
	// Banner returns a human-readable banner to display in the logs when this
	// action is applied.
	//
	// The banner text should be in uppercase, and worded in the present tense,
	// for example "DOING ACTION".
	Banner() string

	// ConfigurePredicate updates o with any options required by the action.
	//
	// It is called before Apply() when the action is used with Test.Expect().
	ConfigurePredicate(o *PredicateOptions)

	// Apply performs the action within the context of a specific test.
	Apply(ctx context.Context, s ActionScope) error
}

// ActionScope encapsulates the state that an action can inspect and manipulate.
type ActionScope struct {
	App              configkit.RichApplication
	VirtualClock     *time.Time
	Engine           *engine.Engine
	Executor         *engine.CommandExecutor
	Recorder         *engine.EventRecorder
	OperationOptions []engine.OperationOption
}
