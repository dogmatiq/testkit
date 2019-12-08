package envelope

import "github.com/dogmatiq/configkit"

// Origin describes the hanlder that produced a message in an envelope.
type Origin struct {
	// HandlerName is the name of the handler that produced this message.
	HandlerName string

	// HandlerType is the type of the handler that produced this message.
	HandlerType configkit.HandlerType

	// InstanceID is the ID of the aggregate or process instance that
	// produced this message.
	//
	// It is empty if HandlerType is neither configkit.AggregateHandlerType nor
	// configkit.ProcessHandlerType.
	InstanceID string
}
