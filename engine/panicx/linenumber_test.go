package panicx_test

// This file contains definitions used within tests that check for specific line
// numbers. To minimize test disruption edit this file as infrequently as
// possible.
//
// New additions should always be made at the end so that the line numbers of
// existing definitions do not change. The padding below can be removed as
// imports statements added.

import (
	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/testkit/engine/panicx"
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

func doNothing()                     {}
func panicWithUnexpectedMessage()    { panic(dogma.UnexpectedMessage) }
func locationOfCallLayer1() Location { return LocationOfCall() }
func locationOfCallLayer2() Location { return locationOfCallLayer1() }
