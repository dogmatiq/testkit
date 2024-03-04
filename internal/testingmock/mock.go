package testingmock

import (
	"fmt"
	"strings"
)

// T is a mock of the testkit.tHelper interface.
type T struct {
	FailSilently bool
	Logs         []string

	failed bool
}

// Failed returns true if the test has failed.
func (t *T) Failed() bool {
	return t.failed
}

// Log is an implementation of testing.TB.Log().
func (t *T) Log(args ...any) {
	lines := strings.Split(fmt.Sprint(args...), "\n")
	t.Logs = append(t.Logs, lines...)
}

// Logf is an implementation of testing.TB.Logf().
func (t *T) Logf(f string, args ...any) {
	lines := strings.Split(fmt.Sprintf(f, args...), "\n")
	t.Logs = append(t.Logs, lines...)
}

// Fatal is an implementation of testing.TB.Fatal().
func (t *T) Fatal(args ...any) {
	lines := strings.Split(fmt.Sprint(args...), "\n")
	t.Logs = append(t.Logs, lines...)
	t.failed = true

	if !t.FailSilently {
		panic("test failed: " + lines[0])
	}
}

// FailNow is an implementation of testing.TB.FailNow().
func (t *T) FailNow() {
	t.failed = true

	if !t.FailSilently {
		panic("test failed")
	}
}

// Helper is an implementation of testing.TB.Helper().
func (t *T) Helper() {
}
