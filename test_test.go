package testkit_test

import (
	"time"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogma/fixtures"
	. "github.com/dogmatiq/testkit"
	"github.com/dogmatiq/testkit/assert"
	"github.com/dogmatiq/testkit/compare"
	"github.com/dogmatiq/testkit/engine/fact"
	"github.com/dogmatiq/testkit/internal/testingmock"
	"github.com/dogmatiq/testkit/render"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("type Test", func() {
	var (
		app  *Application
		t    *testingmock.T
		test *Test
	)

	BeforeEach(func() {
		app = &Application{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "<app-key>")
				c.RegisterAggregate(&AggregateMessageHandler{
					RouteCommandToInstanceFunc: func(m dogma.Message) string {
						return "<instance>"
					},
					ConfigureFunc: func(c dogma.AggregateConfigurer) {
						c.Identity("<aggregate>", "<aggregate-key>")
						c.ConsumesCommandType(MessageC{})
						c.ProducesEventType(MessageE{})
					},
				})
				c.RegisterProjection(&ProjectionMessageHandler{
					ConfigureFunc: func(c dogma.ProjectionConfigurer) {
						c.Identity("<projection>", "<projection-key>")
						c.ConsumesEventType(MessageE{})
					},
				})
			},
		}

		t = &testingmock.T{}
		test = New(app).Begin(t)
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
		It("logs a heading", func() {
			test.ExecuteCommand(
				MessageC1,
				noopAssertion{},
			)
			Expect(t.Logs).To(ContainElement(
				"--- EXECUTING TEST COMMAND ---",
			))
		})
	})

	Describe("func RecordEvent()", func() {
		It("logs a heading", func() {
			test.RecordEvent(
				MessageE1,
				noopAssertion{},
			)
			Expect(t.Logs).To(ContainElement(
				"--- RECORDING TEST EVENT ---",
			))
		})
	})

	Describe("func AdvanceTime()", func() {
		When("passed a By() advancer", func() {
			It("logs a heading", func() {
				test.AdvanceTime(
					ByDuration(3*time.Second),
					noopAssertion{},
				)
				Expect(t.Logs).To(ContainElement(
					"--- ADVANCING TIME BY 3s ---",
				))
			})

			It("can be called without making an assertion", func() {
				test.AdvanceTime(
					ByDuration(3*time.Second),
					assert.Nothing,
				)
				Expect(t.Logs).To(ContainElement(
					"--- ADVANCING TIME BY 3s ---",
				))
			})
		})

		When("passed a ToTime() advancer", func() {
			It("logs a heading", func() {
				test.AdvanceTime(
					ToTime(time.Date(2100, 1, 2, 3, 4, 5, 6, time.UTC)),
					noopAssertion{},
				)
				Expect(t.Logs).To(ContainElement(
					"--- ADVANCING TIME TO 2100-01-02T03:04:05Z ---",
				))
			})

			It("can be called without making an assertion", func() {
				test.AdvanceTime(
					ToTime(time.Date(2100, 1, 2, 3, 4, 5, 6, time.UTC)),
					assert.Nothing,
				)
				Expect(t.Logs).To(ContainElement(
					"--- ADVANCING TIME TO 2100-01-02T03:04:05Z ---",
				))
			})
		})

		It("panics if the advancer produces a time in the past", func() {
			Expect(func() {
				test.AdvanceTime(
					func(time.Time) (time.Time, string) {
						return time.Time{}, ""
					},
					nil,
				)
			}).To(PanicWith("new time must be after the current time"))
		})
	})
})

type noopAssertion struct{}

func (noopAssertion) Begin(compare.Comparator) {}
func (noopAssertion) End()                     {}
func (noopAssertion) TryOk() (bool, bool)      { return true, true }
func (noopAssertion) Ok() bool                 { return true }
func (noopAssertion) Notify(fact.Fact)         {}
func (noopAssertion) BuildReport(ok bool, r render.Renderer) *assert.Report {
	return &assert.Report{
		TreeOk:   ok,
		Ok:       ok,
		Criteria: "pass unconditionally",
	}
}
