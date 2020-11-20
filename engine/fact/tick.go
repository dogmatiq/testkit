package fact

import (
	"time"

	"github.com/dogmatiq/configkit"
)

// TickCycleBegun indicates that Engine.Tick() has been called.
type TickCycleBegun struct {
	EngineTime          time.Time
	EnabledHandlerTypes map[configkit.HandlerType]bool
}

// TickCycleCompleted indicates that a call Engine.Tick() has completed.
type TickCycleCompleted struct {
	Error               error
	EnabledHandlerTypes map[configkit.HandlerType]bool
}

// TickBegun indicates that a call to Controller.Tick() is being made.
type TickBegun struct {
	Handler configkit.RichHandler
}

// TickCompleted indicates that a call to Controller.Tick() has completed.
type TickCompleted struct {
	Handler configkit.RichHandler
	Error   error
}
