package projection_test

import (
	"context"
	"errors"

	"github.com/dogmatiq/dogmatest/engine/envelope"
	"github.com/dogmatiq/dogmatest/internal/enginekit/message"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogmatest/engine/controller/projection"
	"github.com/dogmatiq/dogmatest/engine/fact"
	handlerkit "github.com/dogmatiq/dogmatest/internal/enginekit/handler"
	"github.com/dogmatiq/dogmatest/internal/fixtures"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("type Controller", func() {
	var (
		handler    *fixtures.ProjectionMessageHandler
		controller *Controller
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

	Describe("func Handle()", func() {
		It("forwards the message to the handler", func() {
			called := false
			handler.HandleEventFunc = func(
				_ context.Context,
				_ dogma.ProjectionEventScope,
				hm dogma.Message,
			) error {
				called = true
				Expect(hm).To(Equal(fixtures.MessageA1))
				return nil
			}

			_, err := controller.Handle(
				context.Background(),
				fact.Ignore,
				envelope.New(
					fixtures.MessageA1,
					message.CommandRole,
				),
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

			_, err := controller.Handle(
				context.Background(),
				fact.Ignore,
				envelope.New(fixtures.MessageA{
					Value: "<value>",
				}, message.CommandRole),
			)

			Expect(err).To(Equal(expected))
		})

		It("records a fact when the handler logs a message", func() {
			handler.HandleEventFunc = func(
				_ context.Context,
				s dogma.ProjectionEventScope,
				_ dogma.Message,
			) error {
				s.Log("<format>", "<arg-1>", "<arg-2>")
				return nil
			}

			buf := &fact.Buffer{}
			env := envelope.New(
				fixtures.MessageA1,
				message.CommandRole,
			)

			_, err := controller.Handle(
				context.Background(),
				buf,
				env,
			)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(buf.Facts).To(Equal(
				[]fact.Fact{
					fact.MessageLoggedByProjection{
						HandlerName: "<name>",
						Envelope:    env,
						LogFormat:   "<format>",
						LogArguments: []interface{}{
							"<arg-1>",
							"<arg-2>",
						},
					},
				},
			))
		})
	})
})
