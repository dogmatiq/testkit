package envelope

import (
	"strconv"
	"sync/atomic"
)

// MessageIDGenerator produces sequential message IDs.
type MessageIDGenerator struct {
	messageID uint64 // atomic
}

// Next returns the next ID in the sequence.
func (g *MessageIDGenerator) Next() string {
	return strconv.FormatUint(
		atomic.AddUint64(&g.messageID, 1),
		10,
	)
}

// Reset resets the generator to begin at one (1) again.
func (g *MessageIDGenerator) Reset() {
	atomic.StoreUint64(&g.messageID, 0)
}
