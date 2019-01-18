package config

import "fmt"

// catch calls fn(), and recovers from any panic where the value is a
// ConfigurationError by returning that error.
func catch(fn func()) (err error) {
	defer func() {
		switch r := recover().(type) {
		case ConfigurationError:
			err = r
		case nil:
			return
		default:
			panic(r)
		}
	}()

	fn()

	return
}

// panicf panics with a new ConfigurationError.
func panicf(f string, v ...interface{}) {
	panic(errorf(f, v...))
}

// errorf returns a new ConfigurationError.
func errorf(f string, v ...interface{}) ConfigurationError {
	return ConfigurationError(
		fmt.Sprintf(f, v...),
	)
}

// ConfigurationError is an error representing a fault in an application's
// configuration.
type ConfigurationError string

func (e ConfigurationError) Error() string {
	return string(e)
}
