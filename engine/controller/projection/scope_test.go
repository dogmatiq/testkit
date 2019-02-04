package projection_test

import (
	"context"
	"time"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/enginekit/fixtures"
	"github.com/dogmatiq/enginekit/message"
	. "github.com/dogmatiq/testkit/engine/controller/projection"
	"github.com/dogmatiq/testkit/engine/envelope"
	"github.com/dogmatiq/testkit/engine/fact"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("type scope", func() {
	var (
		handler    *fixtures.ProjectionMessageHandler
		controller *Controller
		event      = envelope.New(
			"1000",
			fixtures.MessageA1,
			message.EventRole,
		)
	)

	BeforeEach(func() {
		handler = &fixtures.ProjectionMessageHandler{}
		controller = NewController("<name>", handler)
	})

	Describe("func Log", func() {
		BeforeEach(func() {
			handler.HandleEventFunc = func(
				_ context.Context,
				s dogma.ProjectionEventScope,
				_ dogma.Message,
			) error {
				s.Log("<format>", "<arg-1>", "<arg-2>")
				return nil
			}
		})

		It("records a fact", func() {
			buf := &fact.Buffer{}
			_, err := controller.Handle(
				context.Background(),
				buf,
				time.Now(),
				event,
			)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(buf.Facts()).To(ContainElement(
				fact.MessageLoggedByProjection{
					HandlerName: "<name>",
					Handler:     handler,
					Envelope:    event,
					LogFormat:   "<format>",
					LogArguments: []interface{}{
						"<arg-1>",
						"<arg-2>",
					},
				},
			))
		})
	})
})
