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
		panic("invalid type: " + t)
	}
}

// MustBe panics if t is not one of the given types.
func (t Type) MustBe(types ...Type) {
	t.MustValidate()

	for _, x := range types {
		x.MustValidate()

		if t == x {
			return
		}
	}

	panic("unexpected type: " + t)
}

// MustNotBe panics if t is one of the given types.
func (t Type) MustNotBe(types ...Type) {
	t.MustValidate()

	for _, x := range types {
		x.MustValidate()

		if t == x {
			panic("unexpected type: " + t)
		}
	}
}
