package projection_test

import (
	"context"
	"errors"
	"time"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/enginekit/config"
	"github.com/dogmatiq/enginekit/config/runtimeconfig"
	. "github.com/dogmatiq/enginekit/enginetest/stubs"
	. "github.com/dogmatiq/testkit/engine/internal/projection"
	"github.com/dogmatiq/testkit/envelope"
	"github.com/dogmatiq/testkit/fact"
	g "github.com/onsi/ginkgo/v2"
	gm "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = g.Describe("type Controller", func() {
	var (
		handler *ProjectionMessageHandlerStub
		cfg     *config.Projection
		ctrl    *Controller
		event   *envelope.Envelope
	)

	g.BeforeEach(func() {
		event = envelope.NewEvent(
			"1000",
			EventA1,
			time.Now(),
		)

		handler = &ProjectionMessageHandlerStub{
			ConfigureFunc: func(c dogma.ProjectionConfigurer) {
				c.Identity("<name>", "fcbe8fe1-1085-497d-ba8e-09bedb031db2")
				c.Routes(
					dogma.HandlesEvent[EventStub[TypeA]](),
				)
			},
		}

		cfg = runtimeconfig.FromProjection(handler)

		ctrl = &Controller{
			Config: cfg,
		}
	})

	g.Describe("func HandlerConfig()", func() {
		g.It("returns the handler config", func() {
			gm.Expect(ctrl.HandlerConfig()).To(gm.BeIdenticalTo(cfg))
		})
	})

	g.Describe("func Tick()", func() {
		g.It("does not return any envelopes", func() {
			envelopes, err := ctrl.Tick(
				context.Background(),
				fact.Ignore,
				time.Now(),
			)
			gm.Expect(err).ShouldNot(gm.HaveOccurred())
			gm.Expect(envelopes).To(gm.BeEmpty())
		})

		g.It("performs projection compaction", func() {
			expect := errors.New("<error>")

			handler.CompactFunc = func(
				context.Context,
				dogma.ProjectionCompactScope,
			) error {
				return expect
			}

			buf := &fact.Buffer{}
			_, err := ctrl.Tick(
				context.Background(),
				buf,
				time.Now(),
			)
			gm.Expect(err).To(gm.Equal(expect))
			gm.Expect(buf.Facts()).To(gm.Equal(
				[]fact.Fact{
					fact.ProjectionCompactionBegun{
						Handler: cfg,
					},
					fact.ProjectionCompactionCompleted{
						Handler: cfg,
						Error:   expect,
					},
				},
			))
		})

		g.It("does not compact again until CompactDuration has elapsed", func() {
			handler.CompactFunc = func(
				context.Context,
				dogma.ProjectionCompactScope,
			) error {
				return errors.New("<called>")
			}

			start := time.Now()
			_, err := ctrl.Tick(
				context.Background(),
				fact.Ignore,
				start,
			)
			gm.Expect(err).To(gm.MatchError("<called>"))

			_, err = ctrl.Tick(
				context.Background(),
				fact.Ignore,
				start.Add(CompactInterval-1), // should not trigger compaction
			)
			gm.Expect(err).ShouldNot(gm.HaveOccurred())

			_, err = ctrl.Tick(
				context.Background(),
				fact.Ignore,
				start.Add(CompactInterval), // should trigger compaction
			)
			gm.Expect(err).To(gm.MatchError("<called>"))
		})
	})

	g.Describe("func Handle()", func() {
		g.It("forwards the message to the handler", func() {
			called := false
			handler.HandleEventFunc = func(
				_ context.Context,
				_, _, _ []byte,
				_ dogma.ProjectionEventScope,
				m dogma.Event,
			) (bool, error) {
				called = true
				gm.Expect(m).To(gm.Equal(EventA1))
				return true, nil
			}

			_, err := ctrl.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				event,
			)

			gm.Expect(err).ShouldNot(gm.HaveOccurred())
			gm.Expect(called).To(gm.BeTrue())
		})

		g.It("propagates handler errors", func() {
			expected := errors.New("<error>")

			handler.HandleEventFunc = func(
				_ context.Context,
				_, _, _ []byte,
				_ dogma.ProjectionEventScope,
				_ dogma.Event,
			) (bool, error) {
				return false, expected
			}

			_, err := ctrl.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				event,
			)

			gm.Expect(err).To(gm.Equal(expected))
		})

		g.It("propagates errors when loading the resource version", func() {
			expected := errors.New("<error>")

			handler.ResourceVersionFunc = func(
				context.Context,
				[]byte,
			) ([]byte, error) {
				return nil, expected
			}

			_, err := ctrl.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				event,
			)

			gm.Expect(err).To(gm.Equal(expected))
		})

		g.It("passes the correct OCC values", func() {
			handler.HandleEventFunc = func(
				ctx context.Context,
				r, c, n []byte,
				_ dogma.ProjectionEventScope,
				_ dogma.Event,
			) (bool, error) {
				gm.Expect(r).To(gm.Equal([]byte(event.MessageID)))
				gm.Expect(c).To(gm.BeEmpty())
				gm.Expect(n).NotTo(gm.BeEmpty())
				return false, nil
			}

			_, err := ctrl.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				event,
			)

			gm.Expect(err).ShouldNot(gm.HaveOccurred())
		})

		g.It("does not handle events that have already been applied", func() {
			handler.ResourceVersionFunc = func(
				context.Context,
				[]byte,
			) ([]byte, error) {
				return []byte("<not empty>"), nil
			}

			handler.HandleEventFunc = func(
				_ context.Context,
				_, _, _ []byte,
				_ dogma.ProjectionEventScope,
				_ dogma.Event,
			) (bool, error) {
				g.Fail("unexpected call")
				return false, nil
			}

			_, err := ctrl.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				event,
			)

			gm.Expect(err).ShouldNot(gm.HaveOccurred())
		})

		g.It("closes the resource if the event is applied", func() {
			called := false
			handler.CloseResourceFunc = func(
				_ context.Context,
				r []byte,
			) error {
				called = true
				gm.Expect(r).To(gm.Equal([]byte(event.MessageID)))
				return nil
			}

			_, err := ctrl.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				event,
			)

			gm.Expect(err).ShouldNot(gm.HaveOccurred())
			gm.Expect(called).To(gm.BeTrue())
		})

		g.It("does not close the resource if the event is not applied", func() {
			handler.HandleEventFunc = func(
				ctx context.Context,
				_, _, _ []byte,
				_ dogma.ProjectionEventScope,
				_ dogma.Event,
			) (bool, error) {
				return false, nil
			}

			handler.CloseResourceFunc = func(
				_ context.Context,
				r []byte,
			) error {
				g.Fail("unexpected call")
				return nil
			}

			_, err := ctrl.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				event,
			)

			gm.Expect(err).ShouldNot(gm.HaveOccurred())
		})

		g.It("provides more context to UnexpectedMessage panics from HandleEvent()", func() {
			handler.HandleEventFunc = func(
				_ context.Context,
				_, _, _ []byte,
				_ dogma.ProjectionEventScope,
				_ dogma.Event,
			) (bool, error) {
				panic(dogma.UnexpectedMessage)
			}

			gm.Expect(func() {
				ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					event,
				)
			}).To(gm.PanicWith(
				MatchFields(
					IgnoreExtras,
					Fields{
						"Handler":   gm.Equal(cfg),
						"Interface": gm.Equal("ProjectionMessageHandler"),
						"Method":    gm.Equal("HandleEvent"),
						"Message":   gm.Equal(event.Message),
					},
				),
			))
		})

		g.When("compact-during-handling is disabled", func() {
			g.It("does not perform compaction", func() {
				handler.CompactFunc = func(
					context.Context,
					dogma.ProjectionCompactScope,
				) error {
					return errors.New("<error>")
				}

				buf := &fact.Buffer{}
				_, err := ctrl.Handle(
					context.Background(),
					buf,
					time.Now(),
					event,
				)
				gm.Expect(err).ShouldNot(gm.HaveOccurred())

				gm.Expect(buf.Facts()).NotTo(gm.ContainElement(
					gm.BeAssignableToTypeOf(fact.ProjectionCompactionBegun{}),
				))
				gm.Expect(buf.Facts()).NotTo(gm.ContainElement(
					gm.BeAssignableToTypeOf(fact.ProjectionCompactionCompleted{}),
				))
			})
		})

		g.When("compact-during-handling is enabled", func() {
			g.BeforeEach(func() {
				ctrl.CompactDuringHandling = true
			})

			g.It("performs projection compaction", func() {
				expect := errors.New("<error>")

				handler.CompactFunc = func(
					context.Context,
					dogma.ProjectionCompactScope,
				) error {
					return expect
				}

				buf := &fact.Buffer{}
				_, err := ctrl.Handle(
					context.Background(),
					buf,
					time.Now(),
					event,
				)
				gm.Expect(err).To(gm.Equal(expect))
				gm.Expect(buf.Facts()).To(gm.Equal(
					[]fact.Fact{
						fact.ProjectionCompactionBegun{
							Handler: cfg,
						},
						fact.ProjectionCompactionCompleted{
							Handler: cfg,
							Error:   expect,
						},
					},
				))
			})
		})
	})

	g.Describe("func Reset()", func() {
		g.It("does nothing", func() {
			ctrl.Reset()
		})
	})
})
