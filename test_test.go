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
				c.RegisterAggregate(&fixtures.AggregateMessageHandler{
					RouteCommandToInstanceFunc: func(m dogma.Message) string {
						return "<instance>"
					},
					ConfigureFunc: func(c dogma.AggregateConfigurer) {
						c.Identity("<aggregate>", "<aggregate-key>")
						c.ConsumesCommandType(fixtures.MessageC{})
						c.ProducesEventType(fixtures.MessageE{})
					},
				})
				c.RegisterProjection(&fixtures.ProjectionMessageHandler{
					ConfigureFunc: func(c dogma.ProjectionConfigurer) {
						c.Identity("<projection>", "<projection-key>")
						c.ConsumesEventType(fixtures.MessageE{})
					},
				})
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
			It("logs a heading", func() {
				test.Prepare()
				Expect(t.Logs).To(ContainElement(
					"--- PREPARING APPLICATION FOR TEST ---",
				))
			})
		})

		Describe("func ExecuteCommand()", func() {
			It("logs file and line information in headings", func() {
				test.ExecuteCommand(
					fixtures.MessageC1,
					noopAssertion{},
				)
				Expect(t.Logs).To(ContainElement(
					"--- EXECUTING TEST COMMAND ---",
				))
				Expect(t.Logs).To(ContainElement(
					"--- ASSERTION REPORT ---\n\n✓ pass unconditionally\n\n",
				))
			})
		})

		Describe("func RecordEvent()", func() {
			It("logs file and line information in headings", func() {
				test.RecordEvent(
					fixtures.MessageE1,
					noopAssertion{},
				)
				Expect(t.Logs).To(ContainElement(
					"--- RECORDING TEST EVENT ---",
				))
				Expect(t.Logs).To(ContainElement(
					"--- ASSERTION REPORT ---\n\n✓ pass unconditionally\n\n",
				))
			})
		})

		Describe("func AdvanceTimeBy()", func() {
			It("logs file and line information in headings", func() {
				test.AdvanceTimeBy(
					3*time.Second,
					noopAssertion{},
				)
				Expect(t.Logs).To(ContainElement(
					"--- ADVANCING TIME BY 3s ---",
				))
				Expect(t.Logs).To(ContainElement(
					"--- ASSERTION REPORT ---\n\n✓ pass unconditionally\n\n",
				))
			})
		})

		Describe("func AdvanceTimeTo()", func() {
			It("logs file and line information in headings", func() {
				test.AdvanceTimeTo(
					time.Date(2100, 1, 2, 3, 4, 5, 6, time.UTC),
					noopAssertion{},
				)
				Expect(t.Logs).To(ContainElement(
					"--- ADVANCING TIME TO 2100-01-02T03:04:05Z ---",
				))
				Expect(t.Logs).To(ContainElement(
					"--- ASSERTION REPORT ---\n\n✓ pass unconditionally\n\n",
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
