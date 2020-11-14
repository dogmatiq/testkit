package testkit_test

import (
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
			test.PrepareX()
			Expect(t.Logs).To(ContainElement(
				"--- PREPARING APPLICATION FOR TEST ---",
			))
		})
	})

	Describe("func ExecuteCommand()", func() {
		It("logs a heading", func() {
			test.ExecuteCommand(
				MessageC1,
				assert.Nothing,
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
				assert.Nothing,
			)
			Expect(t.Logs).To(ContainElement(
				"--- RECORDING TEST EVENT ---",
			))
		})

		It("performs projection compaction", func() {
			test.RecordEvent(
				MessageE1,
				assert.Nothing,
			)
			Expect(t.Logs).To(ContainElement(
				"= ----  ∵ ----  ⋲ ----    Σ    <projection> ● compacted",
			))
		})
	})

	Describe("func Call()", func() {
		It("logs a heading", func() {
			test.Call(
				func() error {
					return nil
				},
				assert.Nothing,
			)

			Expect(t.Logs).To(ContainElement(
				"--- CALLING USER-DEFINED FUNCTION ---",
			))
		})
	})
})
