package handler

// Type is an enumeration of the types of handlers.
type Type string

const (
	// AggregateType is the handler type for dogma.AggregateMessageHandler.
	AggregateType Type = "aggregate"

	// ProcessType is the handler type for dogma.ProcessMessageHandler.
	ProcessType Type = "process"

	// IntegrationType is the handler type for dogma.IntegrationMessageHandler.
	IntegrationType Type = "integration"

	// ProjectionType is the handler type for dogma.ProjectionMessageHandler.
	ProjectionType Type = "projection"
)

// MustValidate panics if r is not a valid message role.
func (t Type) MustValidate() {
	switch t {
	case AggregateType:
	case ProcessType:
	case IntegrationType:
	case ProjectionType:
	default:
		panic("invalid type")
	}
}
