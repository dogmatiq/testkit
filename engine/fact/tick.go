package fact

import (
	"time"

	"github.com/dogmatiq/configkit"
)

// TickCycleBegun indicates that Engine.Tick() has been called.
type TickCycleBegun struct {
	EngineTime      time.Time
	EnabledHandlers map[configkit.HandlerType]bool
}

// TickCycleCompleted indicates that a call Engine.Tick() has completed.
type TickCycleCompleted struct {
	Error           error
	EnabledHandlers map[configkit.HandlerType]bool
}

// TickBegun indicates that a call to Controller.Tick() is being made.
type TickBegun struct {
	HandlerName string
	HandlerType configkit.HandlerType
}

// TickCompleted indicates that a call to Controller.Tick() has
// completed.
type TickCompleted struct {
	HandlerName string
	HandlerType configkit.HandlerType
	Error       error
}
