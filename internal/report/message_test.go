package report_test

import (
	"strings"

	. "github.com/dogmatiq/dogma/fixtures"
	. "github.com/dogmatiq/testkit/internal/report"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("func RenderMessage()", func() {
	It("returns a suitable representation", func() {
		Expect(
			RenderMessage(MessageA1),
		).To(Equal(join(
			"fixtures.MessageA{",
			`    Value: "A1"`,
			"}",
		)))
	})
})

func join(values ...string) string {
	return strings.Join(values, "\n")
}
