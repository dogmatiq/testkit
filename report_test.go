package testkit_test

import (
	"github.com/dogmatiq/testkit/internal/testingmock"
	"github.com/dogmatiq/testkit/internal/x/xtesting"
)

// reportMatcher validates that some action produced a particular test report.
type reportMatcher func(*testingmock.T)

// expectReport is a helper function for testing that a testkit.Test produces the
// correct test report.
func expectReport(expected ...string) reportMatcher {
	// Always expect blank lines surrounding the report content.
	expected = append([]string{""}, expected...)
	expected = append(expected, "")

	return func(t *testingmock.T) {
		// Scan through the logs until we find the start of the test report,
		// then verify that the remainder of the log content matches our
		// expectation.
		for i, l := range t.Logs {
			if l == "--- TEST REPORT ---" {
				remainder := t.Logs[i+1:]
				xtesting.Expect(
					t,
					"unexpected report remainder",
					remainder,
					expected,
				)
				return
			}
		}

		// If we didn't find the test report at all just compare all of the logs
		// to the expectation so at least we know what *was* printed.
		xtesting.Expect(
			t,
			"test report not found; unexpected logs",
			t.Logs,
			expected,
		)
	}
}
