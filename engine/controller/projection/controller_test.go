package projection_test

import (
	"context"
	"errors"
	"time"

	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogma/fixtures"
	"github.com/dogmatiq/testkit/engine/controller"
	. "github.com/dogmatiq/testkit/engine/controller/projection"
	"github.com/dogmatiq/testkit/engine/envelope"
	"github.com/dogmatiq/testkit/engine/fact"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ controller.Controller = &Controller{}

var _ = Describe("type Controller", func() {
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

	Describe("func Identity()", func() {
		It("returns the handler identity", func() {
			Expect(ctrl.Identity()).To(Equal(
				configkit.MustNewIdentity("<name>", "<key>"),
			))
		})
	})

	Describe("func Type()", func() {
		It("returns configkit.ProjectionHandlerType", func() {
			Expect(ctrl.Type()).To(Equal(configkit.ProjectionHandlerType))
		})
	})

	Describe("func Tick()", func() {
		It("does not return any envelopes", func() {
			envelopes, err := ctrl.Tick(
				context.Background(),
				fact.Ignore,
				time.Now(),
			)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(envelopes).To(BeEmpty())
		})

		It("performs projection compaction", func() {
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
						HandlerName: "<name>",
					},
					fact.ProjectionCompactionCompleted{
						HandlerName: "<name>",
						Error:       expect,
					},
				},
			))
		})

		It("does not compact again until CompactDuration has elapsed", func() {
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

	Describe("func Handle()", func() {
		It("forwards the message to the handler", func() {
			called := false
			handler.HandleEventFunc = func(
				_ context.Context,
				_, _, _ []byte,
				_ dogma.ProjectionEventScope,
				m dogma.Message,
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

		It("propagates handler errors", func() {
			expected := errors.New("<error>")

			handler.HandleEventFunc = func(
				_ context.Context,
				_, _, _ []byte,
				_ dogma.ProjectionEventScope,
				_ dogma.Message,
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

		It("propagates errors when loading the resource version", func() {
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

		It("passes the correct OCC values", func() {
			handler.HandleEventFunc = func(
				ctx context.Context,
				r, c, n []byte,
				_ dogma.ProjectionEventScope,
				_ dogma.Message,
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

		It("does not handle events that have already been applied", func() {
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
				_ dogma.Message,
			) (bool, error) {
				Fail("unexpected call")
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

		It("closes the resource if the event is applied", func() {
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

		It("does not close the resource if the event is not applied", func() {
			handler.HandleEventFunc = func(
				ctx context.Context,
				_, _, _ []byte,
				_ dogma.ProjectionEventScope,
				_ dogma.Message,
			) (bool, error) {
				return false, nil
			}

			handler.CloseResourceFunc = func(
				_ context.Context,
				r []byte,
			) error {
				Fail("unexpected call")
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

		XIt("uses the handler's timeout hint", func() {
			hint := 3 * time.Second
			handler.TimeoutHintFunc = func(dogma.Message) time.Duration {
				return hint
			}

			handler.HandleEventFunc = func(
				ctx context.Context,
				_, _, _ []byte,
				_ dogma.ProjectionEventScope,
				_ dogma.Message,
			) (bool, error) {
				dl, ok := ctx.Deadline()
				Expect(ok).To(BeTrue())
				Expect(dl).To(BeTemporally("~", time.Now().Add(hint)))
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

		It("provides more context to UnexpectedMessage panics from HandleEvent()", func() {
			handler.HandleEventFunc = func(
				_ context.Context,
				_, _, _ []byte,
				_ dogma.ProjectionEventScope,
				_ dogma.Message,
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

		It("provides more context to UnexpectedMessage panics from TimeoutHint()", func() {
			handler.TimeoutHintFunc = func(
				dogma.Message,
			) time.Duration {
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
						"Method":    Equal("TimeoutHint"),
						"Message":   Equal(event.Message),
					},
				),
			))
		})
	})

	Describe("func Reset()", func() {
		It("does nothing", func() {
			ctrl.Reset()
		})
	})
})
