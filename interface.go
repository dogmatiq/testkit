package testkit

// TestingT is the interface via which the test framework consumes Go's
// *testing.T value.
//
// It allows use of stand-ins, such as Ginkgo's GinkgoT() value.
type TestingT interface {
	Failed() bool
	Log(args ...interface{})
	Logf(f string, args ...interface{})
	Fatal(args ...interface{})
	FailNow()
	Helper()
}
