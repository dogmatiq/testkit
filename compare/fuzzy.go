package compare

import (
	"math"
	"reflect"
)

// TypeSimilarity measures the similarity between two types. The higher the
// number, the more similar the types are.
type TypeSimilarity uint64

const (
	// SameTypes is a TypeSimilarity value indicating that two types are
	// identical.
	SameTypes TypeSimilarity = math.MaxUint64

	// UnrelatedTypes is a TypeSimilarity value that indicates that two types
	// are totally unrelated.
	UnrelatedTypes TypeSimilarity = 0
)

// FuzzyTypeComparison returns the "similarity" between two types.
func FuzzyTypeComparison(a, b reflect.Type) TypeSimilarity {
	v := SameTypes

	if a == b {
		return v
	}

	if n, ok := pointerDistance(a, b); ok {
		return v - n
	}

	if n, ok := pointerDistance(b, a); ok {
		return v - n
	}

	return UnrelatedTypes
}

// pointerDistance returns the "distance" from the pointer type p, to the
// elemental type t.
func pointerDistance(p, t reflect.Type) (n TypeSimilarity, ok bool) {
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
