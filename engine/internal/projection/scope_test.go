package projection_test

import (
	"context"
	"time"

	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogma/fixtures"
	. "github.com/dogmatiq/testkit/engine/internal/projection"
	"github.com/dogmatiq/testkit/envelope"
	"github.com/dogmatiq/testkit/fact"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("type scope", func() {
	var (
		handler *ProjectionMessageHandler
		config  configkit.RichProjection
		ctrl    *Controller
		event   = envelope.NewEvent(
			"1000",
			MessageA1,
			time.Now(),
		)
	)

	BeforeEach(func() {
		handler = &ProjectionMessageHandler{
			ConfigureFunc: func(c dogma.ProjectionConfigurer) {
				c.Identity("<name>", "<key>")
				c.ConsumesEventType(MessageE{})
			},
		}

		config = configkit.FromProjection(handler)

		ctrl = &Controller{
			Config: config,
		}
	})

	Describe("func RecordedAt()", func() {
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

			_, err := ctrl.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				event,
			)
			Expect(err).ShouldNot(HaveOccurred())
		})
	})

	Describe("func Log()", func() {
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
			_, err := ctrl.Handle(
				context.Background(),
				buf,
				time.Now(),
				event,
			)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(buf.Facts()).To(ContainElement(
				fact.MessageLoggedByProjection{
					Handler:   config,
					Envelope:  event,
					LogFormat: "<format>",
					LogArguments: []interface{}{
						"<arg-1>",
						"<arg-2>",
					},
				},
			))
		})
	})

	Describe("func Now()", func() {
		It("returns the current engine time", func() {
			now := time.Now()

			handler.CompactFunc = func(
				_ context.Context,
				s dogma.ProjectionCompactScope,
			) error {
				Expect(s.Now()).To(Equal(now))
				return nil
			}

			_, err := ctrl.Tick(
				context.Background(),
				fact.Ignore,
				now,
			)

			Expect(err).ShouldNot(HaveOccurred())
		})
	})
})
