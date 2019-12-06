package projection_test

import (
	"context"
	"errors"
	"time"

	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/testkit/engine/controller"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/enginekit/fixtures"
	. "github.com/dogmatiq/testkit/engine/controller/projection"
	"github.com/dogmatiq/testkit/engine/envelope"
	"github.com/dogmatiq/testkit/engine/fact"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ controller.Controller = &Controller{}

var _ = Describe("type Controller", func() {
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
			configkit.MustNewIdentity("<name>", "<key>"),
			handler,
		)
	})

	Describe("func Identity()", func() {
		It("returns the handler identity", func() {
			Expect(controller.Identity()).To(Equal(
				configkit.MustNewIdentity("<name>", "<key>"),
			))
		})
	})

	Describe("func Type()", func() {
		It("returns configkit.ProjectionHandlerType", func() {
			Expect(controller.Type()).To(Equal(configkit.ProjectionHandlerType))
		})
	})

	Describe("func Tick()", func() {
		It("does not return any envelopes", func() {
			envelopes, err := controller.Tick(
				context.Background(),
				fact.Ignore,
				time.Now(),
			)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(envelopes).To(BeEmpty())
		})

		It("does not record any facts", func() {
			buf := &fact.Buffer{}
			_, err := controller.Tick(
				context.Background(),
				buf,
				time.Now(),
			)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(buf.Facts()).To(BeEmpty())
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
				Expect(m).To(Equal(fixtures.MessageA1))
				return true, nil
			}

			_, err := controller.Handle(
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

			_, err := controller.Handle(
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

			_, err := controller.Handle(
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

			_, err := controller.Handle(
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

			_, err := controller.Handle(
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

			_, err := controller.Handle(
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

			_, err := controller.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				event,
			)

			Expect(err).ShouldNot(HaveOccurred())
		})

		It("uses the handler's timeout hint", func() {
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

			_, err := controller.Handle(
				context.Background(),
				fact.Ignore,
				time.Now(),
				event,
			)

			Expect(err).ShouldNot(HaveOccurred())
		})
	})

	Describe("func Reset()", func() {
		It("does nothing", func() {
			controller.Reset()
		})
	})
})
