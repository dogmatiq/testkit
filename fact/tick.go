package fact

import (
	"time"

	"github.com/dogmatiq/enginekit/config"
)

// TickCycleBegun indicates that Engine.Tick() has been called.
type TickCycleBegun struct {
	EngineTime          time.Time
	EnabledHandlerTypes map[config.HandlerType]bool
	EnabledHandlers     map[string]bool
}

// TickCycleCompleted indicates that a call Engine.Tick() has completed.
type TickCycleCompleted struct {
	Error               error
	EnabledHandlerTypes map[config.HandlerType]bool
	EnabledHandlers     map[string]bool
}

// TickBegun indicates that a call to Controller.Tick() is being made.
type TickBegun struct {
	Handler config.Handler
}

// TickCompleted indicates that a call to Controller.Tick() has completed.
type TickCompleted struct {
	Handler config.Handler
	Error   error
}

// TickSkipped indicates that a call to Controller.Tick() has been skipped.
type TickSkipped struct {
	Handler config.Handler
	Reason  HandlerSkipReason
}
