package projection_test

import (
	"context"
	"errors"
	"time"

	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogma/fixtures"
	. "github.com/dogmatiq/testkit/engine/internal/projection"
	"github.com/dogmatiq/testkit/envelope"
	"github.com/dogmatiq/testkit/fact"
	g "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = g.Describe("type Controller", func() {
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

	g.BeforeEach(func() {
		handler = &ProjectionMessageHandler{
			ConfigureFunc: func(c dogma.ProjectionConfigurer) {
				c.Identity("<name>", "fcbe8fe1-1085-497d-ba8e-09bedb031db2")
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

	g.Describe("func HandlerConfig()", func() {
		g.It("returns the handler config", func() {
			Expect(ctrl.HandlerConfig()).To(BeIdenticalTo(config))
		})
	})

	g.Describe("func Tick()", func() {
		g.It("does not return any envelopes", func() {
			envelopes, err := ctrl.Tick(
				context.Background(),
				fact.Ignore,
				time.Now(),
			)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(envelopes).To(BeEmpty())
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
			Expect(err).To(Equal(expect))
			Expect(buf.Facts()).To(Equal(
				[]fact.Fact{
					fact.ProjectionCompactionBegun{
						Handler: config,
					},
					fact.ProjectionCompactionCompleted{
						Handler: config,
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
			Expect(err).To(MatchError("<called>"))

			_, err = ctrl.Tick(
				context.Background(),
				fact.Ignore,
				start.Add(CompactInterval-1), // should not trigger compaction
			)
			Expect(err).ShouldNot(HaveOccurred())

			_, err = ctrl.Tick(
				context.Background(),
				fact.Ignore,
				start.Add(CompactInterval), // should trigger compaction
			)
			Expect(err).To(MatchError("<called>"))
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
				Expect(m).To(Equal(MessageA1))
				return true, nil
			}

			_, err := ctrl.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				event,
			)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(called).To(BeTrue())
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

			Expect(err).To(Equal(expected))
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

			Expect(err).To(Equal(expected))
		})

		g.It("passes the correct OCC values", func() {
			handler.HandleEventFunc = func(
				ctx context.Context,
				r, c, n []byte,
				_ dogma.ProjectionEventScope,
				_ dogma.Event,
			) (bool, error) {
				Expect(r).To(Equal([]byte(event.MessageID)))
				Expect(c).To(BeEmpty())
				Expect(n).NotTo(BeEmpty())
				return false, nil
			}

			_, err := ctrl.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				event,
			)

			Expect(err).ShouldNot(HaveOccurred())
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

			Expect(err).ShouldNot(HaveOccurred())
		})

		g.It("closes the resource if the event is applied", func() {
			called := false
			handler.CloseResourceFunc = func(
				_ context.Context,
				r []byte,
			) error {
				called = true
				Expect(r).To(Equal([]byte(event.MessageID)))
				return nil
			}

			_, err := ctrl.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				event,
			)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(called).To(BeTrue())
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

			Expect(err).ShouldNot(HaveOccurred())
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

			Expect(func() {
				ctrl.Handle(
					context.Background(),
					fact.Ignore,
					time.Now(),
					event,
				)
			}).To(PanicWith(
				MatchFields(
					IgnoreExtras,
					Fields{
						"Handler":   Equal(config),
						"Interface": Equal("ProjectionMessageHandler"),
						"Method":    Equal("HandleEvent"),
						"Message":   Equal(event.Message),
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
				Expect(err).ShouldNot(HaveOccurred())

				Expect(buf.Facts()).NotTo(ContainElement(
					BeAssignableToTypeOf(fact.ProjectionCompactionBegun{}),
				))
				Expect(buf.Facts()).NotTo(ContainElement(
					BeAssignableToTypeOf(fact.ProjectionCompactionCompleted{}),
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
				Expect(err).To(Equal(expect))
				Expect(buf.Facts()).To(Equal(
					[]fact.Fact{
						fact.ProjectionCompactionBegun{
							Handler: config,
						},
						fact.ProjectionCompactionCompleted{
							Handler: config,
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
