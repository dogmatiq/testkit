package fact

import (
	"time"

	"github.com/dogmatiq/enginekit/handler"
)

// EngineTickBegun indicates that Engine.Tick() has been called.
type EngineTickBegun struct {
	Now             time.Time
	EnabledHandlers map[handler.Type]bool
}

// EngineTickCompleted indicates that a call Engine.Tick() has completed.
type EngineTickCompleted struct {
	Now             time.Time
	Error           error
	EnabledHandlers map[handler.Type]bool
}

// ControllerTickBegun indicates that a call to Controller.Tick() is being made.
type ControllerTickBegun struct {
	HandlerName string
	HandlerType handler.Type
	Now         time.Time
}

// ControllerTickCompleted indicates that a call to Controller.Tick() has
// completed.
type ControllerTickCompleted struct {
	HandlerName string
	HandlerType handler.Type
	Now         time.Time
	Error       error
}
