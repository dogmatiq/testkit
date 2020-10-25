package testingmock

import (
	"fmt"
	"strings"
)

// T is a mock of the testkit.tHelper interface.
type T struct {
	FailSilently bool
	Failed       bool
	Logs         []string
}

// Log is an implementation of testing.TB.Log().
func (t *T) Log(args ...interface{}) {
	lines := strings.Split(fmt.Sprint(args...), "\n")
	t.Logs = append(t.Logs, lines...)
}

// Logf is an implementation of testing.TB.Logf().
func (t *T) Logf(f string, args ...interface{}) {
	lines := strings.Split(fmt.Sprintf(f, args...), "\n")
	t.Logs = append(t.Logs, lines...)
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
