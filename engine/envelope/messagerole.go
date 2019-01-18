package envelope

// MessageRole is an enumeration of the roles a message can perform within the
// engine.
type MessageRole string

const (
	// CommandRole is the role for command messages.
	CommandRole MessageRole = "command"

	// EventRole is the class for event messages.
	EventRole MessageRole = "event"
)

// MustValidate panics if r is not a valid message role.
func (r MessageRole) MustValidate() {
	switch r {
	case CommandRole:
	case EventRole:
	default:
		panic("invalid role: " + r)
	}
}
