package assert_test

import (
	. "github.com/dogmatiq/testkit/assert"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

var _ OptionalAssertion = (Assertion)(nil) // ensure OptionalAssertion is always satisfied by Assertion

var _ = Describe("var Nothing", func() {
	It("does not satisfy the Assertion interface", func() {
		gomega.Expect(func() {
			var _ = Nothing.(Assertion)
		}).To(gomega.Panic())
	})
})
