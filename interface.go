package dogmatest

// T is the interface via which the test framework consumes Go's
// *testing.T value.
//
// It allows use of stand-ins, such as Ginkgo's GinkgoT() value.
type T interface {
	Log(args ...interface{})
	Logf(f string, args ...interface{})
	Helper()
	FailNow()
}
