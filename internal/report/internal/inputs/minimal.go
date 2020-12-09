package inputs

import "github.com/dogmatiq/testkit/internal/report"

func init() {
	register(
		"minimal",
		"a report with the absolute minimum amount of information",
		func() report.Report {
			return report.
				New("<%s>", "caption").
				Done()
		},
	)
}
