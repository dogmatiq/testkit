package testkit

// T is the interface via which the test framework consumes Go's *testing.T
// value.
//
// It allows use of stand-ins, such as Ginkgo's GinkgoT() value.
type T interface {
	Log(args ...interface{})
	Logf(f string, args ...interface{})
	FailNow()
}

// tHelper is an interface satisfied by T implementations that can mark
// functions as test helpers.
//
// This is separated from T itself largely because Ginkgo's GinkgoTInterface
// does not yet include this method.
//
// See https://github.com/dogmatiq/testkit/issues/61
// See https://github.com/onsi/ginkgo/pull/585
type tHelper interface {
	T

	Helper()
}
