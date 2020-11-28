package report_test

import (
	"io"
	"io/ioutil"
	"path"
	"strings"

	. "github.com/dogmatiq/testkit/internal/report"
	"github.com/dogmatiq/testkit/internal/report/internal/inputs"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
)

func describeRenderer(
	dataset, extension string,
	render func(io.Writer, Report) error,
) {
	var entries []TableEntry

	for _, r := range inputs.Reports {
		entries = append(
			entries,
			Entry(
				r.Description,
				r,
			),
		)
	}

	DescribeTable(
		"",
		func(r inputs.Report) {
			var w strings.Builder

			err := render(&w, r.Report)
			Expect(err).ShouldNot(HaveOccurred())

			expect, err := ioutil.ReadFile(
				path.Join("testdata", dataset, r.Key+"."+extension),
			)
			Expect(err).ShouldNot(HaveOccurred())

			before := format.TruncatedDiff
			format.TruncatedDiff = true
			defer func() {
				format.TruncatedDiff = before
			}()

			Expect(w.String()).To(Equal(string(expect)))
		},
		entries...,
	)
}
