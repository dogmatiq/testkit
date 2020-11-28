package inputs

import "github.com/dogmatiq/testkit/internal/report"

// Report contains information about a report registered to be tested.
type Report struct {
	Key         string
	Description string
	Report      report.Report
}

// Reports is the set of registered reports.
var Reports = map[string]Report{}

// register adds a report to be tested.
func register(
	k, d string,
	fn func() report.Report,
) {
	if _, ok := Reports[k]; ok {
		panic("report already registered")
	}

	Reports[k] = Report{k, d, fn()}
}
