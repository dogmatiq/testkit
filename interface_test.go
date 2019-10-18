package testkit_test

// mockT is a mock of the T interface.
type mockT struct {
}

func (t *mockT) Log(args ...interface{}) {

}
func (t *mockT) Logf(f string, args ...interface{}) {

}

func (t *mockT) FailNow() {
	panic("test failed")
}
