package controller

import (
	"github.com/dogmatiq/dogmatest/engine/envelope"
	"github.com/dogmatiq/dogmatest/engine/fact"
)

// Scope holds context relevant to the handling of a specific
// message by a controller.
type Scope interface {
	Envelope() *envelope.Envelope
	RecordFacts(...fact.Fact)
}
