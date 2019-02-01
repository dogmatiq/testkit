package envelope

import "sync/atomic"

// MessageIDGenerator produces sequential message IDs.
type MessageIDGenerator struct {
	messageID uint64 // atomic
}

// Next returns the next ID in the sequence.
func (g *MessageIDGenerator) Next() uint64 {
	return atomic.AddUint64(&g.messageID, 1)
}

// Reset resets the generator to begin at one (1) again.
func (g *MessageIDGenerator) Reset() {
	atomic.StoreUint64(&g.messageID, 0)
}
