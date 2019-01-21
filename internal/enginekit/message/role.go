package message

// Role is an enumeration of the roles a message can perform within an engine.
type Role string

const (
	// CommandRole is the role for command messages.
	CommandRole Role = "command"

	// EventRole is the role for event messages.
	EventRole Role = "event"

	// TimeoutRole is the role for timeout messages.
	TimeoutRole Role = "timeout"
)

// MustValidate panics if r is not a valid message role.
func (r Role) MustValidate() {
	switch r {
	case CommandRole:
	case EventRole:
	case TimeoutRole:
	default:
		panic("invalid role")
	}
}
