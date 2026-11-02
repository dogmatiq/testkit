package envelope

import (
	"strconv"
	"sync/atomic"
)

// MessageIDGenerator produces sequential message IDs.
type MessageIDGenerator struct {
	messageID atomic.Uint64
}

// Next returns the next ID in the sequence.
func (g *MessageIDGenerator) Next() string {
	return strconv.FormatUint(
		g.messageID.Add(1),
		10,
	)
}

// Reset resets the generator to begin at one (1) again.
func (g *MessageIDGenerator) Reset() {
	g.messageID.Store(0)
}
