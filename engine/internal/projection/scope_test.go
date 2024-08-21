package projection_test

import (
	"context"
	"time"

	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogma/fixtures"
	. "github.com/dogmatiq/enginekit/enginetest/stubs"
	. "github.com/dogmatiq/testkit/engine/internal/projection"
	"github.com/dogmatiq/testkit/envelope"
	"github.com/dogmatiq/testkit/fact"
	g "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = g.Describe("type scope", func() {
	var (
		handler *ProjectionMessageHandlerStub
		config  configkit.RichProjection
		ctrl    *Controller
		event   *envelope.Envelope
	)

	g.BeforeEach(func() {
		event = envelope.NewEvent(
			"1000",
			MessageE1,
			time.Now(),
		)

		handler = &ProjectionMessageHandlerStub{
			ConfigureFunc: func(c dogma.ProjectionConfigurer) {
				c.Identity("<name>", "deaaf068-bfd3-4ed2-a69d-850cb9bfab8d")
				c.Routes(
					dogma.HandlesEvent[MessageE](),
				)
			},
		}

		config = configkit.FromProjection(handler)

		ctrl = &Controller{
			Config: config,
		}
	})

	g.Describe("func RecordedAt()", func() {
		g.It("returns event creation time", func() {
			handler.HandleEventFunc = func(
				_ context.Context,
				_, _, _ []byte,
				s dogma.ProjectionEventScope,
				_ dogma.Event,
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

	g.Describe("func IsPrimaryDelivery()", func() {
		g.It("returns true", func() {
			handler.HandleEventFunc = func(
				_ context.Context,
				_, _, _ []byte,
				s dogma.ProjectionEventScope,
				_ dogma.Event,
			) (bool, error) {
				Expect(s.IsPrimaryDelivery()).To(BeTrue())
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

	g.Describe("func Log()", func() {
		g.BeforeEach(func() {
			handler.HandleEventFunc = func(
				_ context.Context,
				_, _, _ []byte,
				s dogma.ProjectionEventScope,
				_ dogma.Event,
			) (bool, error) {
				s.Log("<format>", "<arg-1>", "<arg-2>")
				return true, nil
			}
		})

		g.It("records a fact", func() {
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
					LogArguments: []any{
						"<arg-1>",
						"<arg-2>",
					},
				},
			))
		})
	})

	g.Describe("func Now()", func() {
		g.It("returns the current engine time", func() {
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
