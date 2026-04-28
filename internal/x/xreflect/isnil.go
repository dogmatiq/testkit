package xreflect

import "reflect"

// IsNil returns true if v is nil, either at the interface level or because it
// contains a nil pointer, channel, func, map, or slice.
func IsNil(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Pointer,
		reflect.Chan,
		reflect.Func,
		reflect.Interface,
		reflect.Map,
		reflect.Slice:
		return rv.IsNil()
	default:
		return false
	}
}
