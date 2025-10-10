package fixtures

import "google.golang.org/protobuf/proto"

// MessageDescription panics unconditionally.
func (x *ProtoMessage) MessageDescription() string {
	panic("not implemented")
}

// Validate panics unconditionally.
func (x *ProtoMessage) Validate() error {
	panic("not implemented")
}

// MarshalBinary returns the binary representation of the message.
func (x *ProtoMessage) MarshalBinary() ([]byte, error) {
	return proto.Marshal(x)
}

// UnmarshalBinary unmarshals the binary representation of the message.
func (x *ProtoMessage) UnmarshalBinary(data []byte) error {
	return proto.Unmarshal(data, x)
}
