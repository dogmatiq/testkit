package envelope

import (
	"github.com/dogmatiq/dogmatest/internal/enginekit/handler"
)

// Origin describes the hanlder that produced a message in an envelope.
type Origin struct {
	// HandlerName is the name of the handler that produced this message.
	HandlerName string

	// HandlerType is the type of the handler that produced this message.
	HandlerType handler.Type

	// InstanceID is the ID of the aggregate or process instance that
	// produced this message.
	//
	// It is empty if HandlerType is neither handler.AggregateType nor
	// handler.ProcessType.
	InstanceID string
}
