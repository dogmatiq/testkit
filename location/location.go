package location

import (
	"fmt"
	"path"
	"reflect"
	"runtime"
	"strings"
)

// Location describes a location within a code base.
//
// Any of the fields may be zero-valued, indicating that piece of information is
// unknown.
type Location struct {
	// Func is the fully-qualified name of the function.
	//
	// The format of this string is not defined by this package and is subject
	// to change.
	Func string

	// File is the name of the Go source file.
	File string

	// Line is the line number within the file, starting at 1.
	Line int
}

func (l Location) String() string {
	if fl, ok := l.FileLine(); ok {
		// We know the file/line number which is enough to identify the
		// location, so only include the function name if it actually
		// provides more context. That means we DON'T render the function
		// name if it refers to a global closure.
		if l.Func != "" && !strings.Contains(l.Func, "glob..") {
			return fmt.Sprintf(
				"%s [%s(...)]",
				fl,
				l.Func,
			)
		}

		return fl
	}

	if l.Func != "" {
		// We don't know the file/line number, but we do know the function name,
		// so render it even if it refers to a global closure because it's the
		// best information we have.
		return fmt.Sprintf("%s(...)", l.Func)
	}

	return "<unknown>"
}

// FileLine returns the file and line number of the location in <file>:<line>
// format.
//
// ok is false if this information is unknown.
func (l Location) FileLine() (_ string, ok bool) {
	if l.File == "" {
		return "", false
	}

	return fmt.Sprintf(
		"%s:%d",
		path.Base(l.File),
		l.Line,
	), true
}

// OfFunc returns the location of the definition of fn.
func OfFunc(fn any) Location {
	return ofFunc(reflect.ValueOf(fn))
}

func ofFunc(rv reflect.Value) Location {
	if rv.Kind() != reflect.Func {
		panic("fn must be a function")
	}

	var (
		loc Location
		pc  = rv.Pointer()
	)

	if fn := runtime.FuncForPC(pc); fn != nil {
		loc.Func = fn.Name()
		loc.File, loc.Line = fn.FileLine(pc)
	}

	return loc
}

// OfMethod returns the location of the definition of fn.
func OfMethod(recv any, m string) Location {
	rt := reflect.TypeOf(recv)

	rm, ok := rt.MethodByName(m)
	if !ok {
		panic("method does not exist")
	}

	return ofFunc(rm.Func)
}

// OfCall returns the location where its caller was called itself.
func OfCall() Location {
	var loc Location

	eachFrame(
		2, // skip LocationOfCall() and its caller.
		func(fr runtime.Frame) bool {
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

// OfPanic returns the location of the call to panic() that caused the
// stack to start unwinding.
//
// It must be called within a deferred function and only if recover() returned a
// non-nil value. Otherwise the behavior of the function is undefined.
func OfPanic() Location {
	// During a panic() the runtime *adds* frames for each deferred function, so
	// the function that caused the panic is still on the stack, even though it
	// is unwinding.
	//
	// See https://github.com/golang/go/issues/26275
	// See https://github.com/golang/go/issues/26320

	var loc Location
	foundPanicCall := false

	eachFrame(
		0,
		func(fr runtime.Frame) bool {
			if strings.HasPrefix(fr.Function, "runtime.") {
				// We found some function within the runtime package, we keep
				// looking for some user-land code.

				if fr.Function == "runtime.gopanic" {
					// We found the call to runtime.gopanic(), which is the
					// internal implementation of panic().
					//
					// That means that the next function we find that's NOT in
					// the "runtime" package is the function that called
					// panic().
					foundPanicCall = true
				}

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
func eachFrame(skip int, fn func(fr runtime.Frame) bool) {
	var pointers [8]uintptr
	skip += 2 // Always skip runtime.Callers() and eachFrame().

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
