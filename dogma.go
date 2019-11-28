package testkit

func log(t T, args ...interface{}) { t.Log(args...) }

// ABOUT THIS FILE (dogma.go)
//
// Go's built-in test framework includes the filename and line number of any
// calls that are made to the testing.T.Log() function and its variants.
//
// In an effort to make it clearer that the log output produced by this package
// originates within Dogma's tooling we perform all engine logging via the
// functions above.
//
// Note also, that all of the logging calls within this file are made on a
// single-digit line-number, ensuring that all output is aligned and as short as
// possible.
//
// Go *does* include the testing.T.Helper() function, which marks the
// calling-function as a "helper" so that it does not get considered as the
// source of logs. Unfortunately, this only marks a single stack frame as a
// "helper", and so it does not work for log messages generated deeply within
// the engine.
