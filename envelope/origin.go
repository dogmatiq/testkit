package envelope

import (
	"github.com/dogmatiq/enginekit/config"
)

// Origin describes the handler that produced a message in an envelope.
type Origin struct {
	// Handler is the handler that produced this message.
	Handler config.Handler

	// HandlerType is the type of the handler that produced this message.
	HandlerType config.HandlerType

	// InstanceID is the ID of the aggregate or process instance that produced
	// this message.
	//
	// It is empty if HandlerType is neither [config.AggregateHandlerType] nor
	// [config.ProcessHandlerType].
	InstanceID string
}
