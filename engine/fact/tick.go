package fact

import (
	"time"

	"github.com/dogmatiq/configkit"
)

// TickCycleBegun indicates that Engine.Tick() has been called.
type TickCycleBegun struct {
	EngineTime          time.Time
	EnabledHandlerTypes map[configkit.HandlerType]bool
	EnabledHandlers     map[string]bool
}

// TickCycleCompleted indicates that a call Engine.Tick() has completed.
type TickCycleCompleted struct {
	Error               error
	EnabledHandlerTypes map[configkit.HandlerType]bool
	EnabledHandlers     map[string]bool
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

// TickSkipped indicates that a call to Controller.Tick() has been skipped,
// either because all handlers of that type are or the handler itself is
// disabled.
type TickSkipped struct {
	Handler configkit.RichHandler
	Reason  HandlerSkipReason
}
