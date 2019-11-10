package testkit_test

import (
	"time"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/enginekit/fixtures"
	. "github.com/dogmatiq/testkit"
	"github.com/dogmatiq/testkit/assert"
	"github.com/dogmatiq/testkit/compare"
	"github.com/dogmatiq/testkit/engine/fact"
	"github.com/dogmatiq/testkit/render"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("type Test", func() {
	var app *fixtures.Application

	BeforeEach(func() {
		app = &fixtures.Application{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "<app-key>")
			},
		}
	})

	Context("when verbose logging is enabled", func() {
		var (
			t    *mockT
			test *Test
		)

		BeforeEach(func() {
			t = &mockT{}
			test = New(app).Begin(t, Verbose(true))
		})

		Describe("func Prepare()", func() {
			It("logs file and line information in headings", func() {
				test.Prepare() // <-- THIS LINE NUMBER SHOULD APPEAR IN THE OUTPUT
				Expect(t.Logs).To(ContainElement(
					"--- PREPARING APPLICATION FOR TEST (test_test.go:41) ---",
				))
			})
		})

		Describe("func ExecuteCommand()", func() {
			It("logs file and line information in headings", func() {
				test.ExecuteCommand( // <-- THIS LINE NUMBER SHOULD APPEAR IN THE OUTPUT
					fixtures.MessageC1,
					noopAssertion{},
				)
				Expect(t.Logs).To(ContainElement(
					"--- EXECUTING TEST COMMAND (test_test.go:50) ---",
				))
				Expect(t.Logs).To(ContainElement(
					"--- ASSERTION REPORT (test_test.go:50) ---\n\n✓ pass unconditionally\n\n",
				))
			})
		})

		Describe("func RecordEvent()", func() {
			It("logs file and line information in headings", func() {
				test.RecordEvent( // <-- THIS LINE NUMBER SHOULD APPEAR IN THE OUTPUT
					fixtures.MessageC1,
					noopAssertion{},
				)
				Expect(t.Logs).To(ContainElement(
					"--- RECORDING TEST EVENT (test_test.go:65) ---",
				))
				Expect(t.Logs).To(ContainElement(
					"--- ASSERTION REPORT (test_test.go:65) ---\n\n✓ pass unconditionally\n\n",
				))
			})
		})

		Describe("func AdvanceTimeBy()", func() {
			It("logs file and line information in headings", func() {
				test.AdvanceTimeBy( // <-- THIS LINE NUMBER SHOULD APPEAR IN THE OUTPUT
					3*time.Second,
					noopAssertion{},
				)
				Expect(t.Logs).To(ContainElement(
					"--- ADVANCING TIME BY 3s (test_test.go:80) ---",
				))
				Expect(t.Logs).To(ContainElement(
					"--- ASSERTION REPORT (test_test.go:80) ---\n\n✓ pass unconditionally\n\n",
				))
			})
		})

		Describe("func AdvanceTimeTo()", func() {
			It("logs file and line information in headings", func() {
				test.AdvanceTimeTo( // <-- THIS LINE NUMBER SHOULD APPEAR IN THE OUTPUT
					time.Date(2100, 1, 2, 3, 4, 5, 6, time.UTC),
					noopAssertion{},
				)
				Expect(t.Logs).To(ContainElement(
					"--- ADVANCING TIME TO 2100-01-02T03:04:05Z (test_test.go:95) ---",
				))
				Expect(t.Logs).To(ContainElement(
					"--- ASSERTION REPORT (test_test.go:95) ---\n\n✓ pass unconditionally\n\n",
				))
			})
		})
	})
})

type noopAssertion struct{}

func (noopAssertion) Prepare(compare.Comparator) {

}

func (noopAssertion) Ok() bool {
	return true
}

func (noopAssertion) BuildReport(ok bool, r render.Renderer) *assert.Report {
	return &assert.Report{
		TreeOk:   ok,
		Ok:       ok,
		Criteria: "pass unconditionally",
	}
}

func (noopAssertion) Notify(fact.Fact) {

}
