package report_test

import (
	. "github.com/dogmatiq/testkit/internal/report"
	. "github.com/onsi/ginkgo"
)

var _ = Describe("func RenderText()", func() {
	describeRenderer(
		"text",
		"txt",
		RenderText,
	)
})
