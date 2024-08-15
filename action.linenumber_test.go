package testkit_test

// This file contains definitions used within tests that check for specific line
// numbers. To minimize test disruption edit this file as infrequently as
// possible.
//
// New additions should always be made at the end so that the line numbers of
// existing definitions do not change. The padding below can be removed as
// imports statements added.

import (
	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/testkit"
	// import padding
	// import padding
	// import padding
	// import padding
	// import padding
	// import padding
	// import padding
	// import padding
	// import padding
	// import padding
	// import padding
	// import padding
	// import padding
	// import padding
	// import padding
	// import padding
	// import padding
	// import padding
	// import padding
	// import padding
	// import padding
	// import padding
	// import padding
	// import padding
	// import padding
	// import padding
	// import padding
	// import padding
	// import padding
	// import padding
	// import padding
	// import padding
	// import padding
	// import padding
)

func advanceTime(adj TimeAdjustment) Action { return AdvanceTime(adj) }
func call(fn func()) Action                 { return Call(fn) }
func executeCommand(m dogma.Command) Action { return ExecuteCommand(m) }
func recordEvent(m dogma.Event) Action      { return RecordEvent(m) }
