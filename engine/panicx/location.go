package panicx

import (
	"fmt"
	"path"
	"runtime"
	"strings"
)

// Location describes a location within the codebase.
type Location struct {
	Func string
	File string
	Line int
}

func (l Location) String() string {
	if l.Func != "" && l.File != "" {
		return fmt.Sprintf(
			"%s() %s:%d",
			l.Func,
			l.File, //path.Base(l.File),
			l.Line,
		)
	}

	if l.Func != "" {
		return fmt.Sprintf("%s()", l.Func)
	}

	if l.File != "" {
		return fmt.Sprintf(
			"%s:%d",
			path.Base(l.File),
			l.Line,
		)
	}

	return "<unknown>"
}

// LocationOfPanic returns the location of the call to panic() that caused the
// stack to start unwinding.
//
// It must be called within a deferred function and only if recover() returned a
// non-nil value. Otherwise the behavior of the function is undefined.
func LocationOfPanic() Location {
	// During a panic() the runtime *adds* frames for each deferred function, so
	// the function that caused the panic is still on the stack, even though it
	// is unwinding.
	//
	// See https://github.com/golang/go/issues/26275
	// See https://github.com/golang/go/issues/26320

	var loc Location
	foundPanicCall := false

	eachFrame(
		func(fr runtime.Frame) bool {
			if fr.Function == "runtime.gopanic" {
				// We found the call to runtime.gopanic(), which is the internal
				// implementation of panic().
				//
				// That means that the next function we find that's NOT in the
				// "runtime" package is the function that called panic().
				foundPanicCall = true
				return true
			}

			if strings.HasPrefix(fr.Function, "runtime.") {
				// We found some other function within the runtime package, we
				// keep looking for some user-land code.
				return true
			}

			// We found some user-land code. If we haven't found the internal
			// call to runtime.gopanic() that means we're still iterating
			// through the frames from inside the defer() so we keep searching.
			if !foundPanicCall {
				return true
			}

			// Otherwise we've found the function that called panic().
			loc = Location{
				Func: fr.Function,
				File: fr.File,
				Line: fr.Line,
			}

			return false
		},
	)

	return loc
}

// eachFrame calls fn for each frame on the call stack until fn returns false.
func eachFrame(fn func(fr runtime.Frame) bool) {
	var pointers [8]uintptr
	var skip int

	for {
		count := runtime.Callers(skip, pointers[:])
		iter := runtime.CallersFrames(pointers[:count])
		skip += count

		for {
			fr, _ := iter.Next()

			if fr.PC == 0 || !fn(fr) {
				return
			}
		}
	}
}
