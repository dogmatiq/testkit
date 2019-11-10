package testkit_test

import "fmt"

// mockT is a mock of the T interface.
type mockT struct {
	Logs []string
}

func (t *mockT) Log(args ...interface{}) {
	t.Logs = append(t.Logs, fmt.Sprintln(args...))
}

func (t *mockT) Logf(f string, args ...interface{}) {
	t.Logs = append(t.Logs, fmt.Sprintf(f, args...))
}

func (t *mockT) FailNow() {
	panic("test failed")
}
