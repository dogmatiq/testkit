package projection_test

import (
	"context"
	"errors"
	"time"

	"github.com/dogmatiq/dogmatest/engine/controller"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogmatest/engine/controller/projection"
	"github.com/dogmatiq/dogmatest/engine/envelope"
	"github.com/dogmatiq/dogmatest/engine/fact"
	"github.com/dogmatiq/dogmatest/internal/enginekit/fixtures"
	handlerkit "github.com/dogmatiq/dogmatest/internal/enginekit/handler"
	"github.com/dogmatiq/dogmatest/internal/enginekit/message"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ controller.Controller = &Controller{}

var _ = Describe("type Controller", func() {
	var (
		handler    *fixtures.ProjectionMessageHandler
		controller *Controller
		event      = envelope.New(
			fixtures.MessageA1,
			message.EventRole,
		)
	)

	BeforeEach(func() {
		handler = &fixtures.ProjectionMessageHandler{}
		controller = NewController("<name>", handler)
	})

	Describe("func Name()", func() {
		It("returns the handler name", func() {
			Expect(controller.Name()).To(Equal("<name>"))
		})
	})

	Describe("func Type()", func() {
		It("returns handler.ProjectionType", func() {
			Expect(controller.Type()).To(Equal(handlerkit.ProjectionType))
		})
	})

	Describe("func Tick()", func() {
		It("does nothing", func() {
			t, err := controller.Tick(context.Background(), time.Now())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(t).To(BeNil())
		})
	})

	Describe("func Handle()", func() {
		It("forwards the message to the handler", func() {
			called := false
			handler.HandleEventFunc = func(
				_ context.Context,
				_ dogma.ProjectionEventScope,
				m dogma.Message,
			) error {
				called = true
				Expect(m).To(Equal(fixtures.MessageA1))
				return nil
			}

			_, _, err := controller.Handle(
				context.Background(),
				fact.Ignore,
				event,
			)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(called).To(BeTrue())
		})

		It("propagates handler errors", func() {
			expected := errors.New("<error>")

			handler.HandleEventFunc = func(
				_ context.Context,
				_ dogma.ProjectionEventScope,
				_ dogma.Message,
			) error {
				return expected
			}

			_, _, err := controller.Handle(
				context.Background(),
				fact.Ignore,
				event,
			)

			Expect(err).To(Equal(expected))
		})

		It("returns a nil next-tick time", func() {
			t, _, err := controller.Handle(
				context.Background(),
				fact.Ignore,
				event,
			)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(t).To(BeNil())
		})

	})

	Describe("func Reset()", func() {
		It("does nothing", func() {
			controller.Reset()
		})
	})
})
