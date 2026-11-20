package fixtures

import (
	"encoding/json"

	"github.com/dogmatiq/dogma"
)

// ScopeCapturingCommand is a command that records the validation scope that it
// was most recently validated with.
type ScopeCapturingCommand struct {
	Scope dogma.CommandValidationScope `json:"-"`
}

// MessageDescription returns a description of the command.
func (*ScopeCapturingCommand) MessageDescription() string {
	return "command that captures its validation scope"
}

// Validate stores s in m.Scope.
func (m *ScopeCapturingCommand) Validate(s dogma.CommandValidationScope) error {
	m.Scope = s
	return nil
}

// ScopeCapturingEvent is an event that records the validation scope that it was
// most recently validated with.
type ScopeCapturingEvent struct {
	Scope dogma.EventValidationScope `json:"-"`
}

// MessageDescription returns a description of the event.
func (*ScopeCapturingEvent) MessageDescription() string {
	return "event that captures its validation scope"
}

// Validate stores s in m.Scope.
func (m *ScopeCapturingEvent) Validate(s dogma.EventValidationScope) error {
	m.Scope = s
	return nil
}

// ScopeCapturingDeadline is a deadline that records the validation scope that
// it was most recently validated with.
type ScopeCapturingDeadline struct {
	Scope dogma.DeadlineValidationScope `json:"-"`
}

// MessageDescription returns a description of the deadline.
func (*ScopeCapturingDeadline) MessageDescription() string {
	return "deadline that captures its validation scope"
}

// Validate stores s in m.Scope.
func (m *ScopeCapturingDeadline) Validate(s dogma.DeadlineValidationScope) error {
	m.Scope = s
	return nil
}

// MarshalBinary implements [encoding.BinaryMarshaler].
func (m *ScopeCapturingCommand) MarshalBinary() ([]byte, error) {
	return json.Marshal(m)
}

// UnmarshalBinary implements [encoding.BinaryUnmarshaler].
func (m *ScopeCapturingCommand) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, m)
}

// MarshalBinary implements [encoding.BinaryMarshaler].
func (m *ScopeCapturingEvent) MarshalBinary() ([]byte, error) {
	return json.Marshal(m)
}

// UnmarshalBinary implements [encoding.BinaryUnmarshaler].
func (m *ScopeCapturingEvent) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, m)
}

// MarshalBinary implements [encoding.BinaryMarshaler].
func (m *ScopeCapturingDeadline) MarshalBinary() ([]byte, error) {
	return json.Marshal(m)
}

// UnmarshalBinary implements [encoding.BinaryUnmarshaler].
func (m *ScopeCapturingDeadline) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, m)
}

func init() {
	dogma.RegisterCommand[*ScopeCapturingCommand]("a1f1e0a4-5b43-4c0e-9d0a-0c7f5d6f2b11")
	dogma.RegisterEvent[*ScopeCapturingEvent]("5c3d8e92-7a1b-4f6e-8d24-3b9a0e4c7f52")
	dogma.RegisterDeadline[*ScopeCapturingDeadline]("e84b2d17-9c06-4a3f-b158-6f2d1a7c9e03")
}
