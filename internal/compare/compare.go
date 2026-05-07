package compare

import (
	"reflect"

	"google.golang.org/protobuf/proto"
)

// Equal returns true if a and b are considered equal.
//
// It supports comparison of protocol buffers messages using the proto.Equal()
// function. All other types are compared using reflect.DeepEqual().
func Equal(a, b any) bool {
	if pa, ok := a.(proto.Message); ok {
		if pb, ok := b.(proto.Message); ok {
			return proto.Equal(pa, pb)
		}
	}

	return reflect.DeepEqual(a, b)
}
