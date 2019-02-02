package fact

import (
	"time"

	"github.com/dogmatiq/enginekit/handler"
)

// TickCycleBegun indicates that Engine.Tick() has been called.
type TickCycleBegun struct {
	EngineTime      time.Time
	EnabledHandlers map[handler.Type]bool
}

// TickCycleCompleted indicates that a call Engine.Tick() has completed.
type TickCycleCompleted struct {
	Error           error
	EnabledHandlers map[handler.Type]bool
}

// TickBegun indicates that a call to Controller.Tick() is being made.
type TickBegun struct {
	HandlerName string
	HandlerType handler.Type
}

// TickCompleted indicates that a call to Controller.Tick() has
// completed.
type TickCompleted struct {
	HandlerName string
	HandlerType handler.Type
	Error       error
}
