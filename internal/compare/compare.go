package compare

import (
	"reflect"

	"github.com/dogmatiq/testkit/internal/compare/internal/unsafereflect"
	"github.com/dogmatiq/testkit/location"
	"google.golang.org/protobuf/proto"
)

// Equal returns true if a and b are considered equal.
//
// If both a and b implement [proto.Message], they are compared using
// [proto.Equal].
//
// Otherwise, they are compared using semantics equivalent to
// [reflect.DeepEqual], except that function values are compared by their
// definition site rather than by pointer identity.
//
// This non-standard behavior for functions is necessary to support comparison
// of handler state that contains function values, which are commonly used for
// stubbing in tests. [Equal] is used throughout testkit for comparison of
// [dogma.Message], [dogma.AggregateRoot] and [dogma.ProcessRoot] state; none of
// which are expected to contain function value in production implementations.
func Equal(a, b any) bool {
	if pa, ok := a.(proto.Message); ok {
		if pb, ok := b.(proto.Message); ok {
			return proto.Equal(pa, pb)
		}
	}

	return deepEqual(
		reflect.ValueOf(a),
		reflect.ValueOf(b),
	)
}

func deepEqual(a, b reflect.Value) bool {
	if !a.IsValid() || !b.IsValid() {
		return a.IsValid() == b.IsValid()
	}

	if a.Type() != b.Type() {
		return false
	}

	switch a.Kind() {
	case reflect.Func, reflect.Pointer, reflect.Interface, reflect.Slice, reflect.Map:
		if a.IsNil() != b.IsNil() {
			return false
		}

		if a.IsNil() {
			return true
		}
	}

	switch a.Kind() {
	case reflect.Func:
		return funcEqual(a, b)
	case reflect.Pointer, reflect.Interface:
		return deepEqual(a.Elem(), b.Elem())
	case reflect.Array, reflect.Slice:
		return sliceEqual(a, b)
	case reflect.Map:
		return mapEqual(a, b)
	case reflect.Struct:
		return structEqual(a, b)
	default:
		return reflect.DeepEqual(a.Interface(), b.Interface())
	}
}

func funcEqual(a, b reflect.Value) bool {
	if a.Pointer() == b.Pointer() {
		return true
	}

	la := location.OfFunc(a.Interface())
	lb := location.OfFunc(b.Interface())

	return la.File == lb.File && la.Line == lb.Line
}

func sliceEqual(a, b reflect.Value) bool {
	if a.Len() != b.Len() {
		return false
	}

	for i := range a.Len() {
		if !deepEqual(a.Index(i), b.Index(i)) {
			return false
		}
	}

	return true
}

func mapEqual(a, b reflect.Value) bool {
	if a.Len() != b.Len() {
		return false
	}

	for _, k := range a.MapKeys() {
		va := a.MapIndex(k)
		vb := b.MapIndex(k)

		if !vb.IsValid() {
			return false
		}

		if !deepEqual(va, vb) {
			return false
		}
	}

	return true
}

func structEqual(a, b reflect.Value) bool {
	for i := range a.NumField() {
		fa := a.Field(i)
		fb := b.Field(i)

		if !fa.CanInterface() {
			fa = unsafereflect.MakeMutable(fa)
			fb = unsafereflect.MakeMutable(fb)
		}

		if !deepEqual(fa, fb) {
			return false
		}
	}

	return true
}
