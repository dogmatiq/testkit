package testkit_test

import (
	"testing"

	"github.com/dogmatiq/testkit/internal/testingmock"
	"github.com/dogmatiq/testkit/internal/x/xtesting"
)

// reportMatcher validates that some action produced a particular test report.
//
// realT is the real test, which is failed if the report logged to mockT does
// not match. mockT is typically configured to fail silently, so it can not be
// used to report the mismatch itself.
type reportMatcher func(realT testing.TB, mockT *testingmock.T)

// expectReport is a helper function for testing that a testkit.Test produces the
// correct test report.
func expectReport(expected ...string) reportMatcher {
	// Always expect blank lines surrounding the report content.
	expected = append([]string{""}, expected...)
	expected = append(expected, "")

	return func(realT testing.TB, mockT *testingmock.T) {
		realT.Helper()

		// Scan through the logs until we find the start of the test report,
		// then verify that the remainder of the log content matches our
		// expectation.
		for i, l := range mockT.Logs {
			if l == "--- TEST REPORT ---" {
				remainder := mockT.Logs[i+1:]
				xtesting.Expect(
					realT,
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
			realT,
			"test report not found; unexpected logs",
			mockT.Logs,
			expected,
		)
	}
}
