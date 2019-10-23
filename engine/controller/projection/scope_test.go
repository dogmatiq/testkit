package projection_test

import (
	"context"
	"time"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/enginekit/fixtures"
	"github.com/dogmatiq/enginekit/identity"
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
		event      = envelope.NewEvent(
			"1000",
			fixtures.MessageA1,
			time.Now(),
		)
	)

	BeforeEach(func() {
		handler = &fixtures.ProjectionMessageHandler{}
		controller = NewController(
			identity.MustNew("<name>", "<key>"),
			handler,
		)
	})

	Describe("func RecordedAt", func() {
		It("returns event creation time", func() {
			handler.HandleEventFunc = func(
				_ context.Context,
				_, _, _ []byte,
				s dogma.ProjectionEventScope,
				_ dogma.Message,
			) (bool, error) {
				Expect(s.RecordedAt()).To(
					BeTemporally("==", event.CreatedAt),
				)
				return true, nil
			}

			_, err := controller.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				event,
			)
			Expect(err).ShouldNot(HaveOccurred())
		})
	})

	Describe("func Log", func() {
		BeforeEach(func() {
			handler.HandleEventFunc = func(
				_ context.Context,
				_, _, _ []byte,
				s dogma.ProjectionEventScope,
				_ dogma.Message,
			) (bool, error) {
				s.Log("<format>", "<arg-1>", "<arg-2>")
				return true, nil
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
