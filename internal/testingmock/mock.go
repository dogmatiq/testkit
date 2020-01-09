package testingmock

import "fmt"

// T is a mock of the testkit.tHelper interface.
type T struct {
	FailSilently bool
	Failed       bool
	Logs         []string
}

// Log is an implementation of testing.TB.Log().
func (t *T) Log(args ...interface{}) {
	t.Logs = append(t.Logs, fmt.Sprintln(args...))
}

// Logf is an implementation of testing.TB.Logf().
func (t *T) Logf(f string, args ...interface{}) {
	t.Logs = append(t.Logs, fmt.Sprintf(f, args...))
}

// FailNow is an implementation of testing.TB.FailNow().
func (t *T) FailNow() {
	t.Failed = true

	if !t.FailSilently {
		panic("test failed")
	}
}

// Helper is an implementation of testing.TB.Helper().
func (t *T) Helper() {
}
