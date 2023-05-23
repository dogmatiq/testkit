package engine_test

import (
	"context"

	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogma/fixtures"
	. "github.com/dogmatiq/testkit/engine"
	g "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = g.Describe("type EventRecorder", func() {
	var (
		process  *ProcessMessageHandler
		app      *Application
		engine   *Engine
		recorder *EventRecorder
	)

	g.BeforeEach(func() {
		process = &ProcessMessageHandler{
			ConfigureFunc: func(c dogma.ProcessConfigurer) {
				c.Identity("<process>", "173b93f6-8359-4605-8c0a-f1076e14993e")
				c.ConsumesEventType(MessageE{})
				c.ProducesCommandType(MessageC{})
			},
		}

		app = &Application{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "6b4206e5-8d36-440b-828f-9fb4623432e2")
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

	g.Describe("func RecordEvent()", func() {
		g.It("dispatches to the engine", func() {
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

		g.It("panics if the message is not an event", func() {
			Expect(func() {
				recorder.RecordEvent(context.Background(), MessageC1)
			}).To(PanicWith("can not record event, fixtures.MessageC is configured as a command"))
		})

		g.It("panics if the message is unrecognized", func() {
			Expect(func() {
				recorder.RecordEvent(context.Background(), MessageX1)
			}).To(PanicWith("can not record event, fixtures.MessageX is a not a recognized message type"))
		})
	})
})
