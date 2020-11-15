package testkit_test

import (
	"time"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogma/fixtures"
	. "github.com/dogmatiq/testkit"
	"github.com/dogmatiq/testkit/assert"
	"github.com/dogmatiq/testkit/engine"
	"github.com/dogmatiq/testkit/internal/testingmock"
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
		test = New(app).Begin(
			t,
			WithOperationOptions(
				engine.EnableProjections(true),
			),
		)
	})

	Describe("func Prepare()", func() {
		It("logs a heading", func() {
			test.Prepare(
				AdvanceTime(ByDuration(3 * time.Second)),
			)
			Expect(t.Logs).To(ContainElement(
				"--- PREPARE: ADVANCING TIME (by 3s) ---",
			))
		})
	})

	Describe("func Expect()", func() {
		It("logs a heading", func() {
			test.Expect(
				AdvanceTime(ByDuration(3*time.Second)),
				assert.Nothing,
			)
			Expect(t.Logs).To(ContainElement(
				"--- EXPECT: ADVANCING TIME (by 3s) ---",
			))
		})
	})

	Describe("func PrepareX()", func() {
		It("logs a heading", func() {
			test.PrepareX()
			Expect(t.Logs).To(ContainElement(
				"--- PREPARING APPLICATION FOR TEST ---",
			))
		})
	})
})
