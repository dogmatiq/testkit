package engine_test

import (
	"context"

	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogma/fixtures"
	. "github.com/dogmatiq/testkit/engine"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("type EventRecorder", func() {
	var (
		process  *ProcessMessageHandler
		app      *Application
		engine   *Engine
		recorder *EventRecorder
	)

	BeforeEach(func() {
		process = &ProcessMessageHandler{
			ConfigureFunc: func(c dogma.ProcessConfigurer) {
				c.Identity("<process>", "<process-key>")
				c.ConsumesEventType(MessageE{})
				c.ProducesCommandType(MessageC{})
			},
		}

		app = &Application{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "<app-key>")
				c.RegisterProcess(process)
			},
		}

		engine = MustNew(
			configkit.FromApplication(app),
		)

		recorder = &EventRecorder{
			Engine: engine,
		}
	})

	Describe("func RecordEvent()", func() {
		It("dispatches to the engine", func() {
			called := false
			process.RouteEventToInstanceFunc = func(
				context.Context,
				dogma.Message,
			) (string, bool, error) {
				return "<instance>", true, nil
			}

			process.HandleEventFunc = func(
				_ context.Context,
				_ dogma.ProcessRoot,
				_ dogma.ProcessEventScope,
				m dogma.Message,
			) error {
				called = true
				Expect(m).To(Equal(MessageE1))
				return nil
			}

			err := recorder.RecordEvent(context.Background(), MessageE1)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(called).To(BeTrue())
		})

		It("panics if the message is not an event", func() {
			Expect(func() {
				recorder.RecordEvent(context.Background(), MessageC1)
			}).To(PanicWith("can not record event, fixtures.MessageC is configured as a command"))
		})

		It("panics if the message is unrecognized", func() {
			Expect(func() {
				recorder.RecordEvent(context.Background(), MessageX1)
			}).To(PanicWith("can not record event, fixtures.MessageX is a not a recognized message type"))
		})
	})
})
