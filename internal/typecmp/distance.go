package typecmp

import (
	"math"
	"reflect"
)

// Distance measures the distance between two types. The higher the number,
// the more dissimilar the types are.
type Distance uint64

const (
	// Identical is a Distance that indicates two types are identical.
	Identical Distance = 0

	// Unrelated is a Distance that indicate two types are totally unrelated.
	Unrelated Distance = math.MaxUint64
)

// MeasureDistance returns the "distance" between two types.
func MeasureDistance(a, b reflect.Type) Distance {
	if a == b {
		return Identical
	}

	if d, ok := pointerDistance(a, b); ok {
		return d
	}

	if d, ok := pointerDistance(b, a); ok {
		return d
	}

	return Unrelated
}

// pointerDistance returns the "distance" from the pointer type p, to the
// elemental type t.
func pointerDistance(p, t reflect.Type) (n Distance, ok bool) {
	for p.Kind() == reflect.Ptr {
		p = p.Elem()
		n++

		if p == t {
			ok = true
			break
		}
	}

	return
}
