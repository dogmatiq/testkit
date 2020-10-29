package engine_test

import (
	"context"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogma/fixtures"
	. "github.com/dogmatiq/testkit/engine"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("type EventRecorder", func() {
	var (
		projection *ProjectionMessageHandler
		app        *Application
		engine     *Engine
		recorder   *EventRecorder
	)

	BeforeEach(func() {
		projection = &ProjectionMessageHandler{
			ConfigureFunc: func(c dogma.ProjectionConfigurer) {
				c.Identity("<projection>", "<projection-key>")
				c.ConsumesEventType(MessageE{})
			},
		}

		app = &Application{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "<app-key>")
				c.RegisterProjection(projection)
			},
		}

		var err error
		engine, err = New(app)
		Expect(err).ShouldNot(HaveOccurred())

		recorder = &EventRecorder{
			Engine: engine,
		}
	})

	Describe("func RecordEvent()", func() {
		It("dispatches to the engine", func() {
			called := false
			projection.HandleEventFunc = func(
				_ context.Context,
				_, _, _ []byte,
				_ dogma.ProjectionEventScope,
				m dogma.Message,
			) (bool, error) {
				called = true
				Expect(m).To(Equal(MessageE1))
				return true, nil
			}

			err := recorder.RecordEvent(context.Background(), MessageE1)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(called).To(BeTrue())
		})
	})
})
