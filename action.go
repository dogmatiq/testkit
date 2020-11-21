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
	// ExpectOptions returns the options to use by default when this action is
	// used with Test.Expect().
	ExpectOptions() []ExpectOption

	// Apply performs the action within the context of a specific test.
	Apply(ctx context.Context, s ActionScope) error
}

// ActionScope encapsulates the state that an action can inspect and manipulate.
type ActionScope struct {
	App              configkit.RichApplication
	TestingT         TestingT
	Engine           *engine.Engine
	Executor         *engine.CommandExecutor
	Recorder         *engine.EventRecorder
	VirtualClock     *time.Time
	OperationOptions []engine.OperationOption
}
