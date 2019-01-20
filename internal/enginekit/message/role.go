package message

// Role is an enumeration of the roles a message can perform within an engine.
type Role int

const (
	// CommandRole is the role for command messages.
	CommandRole Role = iota

	// EventRole is the class for event messages.
	EventRole

	// TimeoutRole is the class for timeout messages.
	TimeoutRole
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
